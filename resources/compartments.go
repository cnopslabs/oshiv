package resources

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/cnopslabs/oshiv/utils"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/rodaine/table"
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

	// Provide example to have oshiv use env var for setting compartment
	fmt.Println("\nTo set compartment, export OCI_COMPARTMENT_NAME:")
	utils.Yellow.Println("   export OCI_COMPARTMENT_NAME=")
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

	fmt.Println("\nTo set compartment, export OCI_COMPARTMENT_NAME:")
	utils.Yellow.Println("   export OCI_COMPARTMENT_NAME=")
}
