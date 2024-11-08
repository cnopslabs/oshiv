/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sessionCmd represents the session command
var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Create, list, and connect to bastion sessions",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("session called")
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sessionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sessionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
