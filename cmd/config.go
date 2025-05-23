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

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display oshiv configuration",
	Long:  "Display oshiv configuration",
	Run: func(cmd *cobra.Command, args []string) {
		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(identityErr)

		region, envVarExists := os.LookupEnv("OCI_CLI_REGION")
		if envVarExists {
			identityClient.SetRegion(region)
		}

		// Read tenancy ID flag and calculate tenancy
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		utils.SetTenancyConfig(FlagTenancyId, utils.OciConfig())
		tenancyId := viper.GetString("tenancy-id")
		tenancyName := viper.GetString("tenancy-name")

		// Add compartment to Viper config if it was passed as flag
		FlagCompartment := rootCmd.Flags().Lookup("compartment")
		compartments := resources.FetchCompartments(tenancyId, identityClient)
		utils.SetCompartmentConfig(FlagCompartment, compartments, tenancyName)
		compartment := viper.GetString("compartment")

		// Print configuration
		fmt.Print("Tenancy name: ")
		utils.Yellow.Println(tenancyName)

		fmt.Print("Tenancy ID: ")
		utils.Yellow.Println(tenancyId)

		fmt.Print("Compartment: ")
		utils.Yellow.Println(compartment)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
