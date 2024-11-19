package cmd

import (
	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/bastion"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
)

var bastionCmd = &cobra.Command{
	Use:   "bastion",
	Short: "Find, list, and connect via the OCI bastion service",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		ociConfig := utils.SetupOciConfig()

		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
		utils.CheckError(identityErr)

		tenancyId, tenancyName := resources.ValidateTenancyId(identityClient, ociConfig)
		compartments := resources.FetchCompartments(tenancyId, identityClient)
		compartmentId, _ := resources.DetermineCompartment(compartments, identityClient, tenancyId, tenancyName)

		bastionClient, err := bastion.NewBastionClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		bastions := resources.FetchBastions(compartmentId, bastionClient)

		flagList, _ := cmd.Flags().GetBool("list")

		if flagList {
			resources.ListBastions(bastions)
		}
	},
}

func init() {
	rootCmd.AddCommand(bastionCmd)

	// Local flags only exposed to oke command
	bastionCmd.Flags().BoolP("list", "l", false, "List all bastions")
}
