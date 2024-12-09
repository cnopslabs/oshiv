package resources

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
)

// TODO: Add operating system from image
type Instance struct {
	name     string
	id       string
	ip       string
	ad       string
	shape    string
	cDate    common.SDKTime
	imageId  string
	fd       string
	vCPUs    int
	mem      float32
	region   string
	state    core.InstanceLifecycleStateEnum
	subnetId string
}

// Sort instances by name
type instancesByName []Instance

func (instances instancesByName) Len() int           { return len(instances) }
func (instances instancesByName) Less(i, j int) bool { return instances[i].name < instances[j].name }
func (instances instancesByName) Swap(i, j int) {
	instances[i], instances[j] = instances[j], instances[i]
}

// Fetch all VNIC attachments via OCI API call
// This is used to determine instance private IP
func fetchVnicAttachments(client core.ComputeClient, compartmentId string) (map[string]string, map[string]string) {
	attachments := make(map[string]string)
	attachments_subnets := make(map[string]string)

	initialResponse, err := client.ListVnicAttachments(context.Background(), core.ListVnicAttachmentsRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

	for _, attachment := range initialResponse.Items {
		attachments[*attachment.InstanceId] = *attachment.VnicId
		attachments_subnets[*attachment.InstanceId] = *attachment.SubnetId
	}

	if initialResponse.OpcNextPage != nil {
		nextPage := initialResponse.OpcNextPage
		for {
			response, err := client.ListVnicAttachments(context.Background(), core.ListVnicAttachmentsRequest{CompartmentId: &compartmentId, Page: nextPage})
			utils.CheckError(err)

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

	return attachments, attachments_subnets
}

// Fetch private IP from VNIC (OCI API call)
func fetchPrivateIp(client core.VirtualNetworkClient, vnicId string) string {
	response, err := client.GetVnic(context.Background(), core.GetVnicRequest{VnicId: &vnicId})
	utils.CheckError(err)

	return *response.Vnic.PrivateIp
}

// Fetch all subnet IDs via OCI API call
func fetchSubnetIds(client core.VirtualNetworkClient, compartmentId string) []string {
	response, err := client.ListSubnets(context.Background(), core.ListSubnetsRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

	var subnetIds []string

	for _, subnet := range response.Items {
		subnetIds = append(subnetIds, *subnet.Id)
	}

	return subnetIds
}

// Fetch all private IPs via OCI API call
func fetchPrivateIps(client core.VirtualNetworkClient, compartmentId string) map[string]string {
	vNicIdsToIps := make(map[string]string)
	subnetIds := fetchSubnetIds(client, compartmentId)

	for _, subnetId := range subnetIds {
		response, err := client.ListPrivateIps(context.Background(), core.ListPrivateIpsRequest{SubnetId: &subnetId})
		utils.CheckError(err)

		for _, item := range response.Items {
			vNicIdsToIps[*item.VnicId] = *item.IpAddress
		}
	}

	return vNicIdsToIps
}

// Fetch all instances via OCI API call
func fetchInstances(computeClient core.ComputeClient, compartmentId string) []Instance {
	var instances []Instance

	initialResponse, err := computeClient.ListInstances(context.Background(), core.ListInstancesRequest{
		CompartmentId:  &compartmentId,
		LifecycleState: core.InstanceLifecycleStateRunning,
	})
	utils.CheckError(err)

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
			"0", // We have to lookup the subnet separately
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
			utils.CheckError(err)

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
					"0", // We have to lookup the subnet separately
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

	return instances
}

// List and print instances (OCI API call)
func ListInstances(computeClient core.ComputeClient, compartmentId string, vnetClient core.VirtualNetworkClient, compartment string, tenancyName string) {
	instances := fetchInstances(computeClient, compartmentId)
	// returns []Instance

	utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartment + ")")

	for _, instance := range instances {
		region := instance.region
		ad := instance.ad
		strToRemove := "bKwM:" + region + "-"
		ad_short := strings.Replace(ad, strToRemove, "", -1)

		fd := instance.fd
		fd_short := strings.Replace(fd, "FAULT-DOMAIN", "FD", -1)

		fmt.Print("Name: ")
		utils.Blue.Println(instance.name)

		fmt.Print("ID: ")
		utils.Yellow.Println(instance.id)

		fmt.Print("Private IP: ")
		utils.Yellow.Print(instance.ip)

		fmt.Print(" FD: ")
		utils.Yellow.Print(fd_short)

		fmt.Print(" AD: ")
		utils.Yellow.Println(ad_short)

		fmt.Print("Shape: ")
		utils.Yellow.Print(instance.shape)

		fmt.Print(" Mem: ")
		utils.Yellow.Print(instance.mem)

		fmt.Print(" vCPUs: ")
		utils.Yellow.Println(instance.vCPUs)

		fmt.Print("State: ")
		utils.Yellow.Println(instance.state)

		fmt.Print("Created: ")
		utils.Yellow.Println(instance.cDate)

		fmt.Print("Image ID: ")
		utils.Yellow.Println(instance.imageId)

		fmt.Println("")
	}
}

// Match pattern and return instance matches
func matchInstances(pattern string, instances []Instance) []Instance {
	// TODO: Maybe move this back to findInstances for consistency
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

	return matches
}

// Find and print instances (OCI API call)
// TODO: Consider consolidating FindInstances and ListInstances similar to OKE
func FindInstances(computeClient core.ComputeClient, vnetClient core.VirtualNetworkClient, compartmentId string, flagSearchString string, retrieveImageInfo bool, compartment string, tenancyName string) {
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
	utils.Faint.Println(strconv.Itoa(matchCount) + " matches")

	if matchCount > ipFetchAllThreshold {
		batchFetchAllIps = true
	} else {
		batchFetchAllIps = false
	}

	// Get ALL VNIC attachments
	// Once again, doing this because the request does not support filtering in the request
	attachments, attachments_subnets := fetchVnicAttachments(computeClient, compartmentId)
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
			if batchFetchAllIps {
				privateIp = vNicIdsToIps[vnicId]
			} else {
				privateIp = fetchPrivateIp(vnetClient, vnicId)
			}

			instance.ip = privateIp

			subnetId, ok := attachments_subnets[instance.id]
			if ok {
				instance.subnetId = subnetId
				instancesWithIP = append(instancesWithIP, instance) // TODO: Im sure theres a better way to do this using a single slice
			}

		} else {
			fmt.Println("Unable to lookup VNIC for " + instance.id)
		}
	}

	utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartment + ")")
	if len(instancesWithIP) > 0 {
		sort.Sort(instancesByName(instancesWithIP))

		for _, instance := range instancesWithIP {
			fd := instance.fd
			fd_short := strings.Replace(fd, "FAULT-DOMAIN", "FD", -1)

			fmt.Print("Name: ")
			utils.Blue.Println(instance.name)

			fmt.Print("ID: ")
			utils.Yellow.Println(instance.id)

			fmt.Print("Private IP: ")
			utils.Yellow.Print(instance.ip)

			fmt.Print(" FD: ")
			utils.Yellow.Print(fd_short)

			fmt.Print(" AD: ")
			utils.Yellow.Println(instance.ad)

			fmt.Print("Shape: ")
			utils.Yellow.Print(instance.shape)

			fmt.Print(" Mem: ")
			utils.Yellow.Print(instance.mem)

			fmt.Print(" vCPUs: ")
			utils.Yellow.Println(instance.vCPUs)

			fmt.Print("State: ")
			utils.Yellow.Println(instance.state)

			fmt.Print("Created: ")
			utils.Yellow.Println(instance.cDate)

			fmt.Print("Subnet ID: ")
			utils.Yellow.Println(instance.subnetId)

			if retrieveImageInfo {
				image := fetchImage(computeClient, instance.imageId) // TODO: Performance hit: this adds ~100 ms per image lookup

				fmt.Print("Image Name: ")
				utils.Yellow.Println(image.name)

				fmt.Print("Image ID: ")
				utils.Yellow.Println(instance.imageId)

				fmt.Print("Image Created: ")
				utils.Yellow.Println(image.cDate)

				fmt.Println("Image Tags (Free form): ")

				freeformTagKeys := make([]string, 0, len(image.freeTags))
				for key := range image.freeTags {
					freeformTagKeys = append(freeformTagKeys, key)
				}
				sort.Strings(freeformTagKeys)

				utils.Faint.Print("| ")
				for _, key := range freeformTagKeys {
					utils.Faint.Print(key + ": " + image.freeTags[key] + " | ")
				}

				fmt.Println("")

				fmt.Println("Image Tags (Defined): ")
				for tagNs, tags := range image.definedTags {
					utils.Italic.Println(tagNs)

					definedTagKeys := make([]string, 0, len(tags))
					for key := range tags {
						definedTagKeys = append(definedTagKeys, key)
					}
					sort.Strings(definedTagKeys)

					utils.Faint.Print("| ")
					for _, key := range definedTagKeys {
						utils.Faint.Print(key + ": " + tags[key].(string) + " | ")
					}

					fmt.Println("")

				}
			}

			fmt.Println("")
		}
	}
}
