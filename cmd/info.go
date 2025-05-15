package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display your custom OCI tenancy information",
	Long:  "Display your custom OCI tenancy information",
	Run: func(cmd *cobra.Command, args []string) {
		flagTenancyName, _ := cmd.Flags().GetString("lookup-tenancy-id")

		if flagTenancyName != "" {
			TenancyId, err := utils.LookUpTenancyID(flagTenancyName)
			utils.CheckError(err)

			if err == nil {
				fmt.Println(TenancyId)
			}
		} else {
			utils.PrintTenancyMap()
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().StringP("lookup-tenancy-id", "g", "", "Lookup tenancy ID by tenancy name")
}
