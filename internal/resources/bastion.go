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

	bastionInfo := make(map[string]string)

	for _, item := range response.Items {
		bastionInfo[*item.Name] = *item.Id
	}

	return bastionInfo
}

// List and print bastions (OCI API call)
func ListBastions(bastionInfo map[string]string) {
	tbl := table.New("Bastion Name", "OCID")
	tbl.WithHeaderFormatter(utils.HeaderFmt).WithFirstColumnFormatter(utils.ColumnFmt)

	for bastionName := range bastionInfo {
		tbl.AddRow(bastionName, bastionInfo[bastionName])
	}

	tbl.Print()

	fmt.Println("\nTo set bastion name, export OCI_BASTION_NAME:")
	utils.Yellow.Println("   export OCI_BASTION_NAME=")
}

// Determine bastion name and then lookup ID
func CheckForUniqueBastion(bastionInfo map[string]string) string {
	var bastionId string
	var bastionName string

	// If there is only one bastion, no need to require bastion name input
	if len(bastionInfo) == 1 {
		for name, id := range bastionInfo {
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

// func ListBastionSessions(bastionClient bastion.BastionClient, compartmentID string) ([]bastion.SessionSummary, error) {
// 	req := bastion.ListSessionsRequest{
// 		// Limit: 100, // optional, default is 50
// 		Page: nil, // optional, default is first page
// 		// SortBy:        bastion.ListSessionsSortByTimecreated, // optional, default is TimeCreated
// 		// SortOrder:     bastion.SortOrderAsc,                         // optional, default is Ascending
// 		// SortOrder:     bastion.ListSessionsSortOrderAsc,                         // optional, default is Ascending
// 	}

// 	resp, err := bastionClient.ListBastionSessions(context.Background(), req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sessions := resp.Items
// 	for resp.HasMorePages() {
// 		page := resp.OpcNextPage
// 		req.Page = &page
// 		resp, err = bastionClient.ListBastionSessions(context.Background(), req)
// 		if err != nil {
// 			return nil, err
// 		}
// 		sessions = append(sessions, resp.Items...)
// 	}

// 	// Get target resource details for each session
// 	for i, session := range sessions {
// 		sessionReq := bastion.ListBastionSessions{
// 			BastionSessionId: session.Id,
// 		}
// 		sessionResp, err := bastionClient.ListSessions(context.Background(), sessionReq)
// 		if err != nil {
// 			return nil, err
// 		}
// 		sessions[i] = sessionResp.BastionSession
// 	}

// 	return sessions, nil
// }
