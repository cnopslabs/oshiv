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
)

const logLevel = "INFO"   // TODO: switch to logging library
var version = "undefined" // Version gets automatically updated during build
var boldBlue = color.New(color.FgCyan, color.Bold)
var boldYellow = color.New(color.FgYellow, color.Bold)
var yellow = color.New(color.FgYellow)

type SessionInfo struct {
	state bastion.SessionLifecycleStateEnum
	ip    string
	user  string
	port  int
}

type Instance struct {
	name string
	id   string
	ip   string
}

type Cluster struct {
	name                string
	id                  string
	privateEndpointIp   string
	privateEndpointPort string
}

type instancesByName []Instance

func (m instancesByName) Len() int           { return len(m) }
func (m instancesByName) Less(i, j int) bool { return m[i].name < m[j].name }
func (m instancesByName) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	checkError(err)

	return homeDir
}

func initializeOciClients() (identity.IdentityClient, bastion.BastionClient, core.ComputeClient, core.VirtualNetworkClient, containerengine.ContainerEngineClient) {
	var config common.ConfigurationProvider

	profile, exists := os.LookupEnv("OCI_CLI_PROFILE")

	if exists {
		if logLevel == "DEBUG" {
			fmt.Println("Using profile " + profile)
		}

		homeDir := getHomeDir()
		configPath := homeDir + "/.oci/config"

		config = common.CustomProfileConfigProvider(configPath, profile)
	} else {
		if logLevel == "DEBUG" {
			fmt.Println("Using default profile")
		}
		config = common.DefaultConfigProvider()
	}

	identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(config)
	checkError(identityErr)

	computeClient, err := core.NewComputeClientWithConfigurationProvider(config)
	checkError(err)

	vnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(config)
	checkError(err)

	bastionClient, err := bastion.NewBastionClientWithConfigurationProvider(config)
	checkError(err)

	containerEngineClient, err := containerengine.NewContainerEngineClientWithConfigurationProvider(config)
	checkError(err)

	return identityClient, bastionClient, computeClient, vnetClient, containerEngineClient
}

func getTenancyId(tenancyIdFlag string, client identity.IdentityClient) string {
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

func getCompartmentInfo(tenancyId string, client identity.IdentityClient) map[string]string {
	response, err := client.ListCompartments(context.Background(), identity.ListCompartmentsRequest{CompartmentId: &tenancyId})
	checkError(err)

	compartmentInfo := make(map[string]string)

	for _, item := range response.Items {
		compartmentInfo[*item.Name] = *item.Id
	}

	return compartmentInfo
}

func listCompartmentNames(compartmentInfo map[string]string) {
	boldBlue.Println("Compartments")

	compartmentNames := make([]string, 0, len(compartmentInfo))
	for compartmentName := range compartmentInfo {
		compartmentNames = append(compartmentNames, compartmentName)
	}
	sort.Strings(compartmentNames)

	for _, compartmentName := range compartmentNames {
		println(compartmentName)
	}

	fmt.Println("\nTo set compartment, you can export OCI_COMPARTMENT_NAME:")
	yellow.Println("   export OCI_COMPARTMENT_NAME=")
}

func getInstances(client core.ComputeClient, compartmentId string) map[string]string {
	instances := make(map[string]string)

	initialResponse, err := client.ListInstances(context.Background(), core.ListInstancesRequest{
		CompartmentId:  &compartmentId,
		LifecycleState: core.InstanceLifecycleStateRunning,
	})
	checkError(err)

	for _, instance := range initialResponse.Items {
		instances[*instance.Id] = *instance.DisplayName
	}

	if initialResponse.OpcNextPage != nil {
		nextPage := initialResponse.OpcNextPage
		for {
			response, err := client.ListInstances(context.Background(), core.ListInstancesRequest{
				CompartmentId:  &compartmentId,
				LifecycleState: core.InstanceLifecycleStateRunning,
				Page:           nextPage,
			})
			checkError(err)

			for _, instance := range response.Items {
				instances[*instance.Id] = *instance.DisplayName
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

func getClusters(containerEngineClient containerengine.ContainerEngineClient, compartmentId string) []Cluster {
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

func searchClusters(pattern string, clusters []Cluster) []Cluster {
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

func searchInstances(pattern string, instances map[string]string) map[string]string {
	matches := make(map[string]string)

	// Handle simple wildcard
	if pattern == "*" {
		pattern = ".*"
	}

	for instanceId, instanceName := range instances {
		match, _ := regexp.MatchString(pattern, instanceName)
		if match {
			matches[instanceName] = instanceId
		}
	}

	if logLevel == "DEBUG" {
		fmt.Println("\nMatches")
		for instanceName := range matches {
			fmt.Println(instanceName)
		}
	}

	return matches
}

func getVnicAttachments(client core.ComputeClient, compartmentId string) map[string]string {
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

func getPrivateIp(client core.VirtualNetworkClient, vnicId string) string {
	response, err := client.GetVnic(context.Background(), core.GetVnicRequest{VnicId: &vnicId})
	checkError(err)

	return *response.Vnic.PrivateIp
}

func getCompartmentName(flagCompartmentName string) string {
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

func getBastionName(flagBastionName string) string {
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

func getCompartmentId(compartmentInfo map[string]string, compartmentName string) string {
	compartmentId := compartmentInfo[compartmentName]
	if logLevel == "DEBUG" {
		fmt.Println("\n" + compartmentName + "'s compartment ID is " + compartmentId)
	}

	return compartmentId
}

func getBastionInfo(compartmentId string, client bastion.BastionClient) map[string]string {
	response, err := client.ListBastions(context.Background(), bastion.ListBastionsRequest{CompartmentId: &compartmentId})
	checkError(err)

	bastionInfo := make(map[string]string)

	for _, item := range response.Items {
		bastionInfo[*item.Name] = *item.Id
	}

	return bastionInfo
}

func listBastions(compartmentName string, bastionInfo map[string]string) {
	boldBlue.Println("Bastions in compartment " + compartmentName)
	for bastionName := range bastionInfo {
		fmt.Println(bastionName)
	}

	fmt.Println("\nTo set bastion name, you can export OCI_BASTION_NAME:")
	yellow.Println("   export OCI_BASTION_NAME=")
}

func getBastion(bastionName string, bastionId string, client bastion.BastionClient) {
	if logLevel == "DEBUG" {
		fmt.Println("\nGetting bastion for: " + bastionName + " (" + bastionId + ")")
	}

	_, err := client.GetBastion(context.Background(), bastion.GetBastionRequest{BastionId: &bastionId})
	checkError(err)
}

func getSshPubKeyContents(sshPrivateKeyFileLocation string) string {
	homeDir := getHomeDir()

	if sshPrivateKeyFileLocation == "" {
		sshPrivateKeyFileLocation = homeDir + "/.ssh/id_rsa"
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
	boldBlue.Println("\nSession ID")
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
	boldBlue.Println("\nSession ID")
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

	boldBlue.Println("Active bastion sessions")
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
		boldYellow.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p -p 22 " + bastionHost + "' \\")
		fmt.Println("-P " + strconv.Itoa(*sshPort) + " " + *sshUser + "@" + *instanceIp + " -N -L " + color.RedString("LOCAL_PORT") + ":" + *instanceIp + ":" + color.RedString("REMOTE_PORT"))
	} else if *localPort != 0 {
		boldYellow.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p -p 22 " + bastionHost + "' \\")
		fmt.Println("-P " + strconv.Itoa(*sshPort) + " " + *sshUser + "@" + *instanceIp + " -N -L " + strconv.Itoa(*localPort) + ":" + *instanceIp + ":" + strconv.Itoa(tunnelPort))
	} else {
		boldYellow.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p -p 22 " + bastionHost + "' \\")
		fmt.Println("-P " + strconv.Itoa(*sshPort) + " " + *sshUser + "@" + *instanceIp + " -N -L " + strconv.Itoa(tunnelPort) + ":" + *instanceIp + ":" + strconv.Itoa(tunnelPort))
	}

	boldYellow.Println("\nSCP command")
	fmt.Println("scp -i " + sshIdentityFile + " -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P " + strconv.Itoa(*sshPort) + " \\")
	fmt.Println("-o ProxyCommand='ssh -i " + sshIdentityFile + " -W %h:%p -p 22 " + bastionHost + "' \\")
	fmt.Println(color.RedString("SOURCE_PATH ") + *sshUser + "@" + *instanceIp + ":" + color.RedString("TARGET_PATH"))

	boldYellow.Println("\nSSH comand")
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
		boldYellow.Println("\nUpdate kube config (One time operation)")
		fmt.Println("oci ce cluster create-kubeconfig --cluster-id " + *flagOkeClusterId + " --token-version 2.0.0 --kube-endpoint PRIVATE_ENDPOINT --auth security_token")
	}

	boldYellow.Println("\nPort Forwarding command")
	fmt.Println("ssh -i \"" + sshIdentityFile + "\" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")

	if tunnelPort == 0 {
		fmt.Println("-p " + strconv.Itoa(*sshPort) + " -N -L " + color.RedString("LOCAL_PORT") + ":" + *targetIp + ":" + color.RedString("REMOTE_PORT") + " " + bastionHost)
	} else if localTunnelPort != 0 {
		fmt.Println("-p " + strconv.Itoa(*sshPort) + " -N -L " + strconv.Itoa(localTunnelPort) + ":" + *targetIp + ":" + strconv.Itoa(tunnelPort) + " " + bastionHost)
	} else {
		fmt.Println("-p " + strconv.Itoa(*sshPort) + " -N -L " + strconv.Itoa(tunnelPort) + ":" + *targetIp + ":" + strconv.Itoa(tunnelPort) + " " + bastionHost)
	}
}

func findAndPrintMatches(computeClient core.ComputeClient, compartmentId string, flagSearchString string, vnetClient core.VirtualNetworkClient, containerEngineClient containerengine.ContainerEngineClient) {
	pattern := flagSearchString

	clusters := getClusters(containerEngineClient, compartmentId)
	clusterMatches := searchClusters(pattern, clusters)

	if len(clusterMatches) > 0 {
		boldYellow.Println("OKE Clusters")
		for _, cluster := range clusterMatches {
			boldBlue.Println("Name: " + cluster.name)
			fmt.Println("Cluster ID: " + cluster.id)
			fmt.Println("Private endpoint: " + cluster.privateEndpointIp + ":" + cluster.privateEndpointPort)
			fmt.Println("")
		}
	}

	// Get relevant info for ALL instances
	// We have to do this because GetInstance/ListInstancesRequest does not allow filtering in the request
	instances := getInstances(computeClient, compartmentId)
	// returns map of instanceId: instanceName

	//Search instance info and return instance names and instance IDs of matches on instance name
	instanceMatches := searchInstances(pattern, instances)
	// returns map of instanceName: instanceId

	// Get ALL VNIC attachments
	// Once again, doing this because there is no filtering in the request
	attachments := getVnicAttachments(computeClient, compartmentId)
	// returns map of instanceId: vnicId

	// For all matches lookup VNIC ID based on instanceId and then return the private IP associated with the VNIC ID
	var Instances []Instance

	for instanceName, instanceId := range instanceMatches {
		vnicId, ok := attachments[instanceId]
		if ok {
			if logLevel == "DEBUG" {
				fmt.Println("VNic ID: " + vnicId)
			}

			privateIp := getPrivateIp(vnetClient, vnicId)
			instance := Instance{instanceName, instanceId, privateIp}
			Instances = append(Instances, instance)

		} else {
			fmt.Println("Unable to lookup VNIC for " + instanceId)
		}
	}

	if len(Instances) > 0 {
		sort.Sort(instancesByName(Instances))

		boldYellow.Println("Instances")
		for _, instance := range Instances {
			boldBlue.Println("Name: " + instance.name)
			fmt.Println("Instance ID: " + instance.id)
			fmt.Println("Private IP: " + instance.ip)
			fmt.Println("")
		}
	}

	os.Exit(0)
}

func main() {
	// TODO: switch to more mature cmd line flag parsing library
	flagTenancyId := flag.String("t", "", "tenancy ID name")

	flagListCompartments := flag.Bool("lc", false, "list compartments")
	flagListBastions := flag.Bool("lb", false, "list bastions")

	flagSearchString := flag.String("f", "", "search string to search for instance")

	flagInstanceId := flag.String("o", "", "instance ID of host to connect to")
	flagTargetIp := flag.String("i", "", "IP address of host/endpoint to connect to")

	flagCompartmentName := flag.String("c", "", "compartment name")
	flagBastionName := flag.String("b", "", "bastion name")

	flagSessionId := flag.String("s", "", "Session ID to check for")
	flagListSessions := flag.Bool("ls", false, "list sessions")
	flagSshUser := flag.String("u", "opc", "SSH user")
	flagSshPort := flag.Int("p", 22, "SSH port")
	flagSshPrivateKey := flag.String("k", "", "path to SSH private key (identity file)")
	flagSshPublicKey := flag.String("e", "", "path to SSH public key")
	flagSessionTtl := flag.Int("l", 10800, "Session TTL (seconds)")

	flagCreatePortFwSession := flag.Bool("fw", false, "Create an SSH port forward session")

	// For managed SSH sessions: tp will be used for both LOCAL and REMOTE port in tunnel command
	// For port forward SSH sessions: tp will be used for:
	//    1. the session's target port
	//    2. both LOCAL and REMOTE port in tunnel command
	flagSshTunnelPort := flag.Int("tp", 0, "SSH tunnel port")

	// This will override the local port for both managed SSH and and port forward sessions
	flagSshTunnelPortOverrideLocal := flag.Int("tpl", 0, "SSH tunnel local port")

	flagOkeClusterId := flag.String("oke", "", "OKE cluster ID")

	flagVersion := flag.Bool("v", false, "Show version")

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

	if *flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	identityClient, bastionClient, computeClient, vnetClient, containerEngineClient := initializeOciClients()

	tenancyId := getTenancyId(*flagTenancyId, identityClient)
	compartmentInfo := getCompartmentInfo(tenancyId, identityClient)

	if *flagListCompartments {
		listCompartmentNames(compartmentInfo)
		os.Exit(0)
	}

	// Anything past this point requires a compartment and bastion info
	compartmentName := getCompartmentName(*flagCompartmentName)
	compartmentId := getCompartmentId(compartmentInfo, compartmentName)
	bastionInfo := getBastionInfo(compartmentId, bastionClient)

	if *flagSearchString != "" {
		findAndPrintMatches(computeClient, compartmentId, *flagSearchString, vnetClient, containerEngineClient)
	}

	if *flagListBastions {
		listBastions(compartmentName, bastionInfo)
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

			if logLevel == "DEBUG" {
				fmt.Println("Only one bastion found, using it")
				fmt.Println(bastionName + " (" + bastionId + ")")
			}
		}
	} else {
		// There were multiple bastions found so we'll need to know which one to use
		bastionName = getBastionName(*flagBastionName)
		bastionId = bastionInfo[bastionName]
	}

	getBastion(bastionName, bastionId, bastionClient)

	if *flagListSessions {
		listActiveSessions(bastionClient, bastionId)
		os.Exit(0)
	}

	homeDir := getHomeDir()

	var sshPrivateKeyFileLocation string
	if *flagSshPrivateKey == "" {
		// TODO: move this default to flags
		sshPrivateKeyFileLocation = homeDir + "/.ssh/id_rsa"
		if logLevel == "DEBUG" {
			fmt.Println("Using default SSH private key file at " + sshPrivateKeyFileLocation)
		}
	} else {
		sshPrivateKeyFileLocation = *flagSshPrivateKey
	}

	var sshPublicKeyFileLocation string
	if *flagSshPublicKey == "" {
		// TODO: move this default to flags
		sshPublicKeyFileLocation = homeDir + "/.ssh/id_rsa.pub"
		if logLevel == "DEBUG" {
			fmt.Println("\nUsing default SSH public key file at " + sshPublicKeyFileLocation)
		}
	} else {
		sshPublicKeyFileLocation = *flagSshPublicKey
	}

	publicKeyContent := getSshPubKeyContents(sshPublicKeyFileLocation)

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
