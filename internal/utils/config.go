package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Compartment string `yaml:"compartment"`
}

var configFilePath string = filepath.Join(HomeDir(), ".oshiv")

// Create config file if it doesn't exist
func ConfigFileInit() {
	Logger.Debug("Config file path set to:" + configFilePath)
	Logger.Debug("Checking if config file exists:")
	_, err_stat := os.Stat(configFilePath)

	if err_stat != nil {
		Logger.Debug("Config file doesn't exist, creating config file at " + configFilePath)
		_, err_create := os.Create(configFilePath)
		CheckError(err_create)
	}
}

// Validate tenancy ID using the OCI API and lookup/return tenancy name
func validateTenancyId(identityClient identity.IdentityClient, tenancyId string) string {
	response, err := identityClient.GetTenancy(context.Background(), identity.GetTenancyRequest{TenancyId: &tenancyId})
	CheckError(err)

	Logger.Debug("Current tenancy", "response.Tenancy.Name", *response.Tenancy.Name)
	tenancyName := *response.Tenancy.Name
	return tenancyName
}

// Load config file into Viper config
// Note: Only compartment is persisted in config file
func ConfigFileRead() {
	// Configure config file in Viper config
	viper.AddConfigPath(filepath.Dir(configFilePath))
	viper.SetConfigName(filepath.Base(configFilePath))
	viper.SetConfigType("yaml")

	// Read config file to Viper config
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("Error reading config file: ", err)
	} else {
		Logger.Debug("Using oshiv config file", "File", viper.ConfigFileUsed())
	}
}

// Reset config file and wite compartment to file
// Note: Since we only care about persisting compartment to file, we will not use Viper to
// write the config, which would persist all Viper configuration
func WriteCompartmentToFile(compartment string, compartments map[string]string) {
	os.Truncate(configFilePath, 0)

	if compartment != "" {
		compartmentIsValid := validateCompartment(compartment, compartments)
		if !compartmentIsValid {
			fmt.Println("Invalid compartment: " + compartment)
			fmt.Println("Does the " + compartment + " compartment exist in tenancy?")
			fmt.Println("Note: root compartment (I.e. tenancy) will be used if no valid compartment is set")
			os.Exit(0)
		}

		var config Config
		config.Compartment = compartment

		// Marshal the updated struct to YAML
		data, err := yaml.Marshal(config)
		CheckError(err)

		// Write the updated YAML data back to the file
		err = os.WriteFile(configFilePath, data, 0644)
		CheckError(err)
	}
}

func SetTenancyConfig(FlagTenancyId *pflag.Flag, ociConfig common.ConfigurationProvider) {
	// Add tenancy ID to Viper config
	viper.BindPFlag("tenancy-id", FlagTenancyId)
	// Determine tenancyId from Viper: 2)flag, 3)ENV, 4)file, 6) default
	tenancyId := viper.GetString("tenancy-id")

	// Validate tenancy and get tenancy name
	identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
	CheckError(identityErr)
	tenancyName := validateTenancyId(identityClient, tenancyId)
	viper.Set("tenancy-name", tenancyName)
}

func SetCompartmentConfig(FlagCompartment *pflag.Flag, compartments map[string]string, tenancyName string) {
	viper.BindPFlag("compartment", FlagCompartment)

	// Determine compartment from Viper: 2)flag, 4)file
	compartment := viper.GetString("compartment")

	// Validate compartment
	compartmentIsValid := validateCompartment(compartment, compartments)

	if !compartmentIsValid {
		// The compartment is not in the current tenant, use tenancy (root compartment)
		// This occurs if compartment was set in a file (or flag) but Tenancy has changed
		viper.Set("compartment", tenancyName)

		// Reset config file
		os.Truncate(configFilePath, 0)
	}
}

func validateCompartment(compartment string, compartments map[string]string) bool {
	var compartmentIsValid bool = false

	for valid_compartment := range compartments {
		if compartment == valid_compartment {
			compartmentIsValid = true
			break
		}
	}

	return compartmentIsValid
}
