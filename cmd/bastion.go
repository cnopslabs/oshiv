/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// bastionCmd represents the bastion command
var bastionCmd = &cobra.Command{
	Use:   "bastion",
	Short: "Find, list, and connect via the OCI bastion service",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bastion called")
	},
}

func init() {
	rootCmd.AddCommand(bastionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bastionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bastionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
