package resources

import (
	"context"
	"fmt"
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
	// TODO: pagination
	var clusters []Cluster

	initialResponse, err := containerEngineClient.ListClusters(context.Background(), containerengine.ListClustersRequest{
		CompartmentId: &compartmentId,
		// LifecycleState: core.InstanceLifecycleStateRunning,
	})
	utils.CheckError(err)

	for _, cluster := range initialResponse.Items {
		// fmt.Println(cluster)/

		clusterId := *cluster.Id
		clusterName := *cluster.Name
		// clusterPrivateEndpoint := *cluster.Endpoints.PrivateEndpoint

		clusterPrivateEndpointIp, clusterPrivateEndpointPort, found := strings.Cut(*cluster.Endpoints.PrivateEndpoint, ":")
		// clusterPrivateEndpointPort := *cluster.Endpoints.PrivateEndpoint

		if found {
			cluster := Cluster{clusterName, clusterId, clusterPrivateEndpointIp, clusterPrivateEndpointPort}
			clusters = append(clusters, cluster)
		}
	}

	return clusters
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

// Find and print k8s clusters
func FindClusters(containerEngineClient containerengine.ContainerEngineClient, compartmentId string, searchString string) {
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

	if len(clusterMatches) > 0 {
		for _, cluster := range clusterMatches {
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
