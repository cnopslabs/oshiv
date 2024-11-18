package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
)

var okeCmd = &cobra.Command{
	Use:   "oke",
	Short: "Find and list OKE clusters",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		ociConfig := utils.SetupOciConfig()

		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
		utils.CheckError(identityErr)

		tenancyId, tenancyName := resources.ValidateTenancyId(identityClient, ociConfig)
		compartments := resources.FetchCompartments(tenancyId, identityClient)
		compartmentId, _ := resources.DetermineCompartment(compartments, identityClient, tenancyId, tenancyName)

		containerEngineClient, err := containerengine.NewContainerEngineClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")

		if flagList {
			resources.FindClusters(containerEngineClient, compartmentId, "")
		} else if flagFind != "" {
			resources.FindClusters(containerEngineClient, compartmentId, flagFind)
		} else {
			fmt.Println("Invalid flag or flag arguments")
		}
	},
}

func init() {
	rootCmd.AddCommand(okeCmd)

	// Local flags only exposed to oke command
	okeCmd.Flags().BoolP("list", "l", false, "List all OKE clusters")
	okeCmd.Flags().StringP("find", "f", "", "Find OKE cluster by name pattern search")
}
