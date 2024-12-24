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
	Long:  `Print the version number of oshiv CLI`,
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
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version number of oshiv CLI")
	rootCmd.Flags().BoolP("version", "v", false, "Print the short version number of oshiv CLI")

	cobra.OnInitialize(checkVersionFlag)
}

func checkVersionFlag() {
	if versionFlag, _ := rootCmd.Flags().GetBool("version"); versionFlag {
		printVersion()
	}
}
