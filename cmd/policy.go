package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Find and list policies by name or statement",
	Long:  "Find and list policies by name or statement",
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

		flagList, _ := cmd.Flags().GetBool("list")
		flagFindByName, _ := cmd.Flags().GetString("find-by-name")
		flagFindByStatement, _ := cmd.Flags().GetString("find-by-statement")
		flagIncludeStatement, _ := cmd.Flags().GetBool("include-statements")

		if flagList {
			if flagIncludeStatement {
				resources.ListPolicies(identityClient, compartmentId, false)
			} else {
				resources.ListPolicies(identityClient, compartmentId, true)
			}
		} else if flagFindByName != "" || flagFindByStatement != "" {
			if flagIncludeStatement {
				resources.FindPolicies(identityClient, compartmentId, flagFindByName, flagFindByStatement, false)
			} else {
				resources.FindPolicies(identityClient, compartmentId, flagFindByName, flagFindByStatement, true)
			}
		} else {
			fmt.Println("Invalid flag or flag arguments")
		}
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)

	policyCmd.Flags().BoolP("list", "l", false, "List all policies")
	policyCmd.Flags().StringP("find-by-name", "n", "", "Find policy by name search pattern")
	policyCmd.Flags().StringP("find-by-statement", "s", "", "Find policy by statement search pattern")
	policyCmd.Flags().BoolP("include-statements", "a", false, "Include policy statements in results")
}
