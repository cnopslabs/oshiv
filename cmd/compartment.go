package cmd

import (
	"fmt"
	"os"

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
		// Read tenancy ID flag and calculate tenancy
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		utils.SetTenancyConfig(FlagTenancyId, utils.OciConfig())

		// Get tenancy ID and tenancy name from Viper config
		tenancyName := viper.GetString("tenancy-name")
		tenancyId := viper.GetString("tenancy-id")

		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(identityErr)

		region, envVarExists := os.LookupEnv("OCI_CLI_REGION")
		if envVarExists {
			identityClient.SetRegion(region)
		}

		compartments := resources.FetchCompartments(tenancyId, identityClient)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")

		if flagList {
			resources.ListCompartments(compartments, tenancyId, tenancyName)
		} else if flagFind != "" {
			resources.FindCompartments(tenancyId, tenancyName, identityClient, flagFind)
		} else {
			fmt.Println("Invalid sub-command or flag")
		}
	},
}

func init() {
	rootCmd.AddCommand(compartmentCmd)

	compartmentCmd.Flags().BoolP("list", "l", false, "List all compartments")
	compartmentCmd.Flags().StringP("find", "f", "", "Find compartment by name pattern search")
}
