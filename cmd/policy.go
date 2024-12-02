/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// policyCmd represents the policy command
var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Find and list policies by name or statement",
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
