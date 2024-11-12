/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// policyCmd represents the policy command
var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Find and list policies by name or statement",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("policy called")
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// policyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// policyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
