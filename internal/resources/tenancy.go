package resources

import (
	"context"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/viper"
)

// Determine tenancy ID, validate it against the OCI API, and get tenancy name
func ValidateTenancyId(client identity.IdentityClient, ociConfig common.ConfigurationProvider) (string, string) {
	// Check for tenancy overrides
	// Viper uses the following order precedence: 1) flag, 2) env var, 3) config file, 4) key/value store, 4) default
	// For tenancy-id, we are currently only supporting 1, 2, and 4
	// If tenancy ID flag was passed, this has already been added to config as flag

	// Get tenancy ID from OCI config and set as the default (lowest precedence order) in viper config
	ociConfigTenancyId, err := ociConfig.TenancyOCID()
	utils.CheckError(err)
	viper.SetDefault("tenancy-id", ociConfigTenancyId)

	// Attempt to get tenancy ID from environment variable
	viper.BindEnv("tenancy-id", "OCI_CLI_TENANCY")

	// Get tenancy ID from viper config
	tenancyId := viper.GetString("tenancy-id")

	// Validate tenancy ID and get tenancy name
	response, err := client.GetTenancy(context.Background(), identity.GetTenancyRequest{TenancyId: &tenancyId})
	utils.CheckError(err)

	utils.Logger.Debug("Current tenancy", "response.Tenancy.Name", *response.Tenancy.Name)
	tenancyName := *response.Tenancy.Name

	return tenancyId, tenancyName
}
