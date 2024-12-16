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

var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Find and list subnets",
	Long:  "Find and list subnets",
	Run: func(cmd *cobra.Command, args []string) {
		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(identityErr)

		// Read tenancy ID flag and calculate tenancy
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		utils.SetTenancyConfig(FlagTenancyId, utils.OciConfig())
		tenancyId := viper.GetString("tenancy-id")
		tenancyName := viper.GetString("tenancy-name")

		// Read compartment flag and add to Viper config
		FlagCompartment := rootCmd.Flags().Lookup("compartment")
		compartments := resources.FetchCompartments(tenancyId, identityClient)
		utils.SetCompartmentConfig(FlagCompartment, compartments, tenancyName)
		compartment := viper.GetString("compartment")

		compartmentId := resources.LookupCompartmentId(compartments, tenancyId, tenancyName, compartment)

		vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(utils.OciConfig())
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
