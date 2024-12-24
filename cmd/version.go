package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// This will be populated dynamically at build time
var version = "unknown"

// Version command for long form usage
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of oshiv CLI",
	Long:  "Print the version number of oshiv CLI",
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

// Helper function for printing the version
func printVersion() {
	fmt.Printf("oshiv version: %s\n", version)
}

func init() {
	// Add the version command for long form (e.g., `oshiv version`)
	rootCmd.AddCommand(versionCmd)

	// Register a global persistent flag to support short form (e.g., `oshiv -v`)
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version number of oshiv CLI")

	// Override the persistent pre-run hook to check for the `-v` flag
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			printVersion()
			// Exit to avoid running another command if -v is passed
			cobra.CheckErr(nil)
		}
	}
}
