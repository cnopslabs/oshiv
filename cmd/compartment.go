package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/resources"
	"github.com/cnopslabs/oshiv/utils"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/spf13/cobra"
)

var compartmentCmd = &cobra.Command{
	Use:   "compartment",
	Short: "Find and list compartments",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		ociConfig := utils.SetupOciConfig()

		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
		utils.CheckError(identityErr)

		flagTenancyIdOverride, _ := cmd.Flags().GetString("tenancy-id-override")
		tenancyId, tenancyName := resources.ValidateTenancyId(flagTenancyIdOverride, identityClient, ociConfig)

		compartments := resources.FetchCompartments(tenancyId, identityClient)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")
		flagSetCompartment, _ := cmd.Flags().GetString("set-compartment")

		if flagList {
			resources.ListCompartments(compartments, tenancyId, tenancyName)
		} else if flagFind != "" {
			resources.FindCompartments(tenancyId, tenancyName, identityClient, flagFind)
		} else if flagSetCompartment != "" {
			resources.SetCompartmentName(flagSetCompartment)
		} else {
			fmt.Println("Invalid sub-command or flag")
		}

		// flagList, _ := cmd.Flags().GetBool("list")
		// if flagList {
		// 	resources.ListCompartments(compartments, tenancyId, tenancyName)
		// }

		// flagFind, _ := cmd.Flags().GetString("find")
		// if flagFind != "" {
		// 	resources.FindCompartments(tenancyId, tenancyName, identityClient, flagFind)
		// }

		// flagSetCompartment, _ := cmd.Flags().GetString("set-compartment")
		// if flagSetCompartment != "" {
		// 	resources.SetCompartmentName(flagSetCompartment)
		// }
	},
}

func init() {
	rootCmd.AddCommand(compartmentCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compartmentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	compartmentCmd.Flags().BoolP("list", "l", false, "List all compartments")
	compartmentCmd.Flags().StringP("find", "f", "", "Find compartment by name pattern search")
	compartmentCmd.Flags().StringP("set-compartment", "s", "", "Set compartment name")
}
