package resources

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/database"
)

type Database struct {
	name              string
	id                string
	privateEndpointIp string
	port              string
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
		databasePort := getDatabasePort(*database.ConnectionStrings)

		database := Database{databaseName, databaseId, databaseIp, databasePort}
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
				databasePort := getDatabasePort(*database.ConnectionStrings)

				database := Database{databaseName, databaseId, databaseIp, databasePort}
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

// Get database port from connect strings
func getDatabasePort(connStrings database.AutonomousDatabaseConnectionStrings) string {
	profile := connStrings.Profiles[0].Value // TODO: this is pulling the first profile of many which is typically be the "high" profile. Not sure if this is standardized.
	databasePort := matchDbPort(*profile)

	return databasePort
}

// Match and return the database's service port from the connect string
func matchDbPort(dbConnectString string) string {
	// Matches something like "port=1522" in connect string
	re := regexp.MustCompile(`port=(\d+)`)
	match := re.FindStringSubmatch(dbConnectString)
	if len(match) > 1 {
		return match[1]
	}
	return ""
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
			fmt.Print("Port: ")
			utils.Yellow.Println(database.port)
			fmt.Println("")
		}
	}
}
