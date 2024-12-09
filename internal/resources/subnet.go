package resources

import (
	"context"
	"sort"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/rodaine/table"
)

type Subnet struct {
	cidr       string
	name       string
	access     string
	subnetType string
}

// TODO: This sorts alphabetically, so not great for CIDR blocks. Revert to sort by name or create CIDR sort function
// Sort subnets bt CIDR
type subnetsByCidr []Subnet

func (subnets subnetsByCidr) Len() int           { return len(subnets) }
func (subnets subnetsByCidr) Less(i, j int) bool { return subnets[i].cidr < subnets[j].cidr }
func (subnets subnetsByCidr) Swap(i, j int)      { subnets[i], subnets[j] = subnets[j], subnets[i] }

// List and print subnets (OCI API call)
func ListSubnets(client core.VirtualNetworkClient, compartmentId string) {
	response, err := client.ListSubnets(context.Background(), core.ListSubnetsRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

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
	tbl.WithHeaderFormatter(utils.HeaderFmt).WithFirstColumnFormatter(utils.ColumnFmt)

	for _, subnet := range Subnets {
		tbl.AddRow(subnet.cidr, subnet.name, subnet.access, subnet.subnetType)
	}

	tbl.Print()
}
