package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// This will be populated dynamically at build time
var version = "unknown"

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of oshiv cli",
	Long:  `Print the version number of oshiv cli`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("oshiv version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
