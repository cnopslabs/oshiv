package cmd

import (
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display your custom OCI tenancy information",
	Long:  "Display your custom OCI tenancy information",
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintTenancyMap()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
