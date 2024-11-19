package cmd

import (
	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/bastion"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Create, list, and connect to bastion sessions",
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

		bastionInfo := resources.FetchBastions(compartmentId, bastionClient)

		flagSetBastion, _ := cmd.Flags().GetString("set-bastion")
		if flagSetBastion != "" {
			// Set bastion in Viper config and write to config file
			resources.SetBastionName(flagSetBastion)
		}

		// Get bastion details from Viper config
		// Viper uses the following order precedence: 1) flag, 2) env var, 3) config file, 4) key/value store, 4) default
		// Attempt to set bastion name from flag
		viper.BindPFlag("bastion-name", cmd.Flags().Lookup("bastion-name"))

		// Attempt to set bastion name from environment variable
		viper.BindEnv("bastion-name", "OCI_BASTION_NAME")

		// Get bastion name from viper config
		bastionName := viper.GetString("bastion-name")

		if bastionName == "" {
			uniqueBastionName := resources.CheckForUniqueBastion(bastionInfo)
			bastionName = uniqueBastionName
		}

		bastionId := bastionInfo[bastionName]

		flagListBastionSessions, _ := cmd.Flags().GetBool("list")
		if flagListBastionSessions {
			resources.ListBastionSessions(bastionClient, bastionId)
		}
	},
}

func init() {
	bastionCmd.AddCommand(sessionCmd)

	// Local flags only exposed to session sub-command
	sessionCmd.Flags().StringP("bastion-name", "b", "", "Bastion name to use for session commands")
	sessionCmd.Flags().StringP("set-bastion", "s", "", "Set bastion name")
	sessionCmd.Flags().BoolP("list", "l", false, "List all bastion sessions")
}
