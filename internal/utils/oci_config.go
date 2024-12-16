package utils

import (
	"os"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func OciConfig() common.ConfigurationProvider {
	var config common.ConfigurationProvider
	profile := OciProfile()

	if profile == "DEFAULT" { // TODO: Do I actually need this? How is DefaultConfigProvider different
		Logger.Debug("Using default profile")
		config = common.DefaultConfigProvider()
	} else {
		Logger.Debug("Using profile " + profile)
		configPath := HomeDir() + "/.oci/config"
		config = common.CustomProfileConfigProvider(configPath, profile)
	}

	return config
}

func OciProfile() string {
	profile, envVarExists := os.LookupEnv("OCI_CLI_PROFILE")

	if envVarExists {
		return profile
	} else {
		return "DEFAULT"
	}
}
