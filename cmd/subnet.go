/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// subnetCmd represents the subnet command
var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Find and list subnets",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		// Lookup tenancy ID and compartment flags and add to Viper config if passed
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		FlagCompartment := rootCmd.Flags().Lookup("compartment")
		utils.ConfigInit(FlagTenancyId, FlagCompartment)

		// Get tenancy ID and tenancy name from Viper config
		tenancyName := viper.GetString("tenancy-name")
		tenancyId := viper.GetString("tenancy-id")
		compartmentName := viper.GetString("compartment")

		ociConfig := utils.SetupOciConfig()
		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
		utils.CheckError(identityErr)

		compartments := resources.FetchCompartments(tenancyId, identityClient)
		compartmentId := resources.LookupCompartmentId(compartments, tenancyId, tenancyName, compartmentName)

		vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")

		if flagList {
			resources.ListSubnets(vnetClient, compartmentId)
		} else if flagFind != "" {
			// TODO: implement find
			fmt.Println("Subnet search is not yet enabled, listing all subnets. Use grep!")
			resources.ListSubnets(vnetClient, compartmentId)
		} else {
			fmt.Println("Invalid flag or flag arguments")
		}
	},
}

func init() {
	rootCmd.AddCommand(subnetCmd)

	subnetCmd.Flags().BoolP("list", "l", false, "List all subnets")
	subnetCmd.Flags().StringP("find", "f", "", "Find subnet by name pattern search")
}
