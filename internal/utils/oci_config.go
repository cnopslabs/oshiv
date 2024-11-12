package utils

import (
	"os"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func SetupOciConfig() common.ConfigurationProvider {
	var config common.ConfigurationProvider

	profile, envVarExists := os.LookupEnv("OCI_CLI_PROFILE")

	if envVarExists {
		Logger.Debug("Using profile " + profile)
		configPath := HomeDir() + "/.oci/config"
		config = common.CustomProfileConfigProvider(configPath, profile)
	} else {
		Logger.Debug("Using default profile")
		config = common.DefaultConfigProvider()
	}

	return config
}
