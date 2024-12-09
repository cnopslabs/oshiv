package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
)

type Image struct {
	name        string
	id          string
	cDate       common.SDKTime
	freeTags    map[string]string
	definedTags map[string]map[string]interface{}
	launchMode  core.ImageLaunchModeEnum
}

// Fetch image object by ID via OCI API call
func fetchImage(computeClient core.ComputeClient, imageId string) Image {
	var image Image

	response, err := computeClient.GetImage(context.Background(), core.GetImageRequest{ImageId: &imageId})
	utils.CheckError(err)

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

// Fetch all images via OCI API call
func fetchImages(computeClient core.ComputeClient, compartmentId string) []Image {
	var images []Image
	var pageCount int
	pageCount = 0

	initialResponse, err := computeClient.ListImages(context.Background(), core.ListImagesRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

	for _, item := range initialResponse.Items {
		pageCount += 1
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
		nextPage := initialResponse.OpcNextPage

		for {
			response, err := computeClient.ListImages(context.Background(), core.ListImagesRequest{CompartmentId: &compartmentId, Page: nextPage})
			utils.CheckError(err)

			for _, item := range response.Items {
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

	return images
}

// List and print images (OCI API call)
func ListImages(computeClient core.ComputeClient, compartmentId string, compartment string, tenancyName string) {
	images := fetchImages(computeClient, compartmentId)

	utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartment + ")")

	for _, image := range images {
		fmt.Print("Name: ")
		utils.Blue.Println(image.name)

		fmt.Print("ID: ")
		utils.Yellow.Println(image.id)

		fmt.Print("Create date: ")
		utils.Yellow.Println(image.cDate)

		fmt.Println("Tags: ")

		for k, v := range image.freeTags {
			utils.Yellow.Println(k + ": " + v)
		}

		fmt.Print("Launch mode: ")
		utils.Yellow.Println(image.launchMode)

		fmt.Println("")
	}

	fmt.Println(strconv.Itoa(len(images)) + " images found")
}
