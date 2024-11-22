package cmd

import (
	"fmt"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var compartmentCmd = &cobra.Command{
	Use:   "compartment",
	Short: "Find and list compartments",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		// Lookup tenancy ID and compartment flags and add to Viper config if passed
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		FlagCompartment := rootCmd.Flags().Lookup("compartment") // TODO: FlagCompartment is not used in compartment command
		utils.ConfigInit(FlagTenancyId, FlagCompartment)         // TODO: make ConfigInit handle missing compartment

		// Get tenancy ID and tenancy name from Viper config
		tenancyName := viper.GetString("tenancy-name")
		tenancyId := viper.GetString("tenancy-id")

		ociConfig := utils.SetupOciConfig()
		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
		utils.CheckError(identityErr)

		compartments := resources.FetchCompartments(tenancyId, identityClient)

		flagList, _ := cmd.Flags().GetBool("list")
		flagFind, _ := cmd.Flags().GetString("find")
		// flagSetCompartment, _ := cmd.Flags().GetString("set-compartment")
		flagSetCompartment := cmd.Flags().Lookup("set-compartment")

		if flagList {
			resources.ListCompartments(compartments, tenancyId, tenancyName)
		} else if flagFind != "" {
			resources.FindCompartments(tenancyId, tenancyName, identityClient, flagFind)
		} else if flagSetCompartment.Changed {
			// resources.SetCompartmentName(flagSetCompartment)
			viper.BindPFlag("compartment", flagSetCompartment)
			viper.WriteConfig()
		} else {
			fmt.Println("Invalid sub-command or flag")
		}
	},
}

func init() {
	rootCmd.AddCommand(compartmentCmd)

	// Local flags only exposed to compartment command
	compartmentCmd.Flags().BoolP("list", "l", false, "List all compartments")
	compartmentCmd.Flags().StringP("find", "f", "", "Find compartment by name pattern search")
	var flagSetCompartment string
	compartmentCmd.Flags().StringVarP(&flagSetCompartment, "set-compartment", "s", "", "Set compartment name")
}
