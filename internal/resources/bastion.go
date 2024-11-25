package resources

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/bastion"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/rodaine/table"
)

type Session struct {
	State bastion.SessionLifecycleStateEnum
	ip    string
	user  string
	port  int
}

// Fetch all bastions via OCI API call
func FetchBastions(compartmentId string, client bastion.BastionClient) map[string]string {
	response, err := client.ListBastions(context.Background(), bastion.ListBastionsRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

	bastions := make(map[string]string)

	for _, item := range response.Items {
		bastions[*item.Name] = *item.Id
	}

	return bastions
}

// List and print bastions (OCI API call)
func ListBastions(bastions map[string]string, tenancyName string, compartmentName string) {
	tbl := table.New("Bastion Name", "OCID")
	tbl.WithHeaderFormatter(utils.HeaderFmt).WithFirstColumnFormatter(utils.ColumnFmt)

	for bastionName := range bastions {
		tbl.AddRow(bastionName, bastions[bastionName])
	}

	utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartmentName + ")")
	tbl.Print()

	fmt.Print("\nTo specify bastion, pass flag: ")
	utils.Yellow.Println("-b BASTION_NAME")
}

// Determine bastion name and then lookup ID
func CheckForUniqueBastion(bastions map[string]string) (string, string) {
	var bastionId string
	var bastionName string

	// If there is only one bastion, no need to require bastion name input
	if len(bastions) == 1 {
		for name, id := range bastions {
			bastionName = name
			bastionId = id
		}

		utils.Logger.Debug("Only one bastion found, using " + bastionName + " (" + bastionId + ")")
		return bastionName, bastionId

	} else {
		utils.Logger.Debug("Multiple bastions found")
		return "", ""
	}
}

// List and print all active bastion sessions
func ListBastionSessions(bastionClient bastion.BastionClient, bastionId string, tenancyName string, compartmentName string) {
	response, err := bastionClient.ListSessions(context.Background(), bastion.ListSessionsRequest{BastionId: &bastionId})
	utils.CheckError(err)

	utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartmentName + ")")

	for _, session := range response.Items {
		if session.LifecycleState == "ACTIVE" {
			fmt.Print("Name: ")
			utils.Blue.Println(*session.DisplayName)
			fmt.Print("ID: ")
			utils.Yellow.Println(*session.Id)

			fmt.Print("Created: ")
			utils.Yellow.Println(*session.TimeCreated)

			// TODO: Consolidate port-fw and ssh session details into a map and print all at once
			portFwTargetResourceDetails, ok := session.TargetResourceDetails.(bastion.PortForwardingSessionTargetResourceDetails)
			if ok {
				fmt.Print("Type: ")
				utils.Yellow.Println("PortForward")
				fmt.Print("IP:Port: ")
				utils.Yellow.Print(*portFwTargetResourceDetails.TargetResourcePrivateIpAddress)
				utils.Yellow.Print(":")
				utils.Yellow.Println(*portFwTargetResourceDetails.TargetResourcePort)
			}

			sshTargetResourceDetails, ok := session.TargetResourceDetails.(bastion.ManagedSshSessionTargetResourceDetails)
			if ok {
				fmt.Print("Type: ")
				utils.Yellow.Println("SSH")

				fmt.Print("Instance ID: ")
				utils.Yellow.Println(*sshTargetResourceDetails.TargetResourceId)

				fmt.Print("IP:Port: ")
				utils.Yellow.Print(*sshTargetResourceDetails.TargetResourcePrivateIpAddress)
				utils.Yellow.Print(":")
				utils.Yellow.Println(*sshTargetResourceDetails.TargetResourcePort)
			}

			fmt.Println("")
		}
	}
}

// Create a manages SSH bastion session
func CreateManagedSshSession(bastionId string, bastionClient bastion.BastionClient, targetInstance string, targetIp string, publicKeyContent string, sshUser string, sshPort int, sessionTtl int) *string {
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
	response, err := bastionClient.CreateSession(context.Background(), req)
	utils.CheckError(err)

	sessionId := response.Session.Id
	utils.Blue.Println("\nSession ID")
	fmt.Println(*sessionId)
	fmt.Println("")

	return sessionId
}

// Create a port forward SSH bastion session
func CreatePortFwSession(bastionId string, bastionClient bastion.BastionClient, targetIp string, publicKeyContent string, remoteFwPort int, sessionTtl int) *string {
	req := bastion.CreateSessionRequest{
		CreateSessionDetails: bastion.CreateSessionDetails{
			BastionId:           &bastionId,
			DisplayName:         common.String("oshivSession"),
			KeyDetails:          &bastion.PublicKeyDetails{PublicKeyContent: &publicKeyContent},
			SessionTtlInSeconds: common.Int(sessionTtl),
			TargetResourceDetails: bastion.PortForwardingSessionTargetResourceDetails{
				TargetResourcePort:             &remoteFwPort,
				TargetResourcePrivateIpAddress: &targetIp,
			},
		},
	}

	fmt.Println("Creating port forwarding session...")
	response, err := bastionClient.CreateSession(context.Background(), req)
	utils.CheckError(err)

	sessionId := response.Session.Id
	utils.Blue.Println("\nSession ID")
	fmt.Println(*sessionId)
	fmt.Println("")

	return sessionId
}

// Check status of bastion session
func FetchSession(bastionClient bastion.BastionClient, sessionId *string, flagType string) Session {
	response, err := bastionClient.GetSession(context.Background(), bastion.GetSessionRequest{SessionId: sessionId})
	utils.CheckError(err)

	// utils.Logger.Debug("GetSessionResponse", response.Session)
	// utils.Logger.Debug("nEndpoint", client.Endpoint())

	var ipAddress *string
	var sshUser *string
	var sshPort *int

	if flagType == "port-forward" {
		// Required info for port forward SSH connections
		sshSessionTargetResourceDetails := response.Session.TargetResourceDetails.(bastion.PortForwardingSessionTargetResourceDetails)
		ipAddress = sshSessionTargetResourceDetails.TargetResourcePrivateIpAddress
		sshPort = sshSessionTargetResourceDetails.TargetResourcePort
	} else {
		// Required info for managed SSH connections
		sshSessionTargetResourceDetails := response.Session.TargetResourceDetails.(bastion.ManagedSshSessionTargetResourceDetails)
		ipAddress = sshSessionTargetResourceDetails.TargetResourcePrivateIpAddress
		sshUser = sshSessionTargetResourceDetails.TargetResourceOperatingSystemUserName
		sshPort = sshSessionTargetResourceDetails.TargetResourcePort
	}

	var session Session
	if flagType == "port-forward" {
		session = Session{response.Session.LifecycleState, *ipAddress, "", *sshPort}
	} else {
		session = Session{response.Session.LifecycleState, *ipAddress, *sshUser, *sshPort}
	}

	return session
}

// Print port forward SSH commands to connect via bastion
func PrintPortFwSshCommands(bastionClient bastion.BastionClient, sessionId *string, targetIp string, sshPort int, sshIdentityFile string, localFwPort int, remoteFwPort int, flagOkeId string) {
	bastionEndpointUrl, err := url.Parse(bastionClient.Endpoint())
	utils.CheckError(err)

	// sessionIdStr := *sessionId
	// bastionHost := sessionIdStr + "@host." + bastionEndpointUrl.Host

	bastionHost := *sessionId + "@host." + bastionEndpointUrl.Host

	if flagOkeId != "" {
		utils.Yellow.Println("\nUpdate kube config (One time operation)")
		fmt.Println("oci ce cluster create-kubeconfig --cluster-id " + flagOkeId + " --token-version 2.0.0 --kube-endpoint PRIVATE_ENDPOINT --auth security_token")
	}

	utils.Yellow.Println("\nPort Forwarding command")
	fmt.Println("ssh -i \"" + sshIdentityFile + "\" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")

	fmt.Println("-p " + strconv.Itoa(sshPort) + " -N -L " + strconv.Itoa(localFwPort) + ":" + targetIp + ":" + strconv.Itoa(remoteFwPort) + " " + bastionHost)
}
