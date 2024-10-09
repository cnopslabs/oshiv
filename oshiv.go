package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/oracle/oci-go-sdk/v65/bastion"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/rodaine/table"
)

const logLevel = "INFO" // TODO: switch to logging library

var version = "undefined" // Version gets automatically updated during build

var blue = color.New(color.FgCyan)
var yellowBold = color.New(color.FgYellow, color.Bold)
var yellow = color.New(color.FgYellow)
var faint = color.New(color.Faint)
var faintUnder = color.New(color.Faint, color.Underline, color.Italic)
var headerFmt = color.New(color.FgCyan, color.Underline).SprintfFunc()
var columnFmt = color.New(color.FgYellow).SprintfFunc()

type SessionInfo struct {
	state bastion.SessionLifecycleStateEnum
	ip    string
	user  string
	port  int
}

type Cluster struct {
	name                string
	id                  string
	privateEndpointIp   string
	privateEndpointPort string
}

type Image struct {
	name        string
	id          string
	cDate       common.SDKTime
	freeTags    map[string]string
	definedTags map[string]map[string]interface{}
	launchMode  core.ImageLaunchModeEnum
}

type Instance struct {
	name    string
	id      string
	ip      string
	ad      string
	shape   string
	cDate   common.SDKTime
	imageId string
	fd      string
	vCPUs   int
	mem     float32
	region  string
	state   core.InstanceLifecycleStateEnum
}

// Sort instances by name
type instancesByName []Instance

func (instances instancesByName) Len() int           { return len(instances) }
func (instances instancesByName) Less(i, j int) bool { return instances[i].name < instances[j].name }
func (instances instancesByName) Swap(i, j int) {
	instances[i], instances[j] = instances[j], instances[i]
}

type Subnet struct {
	cidr       string
	name       string
	access     string
	subnetType string
}

// TODO: This sorts alphabetically, so not great for CIDR blocks. Prob should revert to sort by name
// Sort subnets bt CIDR
type subnetsByCidr []Subnet

func (subnets subnetsByCidr) Len() int           { return len(subnets) }
func (subnets subnetsByCidr) Less(i, j int) bool { return subnets[i].cidr < subnets[j].cidr }
func (subnets subnetsByCidr) Swap(i, j int)      { subnets[i], subnets[j] = subnets[j], subnets[i] }

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func homeDir() string {
	homeDir, err := os.UserHomeDir()
	checkError(err)

	return homeDir
}

func configureProvider() common.ConfigurationProvider {
	var config common.ConfigurationProvider

	profile, exists := os.LookupEnv("OCI_CLI_PROFILE")

	if exists {
		if logLevel == "DEBUG" {
			fmt.Println("Using profile " + profile)
		}

		configPath := homeDir() + "/.oci/config"

		config = common.CustomProfileConfigProvider(configPath, profile)
	} else {
		if logLevel == "DEBUG" {
			fmt.Println("Using default profile")
		}
		config = common.DefaultConfigProvider()
	}

	return config
}

func parseSubcommand(firstArg string) string {
	if strings.HasPrefix(firstArg, "-") {
		return ""
	} else {
		return firstArg
	}
}

func validateTenancyId(tenancyIdFlag string, client identity.IdentityClient) string {
	tenancyId, exists := os.LookupEnv("OCI_CLI_TENANCY")
	if !exists {
		if tenancyIdFlag == "" {
			fmt.Println("Must pass tenancy ID with -t or set with environment variable OCI_CLI_TENANCY")
			os.Exit(1)
		}
	} else {
		if logLevel == "DEBUG" {
			fmt.Println("\nTenancy ID is set via OCI_CLI_TENANCY to: " + tenancyId)
		}
	}

	// Validate tenancy ID
	response, err := client.GetTenancy(context.Background(), identity.GetTenancyRequest{TenancyId: &tenancyId})
	checkError(err)

	if logLevel == "DEBUG" {
		fmt.Println("\nCurrent tenant: " + *response.Tenancy.Name)
	}

	return tenancyId
}

func checkCompartmentName(flagCompartmentName string) string {
	var compartmentName string
	compartmentIdEnv, exists := os.LookupEnv("OCI_COMPARTMENT_NAME")
	if exists {
		compartmentName = compartmentIdEnv
		if logLevel == "DEBUG" {
			fmt.Println("Compartment name is set via OCI_COMPARTMENT_NAME to: " + compartmentName)
		}
	} else if flagCompartmentName == "" {
		fmt.Println("Must pass compartment name with -c or set with environment variable OCI_COMPARTMENT_NAME")
		os.Exit(1)
	} else {
		compartmentName = flagCompartmentName
	}

	return compartmentName
}

func fetchCompartmentInfo(tenancyId string, client identity.IdentityClient) map[string]string {
	response, err := client.ListCompartments(context.Background(), identity.ListCompartmentsRequest{CompartmentId: &tenancyId})
	checkError(err)

	compartmentInfo := make(map[string]string)

	for _, item := range response.Items {
		compartmentInfo[*item.Name] = *item.Id
	}

	return compartmentInfo
}

func lookupCompartmentId(compartmentInfo map[string]string, compartmentName string) string {
	compartmentId := compartmentInfo[compartmentName]
	if logLevel == "DEBUG" {
		fmt.Println("\n" + compartmentName + "'s compartment ID is " + compartmentId)
	}

	return compartmentId
}

func listCompartmentNames(compartmentInfo map[string]string) {
	tbl := table.New("Compartment Name", "OCID")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	compartmentNames := make([]string, 0, len(compartmentInfo))
	for compartmentName := range compartmentInfo {
		compartmentNames = append(compartmentNames, compartmentName)
	}
	sort.Strings(compartmentNames)

	for _, compartmentName := range compartmentNames {
		tbl.AddRow(compartmentName, compartmentInfo[compartmentName])
	}

	tbl.Print()

	fmt.Println("\nTo set compartment, export OCI_COMPARTMENT_NAME:")
	yellow.Println("   export OCI_COMPARTMENT_NAME=")
}

func fetchVnicAttachments(client core.ComputeClient, compartmentId string) map[string]string {
	attachments := make(map[string]string)

	initialResponse, err := client.ListVnicAttachments(context.Background(), core.ListVnicAttachmentsRequest{CompartmentId: &compartmentId})
	checkError(err)

	for _, attachment := range initialResponse.Items {
		attachments[*attachment.InstanceId] = *attachment.VnicId
	}

	if initialResponse.OpcNextPage != nil {
		nextPage := initialResponse.OpcNextPage
		for {
			response, err := client.ListVnicAttachments(context.Background(), core.ListVnicAttachmentsRequest{CompartmentId: &compartmentId, Page: nextPage})
			checkError(err)

			for _, attachment := range response.Items {
				attachments[*attachment.InstanceId] = *attachment.VnicId
			}

			if response.OpcNextPage != nil {
				nextPage = response.OpcNextPage
			} else {
				break
			}
		}
	}

	return attachments
}

func fetchImage(computeClient core.ComputeClient, imageId string) Image {
	var image Image

	response, err := computeClient.GetImage(context.Background(), core.GetImageRequest{ImageId: &imageId})
	checkError(err)

	image = Image{
		*response.DisplayName,
		*response.Id,
		*response.TimeCreated,
		response.FreeformTags,
		response.DefinedTags,
		response.LaunchMode,
	}

	return image
}

func fetchImages(computeClient core.ComputeClient, compartmentId string) []Image {
	var images []Image
	var pageCount int
	pageCount = 0

	// Todo: pages
	fmt.Println(compartmentId)
	fmt.Println(pageCount)

	initialResponse, err := computeClient.ListImages(context.Background(), core.ListImagesRequest{CompartmentId: &compartmentId})
	checkError(err)

	for _, item := range initialResponse.Items {
		pageCount += 1
		fmt.Println(pageCount)
		// if item.LaunchMode == core.ImageLaunchModeCustom {
		image := Image{
			*item.DisplayName,
			*item.Id,
			*item.TimeCreated,
			item.FreeformTags,
			item.DefinedTags,
			item.LaunchMode,
		}

		images = append(images, image)
		// }
	}

	if initialResponse.OpcNextPage != nil {
		pageCount += 1
		fmt.Println(pageCount)

		nextPage := initialResponse.OpcNextPage
		for {
			response, err := computeClient.ListImages(context.Background(), core.ListImagesRequest{CompartmentId: &compartmentId, Page: nextPage})
			checkError(err)

			for _, item := range initialResponse.Items {
				// if item.LaunchMode == core.ImageLaunchModeCustom {
				image := Image{
					*item.DisplayName,
					*item.Id,
					*item.TimeCreated,
					item.FreeformTags,
					item.DefinedTags,
					item.LaunchMode,
				}

				images = append(images, image)
				// }
			}

			if response.OpcNextPage != nil {
				nextPage = response.OpcNextPage
			} else {
				break
			}
		}
	}

	fmt.Println(strconv.Itoa(pageCount) + "pages")
	return images
}

func listImages(computeClient core.ComputeClient, compartmentId string) {
	images := fetchImages(computeClient, compartmentId)

	for _, image := range images {
		fmt.Print("Name: ")
		blue.Println(image.name)

		fmt.Print("ID: ")
		yellow.Println(image.id)

		fmt.Print("Create date: ")
		yellow.Println(image.cDate)

		fmt.Println("Tags: ")

		for k, v := range image.freeTags {
			yellow.Println(k + ": " + v)
		}

		fmt.Print("Launch mode: ")
		yellow.Println(image.launchMode)

		fmt.Println("")
	}

	fmt.Println(strconv.Itoa(len(images)) + " images found")
}

func fetchInstances(computeClient core.ComputeClient, compartmentId string) []Instance {
	var instances []Instance

	initialResponse, err := computeClient.ListInstances(context.Background(), core.ListInstancesRequest{
		CompartmentId:  &compartmentId,
		LifecycleState: core.InstanceLifecycleStateRunning,
	})
	checkError(err)

	for _, instance := range initialResponse.Items {
		instance := Instance{
			*instance.DisplayName,
			*instance.Id,
			"0", // We have to lookup the private IP address separately
			*instance.AvailabilityDomain,
			*instance.Shape,
			*instance.TimeCreated,
			*instance.ImageId,
			*instance.FaultDomain,
			*instance.ShapeConfig.Vcpus,
			*instance.ShapeConfig.MemoryInGBs,
			*instance.Region,
			instance.LifecycleState,
		}
		instances = append(instances, instance)
	}

	if initialResponse.OpcNextPage != nil {
		nextPage := initialResponse.OpcNextPage
		for {
			response, err := computeClient.ListInstances(context.Background(), core.ListInstancesRequest{
				CompartmentId:  &compartmentId,
				LifecycleState: core.InstanceLifecycleStateRunning,
				Page:           nextPage,
			})
			checkError(err)

			for _, instance := range response.Items {
				instance := Instance{
					*instance.DisplayName,
					*instance.Id,
					"",
					*instance.AvailabilityDomain,
					*instance.Shape,
					*instance.TimeCreated,
					*instance.ImageId,
					*instance.FaultDomain,
					*instance.ShapeConfig.Vcpus,
					*instance.ShapeConfig.MemoryInGBs,
					*instance.Region,
					instance.LifecycleState,
				}
				instances = append(instances, instance)
			}

			if response.OpcNextPage != nil {
				nextPage = response.OpcNextPage
			} else {
				break
			}
		}
	}

	if logLevel == "DEBUG" {
		fmt.Println("")
		for _, name := range instances {
			fmt.Println(name)
		}
		fmt.Println("")
	}

	return instances
}

func matchInstances(pattern string, instances []Instance) []Instance {
	var matches []Instance

	// Handle simple wildcard
	if pattern == "*" {
		pattern = ".*"
	}

	for _, instance := range instances {
		match, _ := regexp.MatchString(pattern, instance.name)
		if match {
			matches = append(matches, instance)
		}
	}

	if logLevel == "DEBUG" {
		fmt.Println("\nMatches")
		for _, instance := range matches {
			fmt.Println(instance)
		}
	}

	return matches
}

func findInstances(computeClient core.ComputeClient, vnetClient core.VirtualNetworkClient, compartmentId string, flagSearchString string, retrieveImageInfo bool) {
	pattern := flagSearchString

	// When more than ~25 private IPs need to be looked up, its faster to batch them all together
	ipFetchAllThreshold := 25

	// Get relevant info for ALL instances
	// We have to do this because GetInstanceRequest/ListInstancesRequests do not allow filtering by pattern
	instances := fetchInstances(computeClient, compartmentId)
	// returns []Instance

	// Search all instances and return instances that match by name
	instanceMatches := matchInstances(pattern, instances)

	var batchFetchAllIps bool
	matchCount := len(instanceMatches)
	faint.Println(strconv.Itoa(matchCount) + " matches")

	if matchCount > ipFetchAllThreshold {
		batchFetchAllIps = true

		if logLevel == "DEBUG" {
			fmt.Print(matchCount)
			fmt.Println(" matches, Fetching all IPs")
		}
	} else {
		batchFetchAllIps = false

		if logLevel == "DEBUG" {
			fmt.Print(matchCount)
			fmt.Println(" matches, fetching IPs per instances")
		}
	}

	// Get ALL VNIC attachments
	// Once again, doing this because the request does not support filtering in the request
	attachments := fetchVnicAttachments(computeClient, compartmentId)
	// returns map of instanceId: vnicId

	vNicIdsToIps := make(map[string]string)
	if batchFetchAllIps {
		vNicIdsToIps = fetchPrivateIps(vnetClient, compartmentId) // This is inefficient when instance search results are small, resort to fetchPrivateIp
		// returns map of vnicId:privateIp
	}

	var instancesWithIP []Instance
	var privateIp string

	for _, instance := range instanceMatches {
		vnicId, ok := attachments[instance.id]
		if ok {
			if logLevel == "DEBUG" {
				fmt.Println("VNic ID: " + vnicId)
			}

			if batchFetchAllIps {
				privateIp = vNicIdsToIps[vnicId]
			} else {
				privateIp = fetchPrivateIp(vnetClient, vnicId)
			}

			instance.ip = privateIp
			instancesWithIP = append(instancesWithIP, instance) // TODO: Im sure theres a better way to do this using a single slice

		} else {
			fmt.Println("Unable to lookup VNIC for " + instance.id)
		}
	}

	if len(instancesWithIP) > 0 {
		sort.Sort(instancesByName(instancesWithIP))

		for _, instance := range instancesWithIP {
			region := instance.region
			ad := instance.ad
			strToRemove := "bKwM:" + region + "-" // TODO: This pattern needs to be updated. bKwM is not the universal prefix
			ad_short := strings.Replace(ad, strToRemove, "", -1)

			fd := instance.fd
			fd_short := strings.Replace(fd, "FAULT-DOMAIN", "FD", -1)

			fmt.Print("Name: ")
			blue.Println(instance.name)

			fmt.Print("ID: ")
			yellow.Println(instance.id)

			fmt.Print("Private IP: ")
			yellow.Print(instance.ip)

			fmt.Print(" FD: ")
			yellow.Print(fd_short)

			fmt.Print(" AD: ")
			yellow.Println(ad_short)

			fmt.Print("Shape: ")
			yellow.Print(instance.shape)

			fmt.Print(" Mem: ")
			yellow.Print(instance.mem)

			fmt.Print(" vCPUs: ")
			yellow.Println(instance.vCPUs)

			fmt.Print("State: ")
			yellow.Println(instance.state)

			fmt.Print("Created: ")
			yellow.Println(instance.cDate)

			if retrieveImageInfo {
				image := fetchImage(computeClient, instance.imageId) // TODO: Performance hit: this adds ~100 ms per image lookup

				fmt.Print("Image Name: ")
				yellow.Println(image.name)

				fmt.Print("Image ID: ")
				yellow.Println(instance.imageId)

				fmt.Println("Image Tags (Free form): ")

				freeformTagKeys := make([]string, 0, len(image.freeTags))
				for key := range image.freeTags {
					freeformTagKeys = append(freeformTagKeys, key)
				}
				sort.Strings(freeformTagKeys)

				faint.Print("| ")
				for _, key := range freeformTagKeys {
					faint.Print(key + ": " + image.freeTags[key] + " | ")
				}

				fmt.Println("")

				fmt.Println("Image Tags (Defined): ")
				for tagNs, tags := range image.definedTags {
					faintUnder.Println(tagNs)

					definedTagKeys := make([]string, 0, len(tags))
					for key := range tags {
						definedTagKeys = append(definedTagKeys, key)
					}
					sort.Strings(definedTagKeys)

					faint.Print("| ")
					for _, key := range definedTagKeys {
						faint.Print(key + ": " + tags[key].(string) + " | ")
					}

					fmt.Println("")

				}
			}

			fmt.Println("")
		}
	}
}

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

	if logLevel == "DEBUG" {
		fmt.Println("\nMatches")
		for cluster := range matches {
			fmt.Println(cluster)
		}
	}

	return matches
}

func findClusters(containerEngineClient containerengine.ContainerEngineClient, compartmentId string, flagSearchString string) {
	pattern := flagSearchString
	clusters := fetchClusters(containerEngineClient, compartmentId)
	clusterMatches := matchClusters(pattern, clusters)

	if len(clusterMatches) > 0 {
		yellow.Println("OKE Clusters")
		faint.Println(strconv.Itoa(len(clusterMatches)) + " matches")

		for _, cluster := range clusterMatches {
			fmt.Print("Name: ")
			blue.Println(cluster.name)
			fmt.Print("Cluster ID: ")
			yellow.Println(cluster.id)
			fmt.Print("Private endpoint: ")
			yellow.Println(cluster.privateEndpointIp + ":" + cluster.privateEndpointPort)
			fmt.Println("")
		}
	}
}

func listInstances(computeClient core.ComputeClient, compartmentId string, vnetClient core.VirtualNetworkClient) {
	instances := fetchInstances(computeClient, compartmentId)
	// returns []Instance

	for _, instance := range instances {
		region := instance.region
		ad := instance.ad
		strToRemove := "bKwM:" + region + "-"
		ad_short := strings.Replace(ad, strToRemove, "", -1)

		fd := instance.fd
		fd_short := strings.Replace(fd, "FAULT-DOMAIN", "FD", -1)

		fmt.Print("Name: ")
		blue.Println(instance.name)

		fmt.Print("ID: ")
		yellow.Println(instance.id)

		fmt.Print("Private IP: ")
		yellow.Print(instance.ip)

		fmt.Print(" FD: ")
		yellow.Print(fd_short)

		fmt.Print(" AD: ")
		yellow.Println(ad_short)

		fmt.Print("Shape: ")
		yellow.Print(instance.shape)

		fmt.Print(" Mem: ")
		yellow.Print(instance.mem)

		fmt.Print(" vCPUs: ")
		yellow.Println(instance.vCPUs)

		fmt.Print("State: ")
		yellow.Println(instance.state)

		fmt.Print("Created: ")
		yellow.Println(instance.cDate)

		fmt.Print("Image ID: ")
		yellow.Println(instance.imageId)

		fmt.Println("")
	}
}

func fetchClusters(containerEngineClient containerengine.ContainerEngineClient, compartmentId string) []Cluster {
	var Clusters []Cluster

	initialResponse, err := containerEngineClient.ListClusters(context.Background(), containerengine.ListClustersRequest{
		CompartmentId: &compartmentId,
		// LifecycleState: core.InstanceLifecycleStateRunning,
	})
	checkError(err)

	for _, cluster := range initialResponse.Items {
		// fmt.Println(cluster)/

		clusterId := *cluster.Id
		clusterName := *cluster.Name
		// clusterPrivateEndpoint := *cluster.Endpoints.PrivateEndpoint

		clusterPrivateEndpointIp, clusterPrivateEndpointPort, found := strings.Cut(*cluster.Endpoints.PrivateEndpoint, ":")
		// clusterPrivateEndpointPort := *cluster.Endpoints.PrivateEndpoint

		if found {
			cluster := Cluster{clusterName, clusterId, clusterPrivateEndpointIp, clusterPrivateEndpointPort}
			Clusters = append(Clusters, cluster)
		}
	}

	return Clusters
}

func fetchSubnetIds(client core.VirtualNetworkClient, compartmentId string) []string {
	response, err := client.ListSubnets(context.Background(), core.ListSubnetsRequest{CompartmentId: &compartmentId})
	checkError(err)

	var subnetIds []string

	for _, subnet := range response.Items {
		subnetIds = append(subnetIds, *subnet.Id)
	}

	return subnetIds
}

func listSubnets(client core.VirtualNetworkClient, compartmentId string) {
	response, err := client.ListSubnets(context.Background(), core.ListSubnetsRequest{CompartmentId: &compartmentId})
	checkError(err)

	var Subnets []Subnet
	var subnetAccess string
	var subnetType string

	for _, s := range response.Items {
		if *s.ProhibitInternetIngress && *s.ProhibitPublicIpOnVnic {
			subnetAccess = "private"
		} else if !*s.ProhibitInternetIngress && !*s.ProhibitPublicIpOnVnic {
			subnetAccess = "public"
		} else {
			subnetAccess = "?"
		}

		if s.AvailabilityDomain == nil {
			subnetType = "Regional"
		} else {
			subnetType = *s.AvailabilityDomain
		}

		subnet := Subnet{*s.CidrBlock, *s.DisplayName, subnetAccess, subnetType}
		Subnets = append(Subnets, subnet)
	}

	if len(Subnets) > 0 {
		sort.Sort(subnetsByCidr(Subnets))
	}

	tbl := table.New("CIDR", "Name", "Access", "Type")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, subnet := range Subnets {
		tbl.AddRow(subnet.cidr, subnet.name, subnet.access, subnet.subnetType)
	}

	tbl.Print()
}

func fetchPrivateIp(client core.VirtualNetworkClient, vnicId string) string {
	response, err := client.GetVnic(context.Background(), core.GetVnicRequest{VnicId: &vnicId})
	checkError(err)

	return *response.Vnic.PrivateIp
}

func fetchPrivateIps(client core.VirtualNetworkClient, compartmentId string) map[string]string {
	vNicIdsToIps := make(map[string]string)
	subnetIds := fetchSubnetIds(client, compartmentId)

	for _, subnetId := range subnetIds {
		response, err := client.ListPrivateIps(context.Background(), core.ListPrivateIpsRequest{SubnetId: &subnetId})
		checkError(err)

		for _, item := range response.Items {
			vNicIdsToIps[*item.VnicId] = *item.IpAddress
		}
	}

	return vNicIdsToIps
}

func checkBastionName(flagBastionName string) string {
	var bastionName string
	bastionNameEnv, exists := os.LookupEnv("OCI_BASTION_NAME")
	if exists {
		bastionName = bastionNameEnv
		if logLevel == "DEBUG" {
			fmt.Println("Bastion name is set via OCI_BASTION_NAME to: " + bastionName)
		}
	} else if flagBastionName == "" {
		fmt.Println("Must pass bastion name with -b or set with environment variable OCI_BASTION_NAME")
		os.Exit(1)
	} else {
		bastionName = flagBastionName
	}

	return bastionName
}

func fetchBastions(compartmentId string, client bastion.BastionClient) map[string]string {
	response, err := client.ListBastions(context.Background(), bastion.ListBastionsRequest{CompartmentId: &compartmentId})
	checkError(err)

	bastionInfo := make(map[string]string)

	for _, item := range response.Items {
		bastionInfo[*item.Name] = *item.Id
	}

	return bastionInfo
}

func fetchBastion(bastionName string, bastionId string, client bastion.BastionClient) {
	if logLevel == "DEBUG" {
		fmt.Println("\nGetting bastion for: " + bastionName + " (" + bastionId + ")")
	}

	_, err := client.GetBastion(context.Background(), bastion.GetBastionRequest{BastionId: &bastionId})
	checkError(err)
}

func listBastions(bastionInfo map[string]string) {
	tbl := table.New("Bastion Name", "OCID")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for bastionName := range bastionInfo {
		// fmt.Print(bastionName)
		// faint.Println(" " + bastionInfo[bastionName])
		tbl.AddRow(bastionName, bastionInfo[bastionName])
	}

	tbl.Print()

	fmt.Println("\nTo set bastion name, export OCI_BASTION_NAME:")
	yellow.Println("   export OCI_BASTION_NAME=")
}

func readSshPubKey(sshPrivateKeyFileLocation string) string {
	if sshPrivateKeyFileLocation == "" {
		sshPrivateKeyFileLocation = homeDir() + "/.ssh/id_rsa"
		fmt.Println("\nUsing default SSH identity file at " + sshPrivateKeyFileLocation)
	}

	sshKeyContents, err := os.ReadFile(sshPrivateKeyFileLocation)
	checkError(err)

	return string(sshKeyContents)
}

func createManagedSshSession(bastionId string, client bastion.BastionClient, targetInstance string, targetIp string, publicKeyContent string, sshUser string, sshPort int, sessionTtl int) *string {
	req := bastion.CreateSessionRequest{
		CreateSessionDetails: bastion.CreateSessionDetails{
			BastionId:           &bastionId,
			DisplayName:         common.String("oshivSession"),
			KeyDetails:          &bastion.PublicKeyDetails{PublicKeyContent: &publicKeyContent},
			SessionTtlInSeconds: common.Int(sessionTtl),
			TargetResourceDetails: bastion.CreateManagedSshSessionTargetResourceDetails{
				TargetResourceId:                      &targetInstance,
				TargetResourceOperatingSystemUserName: &sshUser,
				TargetResourcePort:                    &sshPort,
				TargetResourcePrivateIpAddress:        &targetIp,
			},
		},
	}

	fmt.Println("Creating managed SSH session...")
	response, err := client.CreateSession(context.Background(), req)
	checkError(err)

	if logLevel == "DEBUG" {
		fmt.Println("\nCreateSessionResponse")
		fmt.Println(response)
	}

	sessionId := response.Session.Id
	blue.Println("\nSession ID")
	fmt.Println(*sessionId)
	fmt.Println("")

	return sessionId
}

func createPortFwSession(bastionId string, client bastion.BastionClient, targetIp string, publicKeyContent string, sshTunnelPort int, sessionTtl int) *string {
	req := bastion.CreateSessionRequest{
		CreateSessionDetails: bastion.CreateSessionDetails{
			BastionId:           &bastionId,
			DisplayName:         common.String("oshivSession"),
			KeyDetails:          &bastion.PublicKeyDetails{PublicKeyContent: &publicKeyContent},
			SessionTtlInSeconds: common.Int(sessionTtl),
			TargetResourceDetails: bastion.PortForwardingSessionTargetResourceDetails{
				TargetResourcePort:             &sshTunnelPort,
				TargetResourcePrivateIpAddress: &targetIp,
			},
		},
	}

	fmt.Println("Creating port forwarding session...")
	response, err := client.CreateSession(context.Background(), req)
	checkError(err)

	if logLevel == "DEBUG" {
		fmt.Println("\nCreateSessionResponse")
		fmt.Println(response)
	}

	sessionId := response.Session.Id
	blue.Println("\nSession ID")
	fmt.Println(*sessionId)
	fmt.Println("")

	return sessionId
}

func checkSession(client bastion.BastionClient, sessionId *string, flagCreatePortFwSession bool) SessionInfo {
	response, err := client.GetSession(context.Background(), bastion.GetSessionRequest{SessionId: sessionId})
	checkError(err)

	if logLevel == "DEBUG" {
		fmt.Println("GetSessionResponse")
		fmt.Println(response.Session)

		fmt.Println("\nEndpoint")
		fmt.Println(client.Endpoint())
	}

	var ipAddress *string
	var sshUser *string
	var sshPort *int

	if flagCreatePortFwSession {
		sshSessionTargetResourceDetails := response.Session.TargetResourceDetails.(bastion.PortForwardingSessionTargetResourceDetails)
		ipAddress = sshSessionTargetResourceDetails.TargetResourcePrivateIpAddress
		sshPort = sshSessionTargetResourceDetails.TargetResourcePort
	} else {
		sshSessionTargetResourceDetails := response.Session.TargetResourceDetails.(bastion.ManagedSshSessionTargetResourceDetails)
		ipAddress = sshSessionTargetResourceDetails.TargetResourcePrivateIpAddress
		sshUser = sshSessionTargetResourceDetails.TargetResourceOperatingSystemUserName
		sshPort = sshSessionTargetResourceDetails.TargetResourcePort
	}

	var currentSessionInfo SessionInfo

	if flagCreatePortFwSession {
		currentSessionInfo = SessionInfo{response.Session.LifecycleState, *ipAddress, "", *sshPort}
	} else {
		currentSessionInfo = SessionInfo{response.Session.LifecycleState, *ipAddress, *sshUser, *sshPort}
	}

	return currentSessionInfo
}

func listActiveSessions(client bastion.BastionClient, bastionId string) {
	response, err := client.ListSessions(context.Background(), bastion.ListSessionsRequest{BastionId: &bastionId})
	checkError(err)

	blue.Println("Active bastion sessions")
	for _, session := range response.Items {
		sshSessionTargetResourceDetails := session.TargetResourceDetails.(bastion.ManagedSshSessionTargetResourceDetails)
		instanceName := sshSessionTargetResourceDetails.TargetResourceDisplayName
		ipAddress := sshSessionTargetResourceDetails.TargetResourcePrivateIpAddress
		instanceID := sshSessionTargetResourceDetails.TargetResourceId

		if session.LifecycleState == "ACTIVE" {
			fmt.Println(*session.DisplayName)
			fmt.Println(*session.Id)
			fmt.Println(*session.TimeCreated)
			fmt.Println(*instanceName)
			fmt.Println(*ipAddress)
			fmt.Println(*instanceID)
			fmt.Println("")
		}
	}
}

func printSshCommands(client bastion.BastionClient, sessionId *string, instanceIp *string, sshUser *string, sshPort *int, sshIdentityFile string, tunnelPort int, localPort *int) {
	bastionEndpointUrl, err := url.Parse(client.Endpoint())
	checkError(err)

	sessionIdStr := *sessionId
	bastionHost := sessionIdStr + "@host." + bastionEndpointUrl.Host

	// TODO: Consider proxy jump flag for commands where applicable - https://www.ateam-oracle.com/post/openssh-proxyjump-with-oci-bastion-service
	if tunnelPort == 0 {
		yellow.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p -p 22 " + bastionHost + "' \\")
		fmt.Println("-P " + strconv.Itoa(*sshPort) + " " + *sshUser + "@" + *instanceIp + " -N -L " + color.RedString("LOCAL_PORT") + ":" + *instanceIp + ":" + color.RedString("REMOTE_PORT"))
	} else if *localPort != 0 {
		yellow.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p -p 22 " + bastionHost + "' \\")
		fmt.Println("-P " + strconv.Itoa(*sshPort) + " " + *sshUser + "@" + *instanceIp + " -N -L " + strconv.Itoa(*localPort) + ":" + *instanceIp + ":" + strconv.Itoa(tunnelPort))
	} else {
		yellowBold.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p -p 22 " + bastionHost + "' \\")
		fmt.Println("-P " + strconv.Itoa(*sshPort) + " " + *sshUser + "@" + *instanceIp + " -N -L " + strconv.Itoa(tunnelPort) + ":" + *instanceIp + ":" + strconv.Itoa(tunnelPort))
	}

	yellow.Println("\nSCP command")
	fmt.Println("scp -i " + sshIdentityFile + " -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P " + strconv.Itoa(*sshPort) + " \\")
	fmt.Println("-o ProxyCommand='ssh -i " + sshIdentityFile + " -W %h:%p -p 22 " + bastionHost + "' \\")
	fmt.Println(color.RedString("SOURCE_PATH ") + *sshUser + "@" + *instanceIp + ":" + color.RedString("TARGET_PATH"))

	yellow.Println("\nSSH comand")
	fmt.Println("ssh -i " + sshIdentityFile + " -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
	fmt.Println("-o ProxyCommand='ssh -i " + sshIdentityFile + " -W %h:%p -p 22 " + bastionHost + "' \\")
	fmt.Println("-P " + strconv.Itoa(*sshPort) + " " + *sshUser + "@" + *instanceIp)
}

func printPortFwSshCommands(client bastion.BastionClient, sessionId *string, targetIp *string, sshPort *int, sshIdentityFile string, tunnelPort int, localTunnelPort int, flagOkeClusterId *string) {
	bastionEndpointUrl, err := url.Parse(client.Endpoint())
	checkError(err)

	sessionIdStr := *sessionId
	bastionHost := sessionIdStr + "@host." + bastionEndpointUrl.Host

	if *flagOkeClusterId != "" {
		yellow.Println("\nUpdate kube config (One time operation)")
		fmt.Println("oci ce cluster create-kubeconfig --cluster-id " + *flagOkeClusterId + " --token-version 2.0.0 --kube-endpoint PRIVATE_ENDPOINT --auth security_token")
	}

	yellow.Println("\nPort Forwarding command")
	fmt.Println("ssh -i \"" + sshIdentityFile + "\" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")

	if tunnelPort == 0 {
		fmt.Println("-p " + strconv.Itoa(*sshPort) + " -N -L " + color.RedString("LOCAL_PORT") + ":" + *targetIp + ":" + color.RedString("REMOTE_PORT") + " " + bastionHost)
	} else if localTunnelPort != 0 {
		fmt.Println("-p " + strconv.Itoa(*sshPort) + " -N -L " + strconv.Itoa(localTunnelPort) + ":" + *targetIp + ":" + strconv.Itoa(tunnelPort) + " " + bastionHost)
	} else {
		fmt.Println("-p " + strconv.Itoa(*sshPort) + " -N -L " + strconv.Itoa(tunnelPort) + ":" + *targetIp + ":" + strconv.Itoa(tunnelPort) + " " + bastionHost)
	}
}

func main() {
	// Global flags
	flagVersion := flag.Bool("v", false, "Show version")
	flagTenancyId := flag.String("t", "", "tenancy ID name")

	flagListCompartments := flag.Bool("lc", false, "list compartments")
	flagListBastions := flag.Bool("lb", false, "list bastions")
	flagListSessions := flag.Bool("ls", false, "list sessions")

	flagSearchString := flag.String("f", "", "search string to search for instance")
	flagSearchStringWithImageInfo := flag.String("fi", "", "search string to search for instance and return image info")

	flagCompartmentName := flag.String("c", "", "compartment name")
	flagBastionName := flag.String("b", "", "bastion name")

	flagInstanceId := flag.String("o", "", "instance ID of host to connect to")
	flagOkeClusterId := flag.String("oke", "", "OKE cluster ID")
	flagTargetIp := flag.String("i", "", "IP address of host/endpoint to connect to")

	flagSshUser := flag.String("u", "opc", "SSH user")
	flagSshPort := flag.Int("p", 22, "SSH port")
	flagSshPrivateKey := flag.String("k", "", "path to SSH private key (identity file)")
	flagSshPublicKey := flag.String("e", "", "path to SSH public key")
	flagSessionTtl := flag.Int("l", 10800, "Session TTL (seconds)")

	flagSessionId := flag.String("s", "", "Session ID to check for")

	flagCreatePortFwSession := flag.Bool("fw", false, "Create an SSH port forward session")

	// For managed SSH sessions: tp will be used for both LOCAL and REMOTE port in tunnel command
	// For port forward SSH sessions: tp will be used for both LOCAL and REMOTE port in tunnel command and the session's target port
	flagSshTunnelPort := flag.Int("tp", 0, "SSH tunnel port") // TODO: consider breaking out tunnel port from port forwarding port

	// This will override the local port for both managed SSH and and port forward sessions
	flagSshTunnelPortOverrideLocal := flag.Int("tpl", 0, "SSH tunnel local port override")

	// Extend flag's default usage function
	flag.Usage = func() {
		fmt.Println("OCI authentication:")
		fmt.Println("This tool will use the credentials set in $HOME/.oci/config")
		fmt.Println("This tool will use the profile set by the OCI_CLI_PROFILE environment variable")
		fmt.Println("If the OCI_CLI_PROFILE environment variable is not set it will use the DEFAULT profile")

		fmt.Println("\nEnvironment variables:")
		fmt.Println("The following environment variables will override their flag counterparts")
		fmt.Println("   OCI_CLI_TENANCY")
		fmt.Println("   OCI_COMPARTMENT_NAME")
		fmt.Println("   OCI_BASTION_NAME")

		fmt.Println("\nDefaults:")
		fmt.Println("   SSH private key (-k): $HOME/.ssh/id_rsa")
		fmt.Println("   SSH public key (-e): $HOME/.ssh/id_rsa.pub")
		fmt.Println("   SSH user (-u): opc")

		fmt.Println("\nCommon command patterns:")

		fmt.Println("List compartments")
		fmt.Println("   oshiv -lc")
		fmt.Println("\nList bastions")
		fmt.Println("   oshiv -lb")
		fmt.Println("\nCreate bastion session")
		fmt.Println("   oshiv -i ip_address -o instance_id")
		fmt.Println("\nList active sessions")
		fmt.Println("   oshiv -ls")
		fmt.Println("\nConnect to an active session")
		fmt.Println("   oshiv -s session_ocd")
		fmt.Println("\nCreate bastion session (all flags)")
		fmt.Println("   oshiv -t tenant_id -c compartment_name -b bastion_name -i ip_address -o instance_id -k path_to_ssh_private_key -e path_to_ssh_public_key -u cloud-user")

		fmt.Fprintf(flag.CommandLine.Output(), "\nAll flags for %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Subcommands and flags
	computeCmd := flag.NewFlagSet("compute", flag.ExitOnError)
	flagComputeList := computeCmd.Bool("l", false, "List all instances")

	imageCmd := flag.NewFlagSet("image", flag.ExitOnError)
	flagImageList := imageCmd.Bool("l", false, "List all images")
	flagImageFind := imageCmd.String("f", "", "Find images by search pattern")

	subnetsCmd := flag.NewFlagSet("subnets", flag.ExitOnError)
	flagSubnetsList := subnetsCmd.Bool("l", false, "List all subnets")
	flagSubnetsFind := subnetsCmd.String("f", "", "Find subnets by search pattern")

	subcommand := parseSubcommand(os.Args[1])

	// Main program logic starts here
	// Print version
	if *flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// Configure OCI provider by profile
	ociConfig := configureProvider()

	// identityClient is always required
	identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
	checkError(identityErr)

	// Attempt to get tenancy ID from input and validate it against OCI API
	tenancyId := validateTenancyId(*flagTenancyId, identityClient)
	// All actions except listing compartments require a compartment ID, compartment info will contain a map of compartment names and IDs
	compartmentInfo := fetchCompartmentInfo(tenancyId, identityClient)

	// List all compartments
	if *flagListCompartments {
		listCompartmentNames(compartmentInfo)
		os.Exit(0)
	}

	// <-- Anything beyond this point requires a compartment -->
	// Attempt to get compartment name from input then lookup compartment ID
	compartmentName := checkCompartmentName(*flagCompartmentName)
	compartmentId := lookupCompartmentId(compartmentInfo, compartmentName)

	// Subcommands
	switch subcommand {
	case "compute":
		computeCmd.Parse(os.Args[2:])

		computeClient, err := core.NewComputeClientWithConfigurationProvider(ociConfig)
		checkError(err)

		vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(ociConfig)
		checkError(err)

		if *flagComputeList {
			listInstances(computeClient, compartmentId, vnetClient)
		}

		os.Exit(0)

	case "image":
		imageCmd.Parse(os.Args[2:])

		computeClient, err := core.NewComputeClientWithConfigurationProvider(ociConfig)
		checkError(err)

		if *flagImageList {
			listImages(computeClient, compartmentId)
		} else if *flagImageFind != "" {
			fmt.Println("Image search is not yet enabled, listing all images. Use grep!")
			listImages(computeClient, compartmentId)
		}
		os.Exit(0)

	case "subnets":
		subnetsCmd.Parse(os.Args[2:])

		vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(ociConfig)
		checkError(err)

		if *flagSubnetsList {
			listSubnets(vnetClient, compartmentId)
		} else if *flagSubnetsFind != "" {
			fmt.Println("Subnet search is not yet enabled, listing all subnets. Use grep!")
			listSubnets(vnetClient, compartmentId)
		}
		os.Exit(0)

	// If no subcommand is given, we're in legacy mode (instance/cluster search || list bastions || list sessions || create session || check session)
	case "":
		// We're in instance/cluster search mode
		if *flagSearchString != "" || *flagSearchStringWithImageInfo != "" {
			var retrieveImageInfo bool
			var searchString string

			if *flagSearchStringWithImageInfo != "" {
				retrieveImageInfo = true
				searchString = *flagSearchStringWithImageInfo
			} else {
				retrieveImageInfo = false
				searchString = *flagSearchString
			}

			computeClient, err := core.NewComputeClientWithConfigurationProvider(ociConfig)
			checkError(err)

			vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(ociConfig)
			checkError(err)

			// Default searches for both instances and clusters
			findInstances(computeClient, vnetClient, compartmentId, searchString, retrieveImageInfo)

			containerEngineClient, err := containerengine.NewContainerEngineClientWithConfigurationProvider(ociConfig)
			checkError(err)

			findClusters(containerEngineClient, compartmentId, searchString)

			os.Exit(0)
		}

		// We're in bastion mode
		bastionClient, err := bastion.NewBastionClientWithConfigurationProvider(ociConfig)
		checkError(err)

		// <-- Anything beyond this point requires bastion information -->
		bastionInfo := fetchBastions(compartmentId, bastionClient)

		if *flagListBastions {
			listBastions(bastionInfo)
			os.Exit(0)
		}

		// Anything past this point requires a bastion
		var bastionName string
		var bastionId string

		// If there is only one bastion, no need to require bastion name as input
		if len(bastionInfo) == 1 {
			for name, id := range bastionInfo {
				bastionName = name
				bastionId = id
			}

			if logLevel == "DEBUG" {
				fmt.Println("Only one bastion found, using it")
				fmt.Println(bastionName + " (" + bastionId + ")")
			}
		} else {
			// There were multiple bastions found so we'll need to know which one to use
			if logLevel == "DEBUG" {
				fmt.Println("multiple bastions found, checking to see if one has been specified...")
				fmt.Println(bastionName + " (" + bastionId + ")")
			}

			bastionName = checkBastionName(*flagBastionName)
			bastionId = bastionInfo[bastionName]
		}

		fetchBastion(bastionName, bastionId, bastionClient)

		if *flagListSessions {
			listActiveSessions(bastionClient, bastionId)
			os.Exit(0)
		}

		var sshPrivateKeyFileLocation string
		if *flagSshPrivateKey == "" {
			// TODO: move this default to flags
			sshPrivateKeyFileLocation = homeDir() + "/.ssh/id_rsa"
			if logLevel == "DEBUG" {
				fmt.Println("Using default SSH private key file at " + sshPrivateKeyFileLocation)
			}
		} else {
			sshPrivateKeyFileLocation = *flagSshPrivateKey
		}

		var sshPublicKeyFileLocation string
		if *flagSshPublicKey == "" {
			// TODO: move this default to flags
			sshPublicKeyFileLocation = homeDir() + "/.ssh/id_rsa.pub"
			if logLevel == "DEBUG" {
				fmt.Println("\nUsing default SSH public key file at " + sshPublicKeyFileLocation)
			}
		} else {
			sshPublicKeyFileLocation = *flagSshPublicKey
		}

		publicKeyContent := readSshPubKey(sshPublicKeyFileLocation)

		var sshTunnelPort int
		if *flagOkeClusterId != "" {
			sshTunnelPort = 6443
		} else {
			sshTunnelPort = *flagSshTunnelPort
		}

		// Create bastion sessions
		var sessionId *string
		if *flagSessionId == "" {
			// No session ID passed, create a new session
			if *flagCreatePortFwSession || *flagOkeClusterId != "" {

				sessionId = createPortFwSession(bastionId, bastionClient, *flagTargetIp, publicKeyContent, sshTunnelPort, *flagSessionTtl)
			} else {
				sessionId = createManagedSshSession(bastionId, bastionClient, *flagInstanceId, *flagTargetIp, publicKeyContent, *flagSshUser, *flagSshPort, *flagSessionTtl)
			}
		} else {
			// Check for existing session by session ID
			fmt.Println("Session ID passed, checking session...")
			sessionId = flagSessionId
			sessionInfo := checkSession(bastionClient, sessionId, *flagCreatePortFwSession || *flagOkeClusterId != "")

			if sessionInfo.state == "ACTIVE" {
				if *flagCreatePortFwSession {
					printPortFwSshCommands(bastionClient, sessionId, &sessionInfo.ip, &sessionInfo.port, sshPrivateKeyFileLocation, sshTunnelPort, *flagSshTunnelPortOverrideLocal, flagOkeClusterId)
				} else {
					printSshCommands(bastionClient, sessionId, &sessionInfo.ip, &sessionInfo.user, &sessionInfo.port, sshPrivateKeyFileLocation, sshTunnelPort, flagSshTunnelPortOverrideLocal)
				}
			} else {
				fmt.Println("Session is no longer active. Current state is: " + sessionInfo.state)
			}

			os.Exit(0)
		}

		sessionInfo := checkSession(bastionClient, sessionId, *flagCreatePortFwSession || *flagOkeClusterId != "")

		for sessionInfo.state != "ACTIVE" {
			if sessionInfo.state == "DELETED" {
				fmt.Println("\nSession has been deleted, exiting")
				fmt.Println("State: " + sessionInfo.state)
				fmt.Println("\nSession Info")
				fmt.Println(sessionInfo)
				os.Exit(1)
			} else {
				fmt.Println("Session not yet active, waiting... (State: " + sessionInfo.state + ")")
				time.Sleep(10 * time.Second)
				sessionInfo = checkSession(bastionClient, sessionId, *flagCreatePortFwSession || *flagOkeClusterId != "")
			}
		}

		if *flagCreatePortFwSession || *flagOkeClusterId != "" {
			printPortFwSshCommands(bastionClient, sessionId, flagTargetIp, flagSshPort, sshPrivateKeyFileLocation, sshTunnelPort, *flagSshTunnelPortOverrideLocal, flagOkeClusterId)
		} else {
			printSshCommands(bastionClient, sessionId, flagTargetIp, flagSshUser, flagSshPort, sshPrivateKeyFileLocation, sshTunnelPort, flagSshTunnelPortOverrideLocal)
		}
	}
}

// TODO: Currently all networking functions include all VCNs without indicating which VNC an object belongs to. Need to support VNC ID flag and print VNC when flag not passed
