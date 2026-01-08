// Copyright (c) 2026 Julien Briault
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"resumectl/internal/generator"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	port int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a live preview server",
	Long: `Start a local web server to preview your CV with live reload.

The server watches for changes in your YAML file and automatically
regenerates the HTML, refreshing the browser automatically.

Usage examples:
  resumectl serve                    # Start server on port 8080
  resumectl serve --port 3000        # Use a custom port
  resumectl serve -d my_cv.yaml      # Use a custom YAML file
  resumectl serve --theme elegant    # Use a specific theme`,
	Run: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")
}

// liveReloadScript injects JavaScript for auto-refresh
const liveReloadScript = `<script>
(function() {
    let lastModified = '';
    setInterval(function() {
        fetch('/_reload')
            .then(r => r.text())
            .then(t => {
                if (lastModified && lastModified !== t) {
                    location.reload();
                }
                lastModified = t;
            })
            .catch(() => {});
    }, 500);
})();
</script>`

type liveServer struct {
	dataPath    string
	outputDir   string
	theme       string
	color       string
	lastModTime string
	mu          sync.RWMutex
}

func (s *liveServer) regenerate() error {
	gen, err := generator.NewWithColor(s.dataPath, s.theme, s.color, s.outputDir)
	if err != nil {
		return err
	}

	htmlPath := filepath.Join(s.outputDir, "cv.html")
	if err := gen.GenerateHTML(htmlPath); err != nil {
		return err
	}

	s.mu.Lock()
	s.lastModTime = time.Now().Format(time.RFC3339Nano)
	s.mu.Unlock()

	return nil
}

func (s *liveServer) watchFile(ctx context.Context) {
	var lastModTime time.Time

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(s.dataPath)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastModTime) {
				lastModTime = info.ModTime()
				log.Info("File changed, regenerating...", "file", s.dataPath)
				if err := s.regenerate(); err != nil {
					log.Error("Error regenerating", "error", err)
				} else {
					log.Info("CV regenerated successfully")
				}
			}
		}
	}
}

func (s *liveServer) handleReload(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, s.lastModTime)
}

func (s *liveServer) handleCV(w http.ResponseWriter, r *http.Request) {
	htmlPath := filepath.Join(s.outputDir, "cv.html")
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		http.Error(w, "CV not found", http.StatusNotFound)
		return
	}

	// Inject live reload script before </body>
	html := string(content)
	html = injectLiveReload(html)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func injectLiveReload(html string) string {
	// Insert live reload script before </body>
	const bodyClose = "</body>"
	idx := strings.LastIndex(html, bodyClose)
	if idx != -1 {
		return html[:idx] + liveReloadScript + "\n" + html[idx:]
	}
	// Fallback: append at the end
	return html + liveReloadScript
}

func runServe(cmd *cobra.Command, args []string) {
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		log.Fatal("Data file does not exist", "path", dataPath)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("Error creating output directory", "error", err)
	}

	server := &liveServer{
		dataPath:  dataPath,
		outputDir: outputDir,
		theme:     theme,
		color:     primaryColor,
	}

	// Initial generation
	log.Info("Generating initial CV...")
	if err := server.regenerate(); err != nil {
		log.Fatal("Error generating CV", "error", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start file watcher
	go server.watchFile(ctx)

	// Setup HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/_reload", server.handleReload)
	// Serve all requests through custom handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle reload endpoint
		if r.URL.Path == "/_reload" {
			server.handleReload(w, r)
			return
		}
		// Handle root - serve CV
		if r.URL.Path == "/" {
			server.handleCV(w, r)
			return
		}
		// Try to serve static file from output directory (photos, etc.)
		filePath := filepath.Join(outputDir, r.URL.Path)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, filePath)
			return
		}
		// Fallback to CV
		server.handleCV(w, r)
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutting down server...")
		cancel()
		httpServer.Shutdown(context.Background())
	}()

	url := fmt.Sprintf("http://localhost:%d", port)
	log.Info("Starting live preview server", "url", url)
	log.Info("Watching for changes", "file", dataPath)
	log.Info("Press Ctrl+C to stop")
	fmt.Println()
	fmt.Printf("  🌐 Open in browser: %s\n\n", url)

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("Server error", "error", err)
	}
}
