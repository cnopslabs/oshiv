package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/database"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Find and list databases",
	Long:  "Find and list databases",
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

		databaseClient, err := database.NewDatabaseClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(err)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")

		if flagList {
			databases := resources.FindDatabases(databaseClient, compartmentId, "")
			resources.PrintDatabases(databases, tenancyName, compartment)
		} else if flagFind != "" {
			databases := resources.FindDatabases(databaseClient, compartmentId, flagFind)
			resources.PrintDatabases(databases, tenancyName, compartment)
		} else {
			fmt.Println("Invalid flag or flag arguments")
		}
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)

	dbCmd.Flags().BoolP("list", "l", false, "List all databases")
	dbCmd.Flags().StringP("find", "f", "", "Find databases by name pattern search")
}
