package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var okeCmd = &cobra.Command{
	Use:   "oke",
	Short: "Find and list OKE clusters",
	Long:  "Find and list OKE clusters",
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

		containerEngineClient, err := containerengine.NewContainerEngineClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(err)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")

		if flagList {
			clusters := resources.FindClusters(containerEngineClient, compartmentId, "")
			resources.PrintClusters(clusters, tenancyName, compartment)
		} else if flagFind != "" {
			clusters := resources.FindClusters(containerEngineClient, compartmentId, flagFind)
			resources.PrintClusters(clusters, tenancyName, compartment)
		} else {
			fmt.Println("Invalid flag or flag arguments")
		}
	},
}

func init() {
	rootCmd.AddCommand(okeCmd)

	okeCmd.Flags().BoolP("list", "l", false, "List all OKE clusters")
	okeCmd.Flags().StringP("find", "f", "", "Find OKE cluster by name pattern search")
}
