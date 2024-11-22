package cmd

import (
	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/bastion"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bastionCmd = &cobra.Command{
	Use:   "bastion",
	Short: "Find, list, and connect via the OCI bastion service",
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

		bastionClient, err := bastion.NewBastionClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		bastions := resources.FetchBastions(compartmentId, bastionClient)

		flagList, _ := cmd.Flags().GetBool("list")

		if flagList {
			resources.ListBastions(bastions, tenancyName, compartmentName)
		}
	},
}

func init() {
	rootCmd.AddCommand(bastionCmd)

	// Local flags only exposed to oke command
	bastionCmd.Flags().BoolP("list", "l", false, "List all bastions")
}
