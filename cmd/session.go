package cmd

import (
	"fmt"
	"os"

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

		bastionNameFromFlag, _ := cmd.Flags().GetString("bastion-name")

		var bastionName string
		if bastionNameFromFlag == "" {
			uniqueBastionName, _ := resources.CheckForUniqueBastion(bastions)

			if uniqueBastionName != "" {
				bastionName = uniqueBastionName
			} else {
				fmt.Print("\nMust specify bastion flag: ")
				utils.Yellow.Println("-b BASTION_NAME")
				os.Exit(1)
			}
		} else {
			bastionName = bastionNameFromFlag
		}

		bastionId := bastions[bastionName]

		flagListBastionSessions, _ := cmd.Flags().GetBool("list")
		flagListActiveBastionSessions, _ := cmd.Flags().GetBool("list-active")

		if flagListBastionSessions {
			resources.ListBastionSessions(bastionClient, bastionId, tenancyName, compartmentName, flagListActiveBastionSessions)
		} else if flagListActiveBastionSessions {
			resources.ListBastionSessions(bastionClient, bastionId, tenancyName, compartmentName, flagListActiveBastionSessions)
		}
	},
}

func init() {
	bastionCmd.AddCommand(sessionCmd)

	// Local flags only exposed to session sub-command
	sessionCmd.Flags().StringP("bastion-name", "b", "", "Bastion name to use for session commands")
	sessionCmd.Flags().BoolP("list", "l", false, "List all bastion sessions")
	sessionCmd.Flags().BoolP("list-active", "a", false, "List all bastion sessions")
}
