package cmd

import (
	"os"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oshiv",
	Short: "A tool for finding and connecting to OCI resources via the bastion service",
	Long:  "A tool for finding and connecting to OCI resources via the bastion service",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Maybe add version here. Need to research how `cmd` handles version
	},
}

func Execute() {
	// We need to do some initialization prior to executing any sub-commands
	// Note: all sub-commands depend in this initialization

	// 1. Create it if it doesn't exist
	utils.ConfigFileInit()

	// 2. Load config file into Viper config
	// Note: Only compartment is persisted in config file
	utils.ConfigFileRead()

	// 3. Get tenancy ID from OCI config file and set as the default (lowest precedence order) in viper config
	ociConfig := utils.SetupOciConfig()
	ociConfigTenancyId, err_config := ociConfig.TenancyOCID()
	utils.CheckError(err_config)
	viper.SetDefault("tenancy-id", ociConfigTenancyId)

	// 4. Attempt to add tenancy ID to Viper config from environment variable (3rd lowest precedence order)
	viper.BindEnv("tenancy-id", "OCI_CLI_TENANCY")

	// Execute adds all child commands to the root command and sets flags appropriately.
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// We need a way to override the default tenancy that may be used to authenticate with
	// One way to do that is to provide a flag for Tenancy ID
	// Tenancy ID (default or override) is required by all OCI API calls
	var flagTenancyId string
	rootCmd.PersistentFlags().StringVarP(&flagTenancyId, "tenancy-id", "t", "", "Override's the default tenancy with this tenancy ID")

	// Compartment is required by all OCI API calls except for compartment list
	var flagCompartmentName string
	rootCmd.PersistentFlags().StringVarP(&flagCompartmentName, "compartment", "c", "", "The name of the compartment to use")
}
