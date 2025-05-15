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

	// Get tenancy ID from OCI config file and set as the default (lowest precedence order) in viper config
	ociConfigTenancyId, err_config := utils.OciConfig().TenancyOCID()
	utils.CheckError(err_config)
	viper.SetDefault("tenancy-id", ociConfigTenancyId)

	// Attempt to add tenancy ID to Viper config from environment variable (3rd lowest precedence order)
	// Note: OCI_CLI_TENANCY env var follows OCI CLI convention for Tenancy ID
	_, envVarExists := os.LookupEnv("OCI_CLI_TENANCY")
	if envVarExists {
		viper.BindEnv("tenancy-id", "OCI_CLI_TENANCY")
	} else {
		// Since OCI_CLI_TENANCY (ID) is not set by env var, see if OCI_TENANCY_NAME is
		// OCI_TENANCY_NAME does not follow OCI CLI convention but is nicer for humans
		tenancy_from_env, envVarExists := os.LookupEnv("OCI_TENANCY_NAME")

		if envVarExists {
			// Since OCI_TENANCY_NAME is set, let's use it to lookup and set Tenancy ID
			tenancyId, err := utils.LookUpTenancyID(tenancy_from_env)
			utils.CheckError(err)

			// Override tenancy ID
			viper.Set("tenancy-id", tenancyId)
		}
	}

	// Attempt to add compartment to Viper config from environment variable (3rd lowest precedence order)
	viper.BindEnv("compartment", "OCI_COMPARTMENT")

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
