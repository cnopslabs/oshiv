package utils

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Validate tenancy ID using the OCI API and lookup/return tenancy name
func validateTenancyId(identityClient identity.IdentityClient, tenancyId string) string {
	response, err := identityClient.GetTenancy(context.Background(), identity.GetTenancyRequest{TenancyId: &tenancyId})
	CheckError(err)

	Logger.Debug("Current tenancy", "response.Tenancy.Name", *response.Tenancy.Name)
	tenancyName := *response.Tenancy.Name
	return tenancyName
}

func SetTenancyConfig(FlagTenancyId *pflag.Flag, ociConfig common.ConfigurationProvider) {
	// Add tenancy ID to Viper config
	viper.BindPFlag("tenancy-id", FlagTenancyId)
	// Determine tenancyId from Viper: 2)flag, 3)ENV, 4)file, 6) default
	tenancyId := viper.GetString("tenancy-id")

	// Validate tenancy and get tenancy name
	identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
	CheckError(identityErr)
	tenancyName := validateTenancyId(identityClient, tenancyId)
	viper.Set("tenancy-name", tenancyName)
}

func SetCompartmentConfig(FlagCompartment *pflag.Flag, compartments map[string]string, tenancyName string) {
	viper.BindPFlag("compartment", FlagCompartment)

	// Determine compartment from Viper: 2)flag, 4)file
	compartment := viper.GetString("compartment")

	// Validate compartment
	compartmentIsValid := validateCompartment(compartment, compartments)

	if !compartmentIsValid {
		// The compartment is not in the current tenant, use tenancy (root compartment)
		// This occurs if compartment was set in a file (or flag) but Tenancy has changed
		viper.Set("compartment", tenancyName)
	}
}

func validateCompartment(compartment string, compartments map[string]string) bool {
	var compartmentIsValid bool = false

	for valid_compartment := range compartments {
		if compartment == valid_compartment {
			compartmentIsValid = true
			break
		}
	}

	return compartmentIsValid
}
