package resources

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/rodaine/table"
	"github.com/spf13/viper"
)

// Fetch compartments (names/IDs) from OCI
func FetchCompartments(tenancyId string, identityClient identity.IdentityClient) map[string]string {
	response, err := identityClient.ListCompartments(context.Background(), identity.ListCompartmentsRequest{CompartmentId: &tenancyId})
	utils.CheckError(err)

	compartments := make(map[string]string)

	for _, item := range response.Items {
		compartments[*item.Name] = *item.Id
	}

	return compartments
}

// Sort, list, and print compartments
func ListCompartments(compartments map[string]string, tenancyId string, tenancyName string) {
	tbl := table.New("Compartment Name", "OCID")
	tbl.WithHeaderFormatter(utils.HeaderFmt).WithFirstColumnFormatter(utils.ColumnFmt)

	compartmentNames := make([]string, 0, len(compartments))
	for compartmentName := range compartments {
		compartmentNames = append(compartmentNames, compartmentName)
	}
	sort.Strings(compartmentNames)

	// List the root compartment first
	tbl.AddRow(tenancyName, tenancyId)

	for _, compartmentName := range compartmentNames {
		tbl.AddRow(compartmentName, compartments[compartmentName])
	}

	utils.FaintMagenta.Println("Tenancy: " + tenancyName)
	tbl.Print()

	fmt.Println("\nTo set compartment, run:")
	utils.Yellow.Println("   oshiv compartment -s COMPARTMENT_NAME")
}

func FindCompartments(tenancyId string, tenancyName string, identityClient identity.IdentityClient, namePattern string) {
	compartments := FetchCompartments(tenancyId, identityClient)

	var matches []string

	if namePattern == "*" {
		namePattern = ".*"
	}

	for name := range compartments {
		match, _ := regexp.MatchString(namePattern, name)
		if match {
			matches = append(matches, name)
		}
	}

	matchCount := len(matches)
	utils.Faint.Println(strconv.Itoa(matchCount) + " matches")

	tbl := table.New("Compartment Name", "OCID")
	tbl.WithHeaderFormatter(utils.HeaderFmt).WithFirstColumnFormatter(utils.ColumnFmt)

	for _, compartmentName := range matches {
		compartmentId := compartments[compartmentName]
		tbl.AddRow(compartmentName, compartmentId)
	}

	utils.FaintMagenta.Println("Tenancy: " + tenancyName)
	tbl.Print()

	fmt.Println("\nTo set compartment, run:")
	utils.Yellow.Println("   oshiv compartment -s COMPARTMENT_NAME")
}

// Sets compartment name in Viper config
func SetCompartmentName(compartmentName string) {
	utils.Logger.Debug("Setting compartment: " + compartmentName)
	utils.SetConfigString("compartment-name", compartmentName)
}

// Lookup compartment ID by name
func lookupCompartmentId(compartments map[string]string, compartmentName string) string {
	compartmentId := compartments[compartmentName]
	utils.Logger.Debug(compartmentName + "'s compartment ID is " + compartmentId)

	return compartmentId
}

// Determine compartment name or ID, lookup name from ID if ID is given
func DetermineCompartment(compartments map[string]string, identityClient identity.IdentityClient, tenancyId string, tenancyName string) (string, string) {
	var compartmentId string

	// TODO: Should we move the setting of Viper config up into the cmd package?
	// Viper uses the following order precedence: 1) flag, 2) env var, 3) config file, 4) key/value store, 4) default
	// For compartment-name , we are currently only supporting 1 and 4 as there is no OCI CLI convention for env var
	// If compartment name flag was passed, this has already been added to config as flag

	// Use tenancy (I.e. the root compartment) as the default (lowest precedence order) in viper config
	viper.SetDefault("compartment-name", tenancyName)

	// Get compartment name from viper config
	compartmentName := viper.GetString("compartment-name")

	// Handle root compartment
	if compartmentName == tenancyName {
		compartmentId = tenancyId
	} else {
		compartmentId = lookupCompartmentId(compartments, compartmentName)
	}

	utils.Logger.Debug("Compartment: " + compartmentName + "(" + compartmentId + ")")

	return compartmentId, compartmentName
}
