package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/identity"
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

		tenancyId, tenancyName := resources.ValidateTenancyId(identityClient, ociConfig)
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
	},
}

func init() {
	rootCmd.AddCommand(compartmentCmd)

	// Local flags only exposed to compartment command
	compartmentCmd.Flags().BoolP("list", "l", false, "List all compartments")
	compartmentCmd.Flags().StringP("find", "f", "", "Find compartment by name pattern search")
	compartmentCmd.Flags().StringP("set-compartment", "s", "", "Set compartment name")
}
