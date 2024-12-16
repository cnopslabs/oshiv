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
	Short: "List bastion sessions",
	Long:  "List bastion sessions",
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

		bastionClient, err := bastion.NewBastionClientWithConfigurationProvider(utils.OciConfig())
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

		flagListActiveBastionSessions, _ := cmd.Flags().GetBool("list")
		flagListAllBastionSessions, _ := cmd.Flags().GetBool("list-all")

		if flagListAllBastionSessions {
			resources.ListBastionSessions(bastionClient, bastionId, tenancyName, compartment, false)
		} else if flagListActiveBastionSessions {
			resources.ListBastionSessions(bastionClient, bastionId, tenancyName, compartment, flagListActiveBastionSessions)
		}
	},
}

func init() {
	bastionCmd.AddCommand(sessionCmd)

	sessionCmd.Flags().StringP("bastion-name", "b", "", "Bastion name to use for session commands")
	sessionCmd.Flags().BoolP("list-all", "a", false, "List all bastion sessions")
	sessionCmd.Flags().BoolP("list", "l", true, "List active bastion sessions")
}
