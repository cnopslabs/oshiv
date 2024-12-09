package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var compartmentCmd = &cobra.Command{
	Use:     "compartment",
	Short:   "Find and list compartments",
	Long:    "Find and list compartments",
	Aliases: []string{"compart"},
	Run: func(cmd *cobra.Command, args []string) {
		ociConfig := utils.SetupOciConfig()

		// Read tenancy ID flag and calculate tenancy
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		utils.SetTenancyConfig(FlagTenancyId, ociConfig)

		// Get tenancy ID and tenancy name from Viper config
		tenancyName := viper.GetString("tenancy-name")
		tenancyId := viper.GetString("tenancy-id")

		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
		utils.CheckError(identityErr)

		compartments := resources.FetchCompartments(tenancyId, identityClient)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")

		flagSetCompartment := cmd.Flags().Lookup("set-compartment")
		flagSetCompartmentString, _ := cmd.Flags().GetString("set-compartment")

		if flagList {
			resources.ListCompartments(compartments, tenancyId, tenancyName)
		} else if flagFind != "" {
			resources.FindCompartments(tenancyId, tenancyName, identityClient, flagFind)
		} else if flagSetCompartment.Changed {
			// Reset config file and wite compartment to file
			utils.WriteCompartmentToFile(flagSetCompartmentString, compartments)
		} else {
			fmt.Println("Invalid sub-command or flag")
		}
	},
}

func init() {
	rootCmd.AddCommand(compartmentCmd)

	compartmentCmd.Flags().BoolP("list", "l", false, "List all compartments")
	compartmentCmd.Flags().StringP("find", "f", "", "Find compartment by name pattern search")
	var flagSetCompartment string
	compartmentCmd.Flags().StringVarP(&flagSetCompartment, "set-compartment", "s", "", "Set compartment name")
}
