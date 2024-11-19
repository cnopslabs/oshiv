package resources

import (
	"context"
	"fmt"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/bastion"
	"github.com/rodaine/table"
)

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
func ListBastions(bastions map[string]string) {
	tbl := table.New("Bastion Name", "OCID")
	tbl.WithHeaderFormatter(utils.HeaderFmt).WithFirstColumnFormatter(utils.ColumnFmt)

	for bastionName := range bastions {
		tbl.AddRow(bastionName, bastions[bastionName])
	}

	tbl.Print()

	fmt.Println("\nTo set bastion name, run:")
	utils.Yellow.Println("   oshiv bastion -s BASTION_NAME")
}

// Determine bastion name and then lookup ID
func CheckForUniqueBastion(bastions map[string]string) string {
	var bastionId string
	var bastionName string

	// If there is only one bastion, no need to require bastion name input
	if len(bastions) == 1 {
		for name, id := range bastions {
			bastionName = name
			bastionId = id
		}

		utils.Logger.Debug("Only one bastion found, using " + bastionName + " (" + bastionId + ")")
		return bastionName

	} else {
		utils.Logger.Debug("Multiple bastions found")
		return ""
	}
}

// Sets bastion name in Viper config
func SetBastionName(bastionName string) {
	utils.Logger.Debug("Setting bastion: " + bastionName)
	utils.SetConfigString("bastion-name", bastionName)
}

// List and print all active bastion sessions
func ListBastionSessions(client bastion.BastionClient, bastionId string) {
	response, err := client.ListSessions(context.Background(), bastion.ListSessionsRequest{BastionId: &bastionId})
	utils.CheckError(err)

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
