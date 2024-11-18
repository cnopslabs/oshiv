package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Add config to Viper config
func SetConfigString(name string, value string) {
	viper.Set(name, value)
	Logger.Debug("Writing config: " + name + ": " + value)
	viper.WriteConfig()
}

// Initialize configuration
func ConfigInit(flagConfigFilePath string) {
	var configFilePath string

	// Configure default config file path
	if flagConfigFilePath == "" {
		configFilePath = filepath.Join(HomeDir(), ".oshiv")
		Logger.Debug("Using default config file path")
	}

	Logger.Debug("Config file path set to:" + configFilePath)
	checkConfigFile(configFilePath)

	// Configure config file for Viper config
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

// Check if config file exists. If it doesn't, create it
func checkConfigFile(configFilePath string) {
	Logger.Debug("Checking if config file exists:")
	_, err := os.Stat(configFilePath)

	if err != nil {
		Logger.Debug("Config file doesn't exist, creating config file at " + configFilePath)
		_, err := os.Create(configFilePath)
		CheckError(err)
	}
}
