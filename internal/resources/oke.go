package resources

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
)

type Cluster struct {
	name                string
	id                  string
	privateEndpointIp   string
	privateEndpointPort string
}

// Fetch all clusters via OCI API call
func fetchClusters(containerEngineClient containerengine.ContainerEngineClient, compartmentId string) []Cluster {
	var clusters []Cluster

	initialResponse, err := containerEngineClient.ListClusters(context.Background(), containerengine.ListClustersRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

	for _, cluster := range initialResponse.Items {
		clusterId := *cluster.Id
		clusterName := *cluster.Name

		clusterPrivateEndpointIp, clusterPrivateEndpointPort, found := strings.Cut(*cluster.Endpoints.PrivateEndpoint, ":")
		if found {
			cluster := Cluster{clusterName, clusterId, clusterPrivateEndpointIp, clusterPrivateEndpointPort}
			clusters = append(clusters, cluster)
		}
	}

	if initialResponse.OpcNextPage != nil {
		nextPage := initialResponse.OpcNextPage

		for {
			response, err := containerEngineClient.ListClusters(context.Background(), containerengine.ListClustersRequest{CompartmentId: &compartmentId, Page: nextPage})
			utils.CheckError(err)

			for _, cluster := range response.Items {
				clusterId := *cluster.Id
				clusterName := *cluster.Name

				clusterPrivateEndpointIp, clusterPrivateEndpointPort, found := strings.Cut(*cluster.Endpoints.PrivateEndpoint, ":")
				if found {
					cluster := Cluster{clusterName, clusterId, clusterPrivateEndpointIp, clusterPrivateEndpointPort}
					clusters = append(clusters, cluster)
				}
			}

			if response.OpcNextPage != nil {
				nextPage = response.OpcNextPage
			} else {
				break
			}
		}
	}

	return clusters
}

func FetchClusterId(containerEngineClient containerengine.ContainerEngineClient, compartmentId string, clusterName string) string {
	clusters := fetchClusters(containerEngineClient, compartmentId)
	var clusterId string

	for _, cluster := range clusters {
		if cluster.name == clusterName {
			clusterId = cluster.id
		} else {
			fmt.Println("Unable to find ID of " + clusterName)
			os.Exit(1)
		}
	}

	return clusterId
}

// Match pattern and return cluster matches
func matchClusters(pattern string, clusters []Cluster) []Cluster {
	var matches []Cluster

	// Handle simple wildcard
	if pattern == "*" {
		pattern = ".*"
	}

	for _, cluster := range clusters {
		match, _ := regexp.MatchString(pattern, cluster.name)
		if match {
			matches = append(matches, cluster)
		}
	}

	return matches
}

// Find clusters matching search pattern
func FindClusters(containerEngineClient containerengine.ContainerEngineClient, compartmentId string, searchString string) []Cluster {
	var clusterMatches []Cluster
	clusters := fetchClusters(containerEngineClient, compartmentId)

	if searchString != "" {
		// Find matching clusters
		pattern := searchString
		clusterMatches = matchClusters(pattern, clusters)
		utils.Faint.Println(strconv.Itoa(len(clusterMatches)) + " matches")
	} else {
		// List all clusters
		clusterMatches = clusters
		utils.Faint.Println(strconv.Itoa(len(clusterMatches)) + " cluster(s)")
	}

	return clusterMatches
}

// Print clusters
func PrintClusters(clusters []Cluster, tenancyName string, compartmentName string) {
	if len(clusters) > 0 {
		utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartmentName + ")")

		for _, cluster := range clusters {
			fmt.Print("Name: ")
			utils.Blue.Println(cluster.name)
			fmt.Print("Cluster ID: ")
			utils.Yellow.Println(cluster.id)
			fmt.Print("Private endpoint: ")
			utils.Yellow.Println(cluster.privateEndpointIp + ":" + cluster.privateEndpointPort)
			fmt.Println("")
		}
	}
}
