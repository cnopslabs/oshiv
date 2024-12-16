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

var instanceCmd = &cobra.Command{
	Use:     "instance",
	Short:   "Find and list OCI instances",
	Long:    "Find and list OCI instances",
	Aliases: []string{"inst"},
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

		computeClient, err := core.NewComputeClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(err)

		vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(err)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")
		flagDisplayImageDetails, _ := cmd.Flags().GetBool("image-details")

		if flagList {
			resources.ListInstances(computeClient, compartmentId, vnetClient, flagDisplayImageDetails, compartment, tenancyName)
		} else if flagFind != "" {
			resources.FindInstances(computeClient, vnetClient, compartmentId, flagFind, flagDisplayImageDetails, compartment, tenancyName)
		} else {
			fmt.Println("Invalid flag or flag arguments")
		}
	},
}

func init() {
	rootCmd.AddCommand(instanceCmd)

	instanceCmd.Flags().BoolP("list", "l", false, "List all instances")
	instanceCmd.Flags().StringP("find", "f", "", "Find instance by name pattern search")
	instanceCmd.Flags().BoolP("image-details", "i", false, "Display image details")
}
