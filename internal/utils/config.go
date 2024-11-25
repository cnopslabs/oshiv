package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Add config to Viper config
func SetConfigString(name string, value string) {
	viper.Set(name, value)
	Logger.Debug("Writing config: " + name + ": " + value)
	viper.WriteConfig()
}

// Initialize configuration
func ConfigInit(FlagTenancyId *pflag.Flag, FlagCompartment *pflag.Flag) {
	ociConfig := SetupOciConfig()

	// Lookup tenancy ID from flags and attempt to add tenancy ID to Viper config. Note: "" will be treated as not set
	// FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
	viper.BindPFlag("tenancy-id", FlagTenancyId)

	// Get tenancy ID from Viper config
	// Viper order precedence: 1) set, 2) flag, 3) env var, 4) config file, 5) key/value store, 6) default
	// Note: (3) env var, (4) config file, and (6) default were attempted in rootCmd
	tenancyId := viper.GetString("tenancy-id")

	// Validate tenancy ID, look up tenancy name and set it in Viper config
	identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
	CheckError(identityErr)
	tenancyName := ValidateTenancyId(identityClient, tenancyId)
	viper.Set("tenancy-name", tenancyName)

	// Do not update config file with tenancy-name and tenancy-id if tenancy-id was passed via flag because
	// flags are treated as runtime overrides
	if !FlagTenancyId.Changed {
		// This means tenancy-id was set via env var, file, or default (oci config)
		// Update config file with tenancy-name and tenancy-id
		viper.WriteConfig()
	}

	// Lookup compartment from flags and attempt to add compartment to Viper config. Note: "" will be treated as nil
	// FlagCompartment := rootCmd.Flags().Lookup("compartment")
	viper.BindPFlag("compartment", FlagCompartment)
}

// Determine tenancy ID, validate it against the OCI API, and get tenancy name
// func ValidateTenancyId(identityClient identity.IdentityClient, ociConfig common.ConfigurationProvider) (string, string) {
func ValidateTenancyId(identityClient identity.IdentityClient, tenancyId string) string {
	// Check for tenancy overrides
	// Viper uses the following order precedence: 1) flag, 2) env var, 3) config file, 4) key/value store, 4) default
	// For tenancy-id, we are currently only supporting 1, 2, and 4
	// If tenancy ID flag was passed, this has already been added to config as flag

	// Get tenancy ID from OCI config and set as the default (lowest precedence order) in viper config
	// ociConfigTenancyId, err := ociConfig.TenancyOCID()
	// utils.CheckError(err)
	// fmt.Println("*** viper.SetDefault tenancy-id ***")
	// viper.SetDefault("tenancy-id", ociConfigTenancyId)

	// Attempt to get tenancy ID from environment variable
	// fmt.Println("*** viper.BindEnv tenancy-id (OCI_CLI_TENANCY) ***")
	// viper.BindEnv("tenancy-id", "OCI_CLI_TENANCY")

	// Get tenancy ID from viper config
	// tenancyId := viper.GetString("tenancy-id")

	// Validate tenancy ID and get tenancy name
	response, err := identityClient.GetTenancy(context.Background(), identity.GetTenancyRequest{TenancyId: &tenancyId})
	CheckError(err)

	Logger.Debug("Current tenancy", "response.Tenancy.Name", *response.Tenancy.Name)
	tenancyName := *response.Tenancy.Name
	// fmt.Println("*** viper.Set tenancy-id ***")
	// viper.Set("tenancy-name", tenancyName)

	// ##################### DEBUG #######################
	// fmt.Println("\n--- ValidateTenancyId --->")
	// cTenancyName := viper.GetString("tenancy-name")
	// fmt.Print("Tenancy name: ")
	// utils.Yellow.Println(cTenancyName)

	// cTenancyId := viper.GetString("tenancy-id")
	// fmt.Print("Tenancy ID: ")
	// utils.Yellow.Println(cTenancyId)

	// compartment := viper.GetString("compartment-name")
	// fmt.Print("Compartment: ")
	// utils.Yellow.Println(compartment)
	// fmt.Println("<---\n")
	// ##################################################

	return tenancyName
}

// 4. Configures config file for Viper config
// 5. ReadsIn config file to Viper config
func ConfigFileLoad(configFilePath string) {

	// Configure config file in Viper config
	viper.AddConfigPath(filepath.Dir(configFilePath))
	viper.SetConfigName(filepath.Base(configFilePath))
	viper.SetConfigType("yaml")

	// Read config file to Viper config
	// fmt.Println("\n*** viper.ReadInConfig ***")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("Error reading config file: ", err)
	} else {
		Logger.Debug("Using oshiv config file", "File", viper.ConfigFileUsed())
	}
}

// 1. Checks for custom config file flag
// 2. Determines default path if no flag is passed
// 3. Creates config file if not exits
func ConfigFileInit() {
	// var configFilePath string

	// Configure default config file path
	// if flagConfigFilePath == "" {
	// 	configFilePath = filepath.Join(HomeDir(), ".oshiv")
	// 	Logger.Debug("Using default config file path")
	// }

	configFilePath := filepath.Join(HomeDir(), ".oshiv")
	Logger.Debug("Config file path set to:" + configFilePath)

	// checkConfigFile(configFilePath)
	Logger.Debug("Checking if config file exists:")
	_, err_stat := os.Stat(configFilePath)

	if err_stat != nil {
		Logger.Debug("Config file doesn't exist, creating config file at " + configFilePath)
		_, err_create := os.Create(configFilePath)
		CheckError(err_create)
	}
}
