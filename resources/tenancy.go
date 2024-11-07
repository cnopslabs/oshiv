package resources

import (
	"context"
	"os"

	"github.com/cnopslabs/oshiv/utils"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/identity"
)

// Determine tenancy ID, validate it against the OCI API, and get tenancy name
func ValidateTenancyId(flagTenancyIdOverride string, client identity.IdentityClient, ociConfig common.ConfigurationProvider) (string, string) {
	var tenancyId string

	// Get default tenancy ID from OCI config
	defaultTenancyId, err := ociConfig.TenancyOCID()
	utils.CheckError(err)

	// Attempt to get tenancy ID from environment variable
	envTenancyIdOverride, envTenancyIdOverrideExists := os.LookupEnv("OCI_CLI_TENANCY")

	// We need a way to override the default tenancy we use to authenticate against. Default tenancy can be overridden by flag or env var
	// Precedence: 1) CLI flag, 2) Environment variable, 3) Default (OCI config file)
	if flagTenancyIdOverride != "" {
		tenancyId = flagTenancyIdOverride
		utils.Logger.Debug("Tenancy ID is set via flag to: " + tenancyId)
	} else if envTenancyIdOverrideExists {
		tenancyId = envTenancyIdOverride
		utils.Logger.Debug("Tenancy ID is set via OCI_CLI_TENANCY environment variable to: " + tenancyId)
	} else {
		tenancyId = defaultTenancyId
		utils.Logger.Debug("Tenancy ID is set via OCI config file to: " + tenancyId)
	}

	// Validate tenancy ID and get tenancy name
	response, err := client.GetTenancy(context.Background(), identity.GetTenancyRequest{TenancyId: &tenancyId})
	utils.CheckError(err)

	utils.Logger.Debug("Current tenancy", "response.Tenancy.Name", *response.Tenancy.Name)
	tenancyName := *response.Tenancy.Name

	return tenancyId, tenancyName
}
