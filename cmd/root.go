package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oshiv",
	Short: "A tool for finding and connecting to OCI resources",
	Long:  "A tool for finding OCI resources and for connecting to instances and OKE clusters via the OCI bastion service.",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oshiv.yaml)")

	// We need a way to override the default tenancy that we use to authenticate against
	// One way to do that is to provide a flag for Tenancy ID
	var FlagTenancyIdOverride string
	rootCmd.PersistentFlags().StringVarP(&FlagTenancyIdOverride, "tenancy-id-override", "t", "", "Override's the default tenancy with this tenancy ID")
	// rootCmd.MarkFlagRequired("tenancy-id")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
