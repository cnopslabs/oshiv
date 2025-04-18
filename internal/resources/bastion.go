package resources

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/fatih/color"
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

// Check status of bastion session
func FetchSession(bastionClient bastion.BastionClient, sessionId *string, flagType string) Session {
	response, err := bastionClient.GetSession(context.Background(), bastion.GetSessionRequest{SessionId: sessionId})
	utils.CheckError(err)

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
func ListBastionSessions(bastionClient bastion.BastionClient, bastionId string, tenancyName string, compartmentName string, listOnlyActiveSessions bool) {
	response, err := bastionClient.ListSessions(context.Background(), bastion.ListSessionsRequest{BastionId: &bastionId})
	utils.CheckError(err)

	utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartmentName + ")")

	// TODO: Fix inefficient code for bastion session object
	// put sessions into []Session and then handle active vs all
	// or maybe get rid of, listing non-active session may only be useful for troubleshooting
	for _, session := range response.Items {
		if listOnlyActiveSessions {
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
		} else {
			fmt.Print("Name: ")
			utils.Blue.Println(*session.DisplayName)
			fmt.Print("State: ")
			utils.Blue.Println(session.LifecycleState)
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

// Create a port forward SSH bastion session
func CreateBastionSession(bastionClient bastion.BastionClient, bastionId string, sessionType string, publicKeyContent string, targetIp string, sshPort int, hostFwPort int, sessionTtl int, targetInstanceId string, sshUser string) *string {
	var req bastion.CreateSessionRequest

	id := utils.GenerateID(4) // 4 bytes = ~6 chars
	targetIpSafe := strings.ReplaceAll(targetIp, ".", "-")

	switch sessionType {
	case "port-forward":
		fmt.Println("Creating port forwarding SSH session...")

		req = bastion.CreateSessionRequest{
			CreateSessionDetails: bastion.CreateSessionDetails{
				BastionId:           &bastionId,
				DisplayName:         common.String("oshiv-" + "pt-fw-" + targetIpSafe + "-" + strconv.Itoa(hostFwPort) + id),
				KeyDetails:          &bastion.PublicKeyDetails{PublicKeyContent: &publicKeyContent},
				SessionTtlInSeconds: common.Int(sessionTtl),
				TargetResourceDetails: bastion.PortForwardingSessionTargetResourceDetails{
					TargetResourcePort:             &hostFwPort, // In the case of a port fw session, port represents the host-port to forward (as in ssh -L port:host:host-port)
					TargetResourcePrivateIpAddress: &targetIp,
				},
			},
		}

	case "managed":
		fmt.Println("Creating managed SSH session...")

		utils.Logger.Debug("targetInstanceId: " + targetInstanceId)
		utils.Logger.Debug("sshUser: " + sshUser)
		utils.Logger.Debug("sshPort: " + strconv.Itoa(sshPort))
		utils.Logger.Debug("targetIp: " + targetIp)

		req = bastion.CreateSessionRequest{
			CreateSessionDetails: bastion.CreateSessionDetails{
				BastionId:           &bastionId,
				DisplayName:         common.String("oshiv-" + "mng-ssh-" + targetIpSafe + id),
				KeyDetails:          &bastion.PublicKeyDetails{PublicKeyContent: &publicKeyContent},
				SessionTtlInSeconds: common.Int(sessionTtl),
				TargetResourceDetails: bastion.CreateManagedSshSessionTargetResourceDetails{
					TargetResourceId:                      &targetInstanceId,
					TargetResourceOperatingSystemUserName: &sshUser,
					TargetResourcePort:                    &sshPort, // In the case of a managed session, port represents the port to connect to on the remote host (as in ssh -p 22)
					TargetResourcePrivateIpAddress:        &targetIp,
				},
			},
		}
	}

	response, err := bastionClient.CreateSession(context.Background(), req)
	utils.CheckError(err)

	sessionId := response.Session.Id
	utils.Blue.Println("\nSession ID")
	fmt.Println(*sessionId)
	fmt.Println("")

	return sessionId
}

// Print port forward SSH commands to connect via bastion
func PrintPortFwSshCommands(bastionClient bastion.BastionClient, sessionId *string, targetIp string, sshPort int, sshPrivateKey string, localFwPort int, hostFwPort int, flagOkeId string) {
	bastionEndpointUrl, err := url.Parse(bastionClient.Endpoint())
	utils.CheckError(err)

	bastionHost := *sessionId + "@host." + bastionEndpointUrl.Host

	if flagOkeId != "" {
		utils.Yellow.Println("\nUpdate kube config (One time operation)")
		fmt.Println("oci ce cluster create-kubeconfig --cluster-id " + flagOkeId + " --token-version 2.0.0 --kube-endpoint PRIVATE_ENDPOINT --auth security_token")
	}

	utils.Yellow.Println("\nPort Forwarding command")
	fmt.Println("ssh -i \"" + sshPrivateKey + "\" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")

	fmt.Println("-N -L " + strconv.Itoa(localFwPort) + ":" + targetIp + ":" + strconv.Itoa(hostFwPort) + " " + bastionHost)
}

// Print SSH commands to connect via bastion
func PrintManagedSshCommands(bastionClient bastion.BastionClient, sessionId *string, instanceIp string, sshUser string, sshPort int, sshIdentityFile string, localFwPort int, hostFwPort int) {
	bastionEndpointUrl, err := url.Parse(bastionClient.Endpoint())
	utils.CheckError(err)

	sessionIdStr := *sessionId
	bastionHost := sessionIdStr + "@host." + bastionEndpointUrl.Host

	// TODO: Consider proxy jump flag for commands where applicable - https://www.ateam-oracle.com/post/openssh-proxyjump-with-oci-bastion-service
	if hostFwPort == 0 {
		utils.Yellow.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p " + bastionHost + "' \\")
		fmt.Println(sshUser + "@" + instanceIp + " -N -L " + color.RedString("LOCAL_PORT") + ":" + instanceIp + ":" + color.RedString("REMOTE_PORT"))
	} else if localFwPort != 0 {
		utils.Yellow.Println("\nTunnel command")

		// Use sudo for privileged ports
		if localFwPort < 1024 {
			fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		} else {
			fmt.Println("ssh -i \"" + sshIdentityFile + "\" \\")
		}

		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p " + bastionHost + "' \\")
		fmt.Println(sshUser + "@" + instanceIp + " -N -L " + strconv.Itoa(localFwPort) + ":" + instanceIp + ":" + strconv.Itoa(hostFwPort))
	} else {
		utils.Yellow.Println("\nTunnel command")
		fmt.Println("sudo ssh -i \"" + sshIdentityFile + "\" \\")
		fmt.Println("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
		fmt.Println("-o ProxyCommand='ssh -i \"" + sshIdentityFile + "\" -W %h:%p " + bastionHost + "' \\")
		fmt.Println(sshUser + "@" + instanceIp + " -N -L " + strconv.Itoa(hostFwPort) + ":" + instanceIp + ":" + strconv.Itoa(hostFwPort))
	}

	utils.Yellow.Println("\nSCP command")
	fmt.Println("scp -i " + sshIdentityFile + " -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" + " \\")
	fmt.Println("-o ProxyCommand='ssh -i " + sshIdentityFile + " -W %h:%p " + bastionHost + "' \\")
	fmt.Println(color.RedString("SOURCE_PATH ") + sshUser + "@" + instanceIp + ":" + color.RedString("TARGET_PATH"))

	utils.Yellow.Println("\nSSH command")
	fmt.Println("ssh -i " + sshIdentityFile + " -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\")
	fmt.Println("-o ProxyCommand='ssh -i " + sshIdentityFile + " -W %h:%p " + bastionHost + "' \\")
	fmt.Println(sshUser + "@" + instanceIp)
}
