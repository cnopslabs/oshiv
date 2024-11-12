/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// instanceCmd represents the instance command
var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Find and list OCI instances",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("instance called")
	},
}

func init() {
	rootCmd.AddCommand(instanceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// instanceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// instanceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
