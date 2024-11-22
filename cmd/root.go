package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oshiv",
	Short: "A tool for finding and connecting to OCI resources",
	Long:  "A tool for finding OCI resources and for connecting to instances and OKE clusters via the OCI bastion service.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: This never gets executed")
	}, // TODO: Maybe add version here
}

func Execute() {
	// Note: Viper uses the following order precedence: 1) flag, 2) env var, 3) config file, 4) key/value store, 4) default
	// tenancy ID flag has to be looked up after rootCmd.Execute (in a sub-command run)

	// We need to do some initialization prior to executing any child commands
	// Note: all child commands depend in this initialization
	// 1. Determine config file and create it if it doesn't exist
	// 2. Load config file into Viper config (this may set tenancy-id, tenancy-name, and/or compartment)
	// 3. Get tenancy ID from OCI config file and set as default in Viper config
	// 4. Add tenancy ID to Viper config from environment variable (if exists)

	// 1.
	// If config file doesn't exist, we need to create it
	// flagConfigFilePath, _ := rootCmd.Flags().GetString("config") TODO: rm
	utils.ConfigFileInit()

	// 2.
	// Load config from file into Viper config
	configFilePath := filepath.Join(utils.HomeDir(), ".oshiv")
	utils.ConfigFileLoad(configFilePath)

	// 3.
	// Get tenancy ID from OCI config file and set as the default (lowest precedence order) in viper config
	ociConfig := utils.SetupOciConfig()
	ociConfigTenancyId, err_config := ociConfig.TenancyOCID()
	utils.CheckError(err_config)
	viper.SetDefault("tenancy-id", ociConfigTenancyId)

	// 4.
	// Attempt to add tenancy ID to Viper config from environment variable
	viper.BindEnv("tenancy-id", "OCI_CLI_TENANCY")

	// ########################## rootCmd.Execute ##############################################################################
	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.
	// fmt.Println("--> calling rootCmd.Execute")
	err := rootCmd.Execute()
	// fmt.Println("<-- rootCmd.Execute finished")
	if err != nil {
		os.Exit(1)
	}
	// #########################################################################################################################
}

func init() {
	// fmt.Println("cmd.init(root.go) called")
	// Config file is required by all OCI API resource functions except compartment list
	// var flagConfigFilePath string
	// rootCmd.PersistentFlags().StringVarP(&flagConfigFilePath, "config", "i", "", "config file (default is $HOME/.oshiv.yaml)")

	// We need a way to override the default tenancy that may be used to authenticate against
	// One way to do that is to provide a flag for Tenancy ID
	// Tenancy ID (default or override) is required by all OCI API resources
	var flagTenancyId string
	rootCmd.PersistentFlags().StringVarP(&flagTenancyId, "tenancy-id", "t", "", "Override's the default tenancy with this tenancy ID")
	// rootCmd.Flags().StringP("tenancy-id", "t", "", "Override's the default tenancy with this tenancy ID")

	// Compartment is required by all OCI API resource functions except compartment list
	var flagCompartmentName string
	rootCmd.PersistentFlags().StringVarP(&flagCompartmentName, "compartment", "c", "", "The name of the compartment to use")
}
