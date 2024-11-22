package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display oshiv configuration",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		// Lookup tenancy ID and compartment flags and add to Viper config if passed
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		FlagCompartment := rootCmd.Flags().Lookup("compartment")

		// Initialize configuration
		utils.ConfigInit(FlagTenancyId, FlagCompartment)

		// Print applied configuration
		tenancyName := viper.GetString("tenancy-name")
		fmt.Print("Tenancy name is set to: ")
		utils.Yellow.Println(tenancyName)

		tenancyId := viper.GetString("tenancy-id")
		fmt.Print("Tenancy ID is set to: ")
		utils.Yellow.Println(tenancyId)

		compartment := viper.GetString("compartment")
		fmt.Print("Compartment is set to: ")
		utils.Yellow.Println(compartment)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
