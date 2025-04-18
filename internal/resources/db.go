package resources

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/database"
)

type Database struct {
	name              string
	id                string
	privateEndpointIp string
	connectStrings    map[string]string
	profiles          []database.DatabaseConnectionStringProfile
}

// Fetch all databases via OCI API call
func fetchDatabases(databaseClient database.DatabaseClient, compartmentId string) []Database {
	var databases []Database

	initialResponse, err := databaseClient.ListAutonomousDatabases(context.Background(), database.ListAutonomousDatabasesRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

	for _, database := range initialResponse.Items {
		databaseName := *database.DbName
		databaseId := *database.Id
		databaseIp := *database.PrivateEndpointIp
		databaseConnectStrings := database.ConnectionStrings.AllConnectionStrings
		databaseProfiles := database.ConnectionStrings.Profiles

		database := Database{databaseName, databaseId, databaseIp, databaseConnectStrings, databaseProfiles}
		databases = append(databases, database)
	}

	if initialResponse.OpcNextPage != nil {
		nextPage := initialResponse.OpcNextPage

		for {
			response, err := databaseClient.ListAutonomousDatabases(context.Background(), database.ListAutonomousDatabasesRequest{CompartmentId: &compartmentId, Page: nextPage})
			utils.CheckError(err)

			for _, database := range response.Items {
				databaseName := *database.DbName
				databaseId := *database.Id
				databaseIp := *database.PrivateEndpoint
				databaseConnectStrings := database.ConnectionStrings.AllConnectionStrings
				databaseProfiles := database.ConnectionStrings.Profiles

				database := Database{databaseName, databaseId, databaseIp, databaseConnectStrings, databaseProfiles}
				databases = append(databases, database)
			}

			if response.OpcNextPage != nil {
				nextPage = response.OpcNextPage
			} else {
				break
			}
		}
	}

	return databases
}

// Match pattern and return database matches
func matchDatabases(pattern string, databases []Database) []Database {
	var matches []Database

	// Handle simple wildcard
	if pattern == "*" {
		pattern = ".*"
	}

	for _, database := range databases {
		match, _ := regexp.MatchString("(?i)"+pattern, database.name)
		if match {
			matches = append(matches, database)
		}
	}

	return matches
}

// Find databases matching search pattern
func FindDatabases(databaseClient database.DatabaseClient, compartmentId string, searchString string) []Database {
	var databaseMatches []Database
	databases := fetchDatabases(databaseClient, compartmentId)

	if searchString != "" {
		// Find matching databases
		pattern := searchString
		databaseMatches = matchDatabases(pattern, databases)
		utils.Faint.Println(strconv.Itoa(len(databaseMatches)) + " matches")
	} else {
		// List all databases
		databaseMatches = databases
		utils.Faint.Println(strconv.Itoa(len(databaseMatches)) + " databases(s)")
	}

	return databaseMatches
}

// Print databases
func PrintDatabases(databases []Database, tenancyName string, compartmentName string) {
	if len(databases) > 0 {
		utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartmentName + ")")

		for _, database := range databases {
			fmt.Print("Name: ")
			utils.Blue.Println(database.name)
			fmt.Print("Database ID: ")
			utils.Yellow.Println(database.id)
			fmt.Print("Private endpoint: ")
			utils.Yellow.Println(database.privateEndpointIp)

			for serviceType, connectString := range database.connectStrings {
				connectStringParts := strings.Split(connectString, "/")
				serviceName := connectStringParts[1]
				commonNamePort := connectStringParts[0]
				commonNamePortParts := strings.Split(commonNamePort, ":")
				commonName := commonNamePortParts[0]

				// Use "High" service for admin / troubleshooting
				if serviceType == "HIGH" {
					fmt.Print("Service name: ")
					utils.Yellow.Println(serviceName)

					fmt.Print("Common name (CN): ")
					utils.Yellow.Println(commonName)
				}
			}

			fmt.Println("")
			fmt.Println("Connect strings:")

			for _, profile := range database.profiles {
				// Use "High" service for admin / troubleshooting
				if strings.Contains(*profile.DisplayName, "high") {
					if strings.Contains(*profile.Value, "1521") {
						utils.Italic.Println("Standard")
						utils.Yellow.Println(*profile.Value)
					}

					if strings.Contains(*profile.Value, "1522") {
						fmt.Println("")
						utils.Italic.Println("MTLS")
						utils.Yellow.Println(*profile.Value)
					}
				}
			}
		}
	}
}
