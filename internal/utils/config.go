package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

func SetConfigString(name string, value string) {
	viper.Set(name, value)
	viper.WriteConfig()
}

func configInit() {
	viper.AddConfigPath(HomeDir())
	viper.SetConfigName(".oshiv")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()

	if err != nil {
		_, fileNotFound := err.(viper.ConfigFileNotFoundError)

		if fileNotFound {
			fmt.Println("Config file doesn't exists")
		} else {
			fmt.Println("Error reading config file: ", err)
		}
	} else {
		Logger.Debug("Using oshiv config file", "File", viper.ConfigFileUsed())
	}
}
