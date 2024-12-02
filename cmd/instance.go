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

// instanceCmd represents the instance command
var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Find and list OCI instances",
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

		computeClient, err := core.NewComputeClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")

		if flagList {
			resources.ListInstances(computeClient, compartmentId, vnetClient)
		} else if flagFind != "" {
			resources.FindInstances(computeClient, vnetClient, compartmentId, flagFind, true)
		} else {
			fmt.Println("Invalid flag or flag arguments")
		}
	},
}

func init() {
	rootCmd.AddCommand(instanceCmd)

	instanceCmd.Flags().BoolP("list", "l", false, "List all instances")
	instanceCmd.Flags().StringP("find", "f", "", "Find instance by name pattern search")
}
