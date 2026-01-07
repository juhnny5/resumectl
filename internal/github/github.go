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

package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// DebugMode enables debug output
var DebugMode bool

// Repository represents a GitHub repository
type Repository struct {
	Name            string   `json:"name"`
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	HTMLURL         string   `json:"html_url"`
	StargazersCount int      `json:"stargazers_count"`
	ForksCount      int      `json:"forks_count"`
	Language        string   `json:"language"`
	Topics          []string `json:"topics"`
	Fork            bool     `json:"fork"`
	Archived        bool     `json:"archived"`
}

// GitHubProfile represents a GitHub user profile
type GitHubProfile struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	HTMLURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
	Location  string `json:"location"`
	Email     string `json:"email"`
	Blog      string `json:"blog"`
}

// Project represents a project to add to CV
type Project struct {
	Name         string
	Description  string
	URL          string
	Technologies []string
	Stars        int
}

// FetchTopProjects fetches the top N projects by stars from a GitHub user
func FetchTopProjects(username string, count int) ([]Project, error) {
	if count <= 0 {
		count = 5
	}

	// Fetch user's repositories
	repos, err := fetchUserRepos(username)
	if err != nil {
		return nil, err
	}

	if DebugMode {
		fmt.Printf("[DEBUG] Found %d repositories for user %s\n", len(repos), username)
	}

	// Filter out forks and archived repos, then sort by stars
	var ownRepos []Repository
	for _, repo := range repos {
		if !repo.Fork && !repo.Archived {
			ownRepos = append(ownRepos, repo)
		}
	}

	if DebugMode {
		fmt.Printf("[DEBUG] %d non-fork, non-archived repositories\n", len(ownRepos))
	}

	// Sort by stars descending
	sort.Slice(ownRepos, func(i, j int) bool {
		return ownRepos[i].StargazersCount > ownRepos[j].StargazersCount
	})

	// Take top N
	if len(ownRepos) > count {
		ownRepos = ownRepos[:count]
	}

	// Convert to Project structs
	var projects []Project
	for _, repo := range ownRepos {
		proj := Project{
			Name:        repo.Name,
			Description: repo.Description,
			URL:         repo.HTMLURL,
			Stars:       repo.StargazersCount,
		}

		// Add language as technology if present
		if repo.Language != "" {
			proj.Technologies = append(proj.Technologies, repo.Language)
		}

		// Add topics as technologies
		for _, topic := range repo.Topics {
			// Avoid duplicates
			if !containsIgnoreCase(proj.Technologies, topic) {
				proj.Technologies = append(proj.Technologies, topic)
			}
		}

		if DebugMode {
			fmt.Printf("[DEBUG] Adding project: %s (⭐ %d)\n", repo.Name, repo.StargazersCount)
		}

		projects = append(projects, proj)
	}

	return projects, nil
}

// FetchProfile fetches basic GitHub profile information
func FetchProfile(username string) (*GitHubProfile, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "resumectl/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("GitHub user '%s' not found", username)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var profile GitHubProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// fetchUserRepos fetches all public repositories for a user
func fetchUserRepos(username string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=%d&page=%d&sort=pushed", username, perPage, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "resumectl/1.0")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("GitHub user '%s' not found", username)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var repos []Repository
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, err
		}

		if len(repos) == 0 {
			break
		}

		allRepos = append(allRepos, repos...)

		if len(repos) < perPage {
			break
		}

		page++

		// Safety limit
		if page > 10 {
			break
		}
	}

	return allRepos, nil
}

// containsIgnoreCase checks if a slice contains a string (case insensitive)
func containsIgnoreCase(slice []string, str string) bool {
	strLower := strings.ToLower(str)
	for _, s := range slice {
		if strings.ToLower(s) == strLower {
			return true
		}
	}
	return false
}

// ExtractUsernameFromURL extracts the username from a GitHub URL or returns as-is
func ExtractUsernameFromURL(input string) string {
	input = strings.TrimSpace(input)

	// Remove trailing slash
	input = strings.TrimSuffix(input, "/")

	// Handle full URL
	if strings.Contains(input, "github.com/") {
		parts := strings.Split(input, "github.com/")
		if len(parts) > 1 {
			// Get the username part (before any additional path)
			username := strings.Split(parts[1], "/")[0]
			return username
		}
	}

	// Return as-is if it's just a username
	return input
}
