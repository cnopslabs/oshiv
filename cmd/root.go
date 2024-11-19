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
	Short: "A tool for finding and connecting to OCI resources",
	Long:  "A tool for finding OCI resources and for connecting to instances and OKE clusters via the OCI bastion service.",
	Run:   func(cmd *cobra.Command, args []string) {}, // TODO: Maybe add version here
}

func Execute() {
	// We need to initialize the config file prior to executing any child commands
	flagConfigFilePath, _ := rootCmd.Flags().GetString("config")
	utils.ConfigInit(flagConfigFilePath)

	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	// Only add tenancy ID to config if its passed in
	flagTenancyId, _ := rootCmd.Flags().GetString("tenancy-id")
	if flagTenancyId != "" {
		viper.BindPFlag("tenancy-id", rootCmd.Flags().Lookup("tenancy-id"))
	}

	// Only add compartment name to config if its passed in
	flagCompartmentName, _ := rootCmd.Flags().GetString("compartment-name")
	if flagCompartmentName != "" {
		viper.BindPFlag("compartment-name", rootCmd.Flags().Lookup("compartment-name"))
	}
}

func init() {
	// Config file is required by all OCI API resource functions except compartment list
	var flagConfigFilePath string
	rootCmd.PersistentFlags().StringVarP(&flagConfigFilePath, "config", "i", "", "config file (default is $HOME/.oshiv.yaml)")

	// We need a way to override the default tenancy that may be used to authenticate against
	// One way to do that is to provide a flag for Tenancy ID
	// Tenancy ID (default or override) is required by all OCI API resources
	var flagTenancyId string
	rootCmd.PersistentFlags().StringVarP(&flagTenancyId, "tenancy-id", "t", "", "Override's the default tenancy with this tenancy ID")

	// Compartment is required by all OCI API resource functions except compartment list
	var flagCompartmentName string
	rootCmd.PersistentFlags().StringVarP(&flagCompartmentName, "compartment-name", "c", "", "The name of the compartment to use")
}
