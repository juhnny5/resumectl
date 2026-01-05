package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the program version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("resumectl version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
