/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// subnetCmd represents the subnet command
var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Find and list subnets",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("subnet called")
	},
}

func init() {
	rootCmd.AddCommand(subnetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// subnetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// subnetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
