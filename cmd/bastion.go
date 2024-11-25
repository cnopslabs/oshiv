package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/cnopslabs/oshiv/internal/resources"
	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/bastion"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bastionCmd = &cobra.Command{
	Use:   "bastion",
	Short: "Find, list, and connect via the OCI bastion service",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		// Lookup tenancy ID and compartment flags and add to Viper config if passed
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		FlagCompartment := rootCmd.Flags().Lookup("compartment")
		utils.ConfigInit(FlagTenancyId, FlagCompartment)

		// Get tenancy ID and tenancy name from Viper config
		tenancyName := viper.GetString("tenancy-name")
		tenancyId := viper.GetString("tenancy-id")
		compartmentName := viper.GetString("compartment")

		ociConfig := utils.SetupOciConfig()
		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(ociConfig)
		utils.CheckError(identityErr)

		compartments := resources.FetchCompartments(tenancyId, identityClient)
		compartmentId := resources.LookupCompartmentId(compartments, tenancyId, tenancyName, compartmentName)

		containerEngineClient, err := containerengine.NewContainerEngineClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		bastionClient, err := bastion.NewBastionClientWithConfigurationProvider(ociConfig)
		utils.CheckError(err)

		bastions := resources.FetchBastions(compartmentId, bastionClient)

		flagList, _ := cmd.Flags().GetBool("list")

		flagCreate, _ := cmd.Flags().GetBool("create")
		flagType, _ := cmd.Flags().GetString("type")
		flagBastionId, _ := cmd.Flags().GetString("bastion-id")
		flagTargetIp, _ := cmd.Flags().GetString("target-ip")
		flagSshPort, _ := cmd.Flags().GetInt("ssh-port")
		flagTtl, _ := cmd.Flags().GetInt("ttl")

		flagSshPrivateKey, _ := cmd.Flags().GetString("private-key")
		flagSshPublicKey, _ := cmd.Flags().GetString("public-key")

		flagOkeName, _ := cmd.Flags().GetString("oke-name")

		flagLocalFwPort, _ := cmd.Flags().GetInt("local-fw-port")
		flagRemoteFwPort, _ := cmd.Flags().GetInt("remote-fw-port")

		if flagList {
			resources.ListBastions(bastions, tenancyName, compartmentName)
			os.Exit(0)
		} else if flagCreate {
			// Check if there's only one bastion, if so use it (no input required)
			var bastionId string
			_, uniqueBastionId := resources.CheckForUniqueBastion(bastions)
			if uniqueBastionId != "" {
				bastionId = uniqueBastionId
			} else {
				bastionId = flagBastionId
			}

			publicKeyContent, err := os.ReadFile(flagSshPublicKey)
			utils.CheckError(err)

			switch flagType {
			case "port-forward":
				utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartmentName + ")")

				sessionId := resources.CreatePortFwSession(bastionId, bastionClient, flagTargetIp, string(publicKeyContent), flagRemoteFwPort, flagTtl)
				session := resources.FetchSession(bastionClient, sessionId, flagType)

				for session.State != "ACTIVE" {
					if session.State == "DELETED" {
						fmt.Println("\nSession has been deleted, exiting")
						fmt.Println("State: " + session.State)
						fmt.Println("\nSession Info")
						fmt.Println(session)
						os.Exit(1)
					} else {
						fmt.Println("Session not yet active, waiting... (State: " + session.State + ")")
						time.Sleep(10 * time.Second)
						session = resources.FetchSession(bastionClient, sessionId, flagType)
					}
				}

				if flagOkeName != "" {
					// If creating bastion session to an OKE cluster, lookup cluster ID
					flagOkeId := resources.FetchClusterId(containerEngineClient, compartmentId, flagOkeName)
					resources.PrintPortFwSshCommands(bastionClient, sessionId, flagTargetIp, flagSshPort, flagSshPrivateKey, flagLocalFwPort, flagRemoteFwPort, flagOkeId)
				} else {
					resources.PrintPortFwSshCommands(bastionClient, sessionId, flagTargetIp, flagSshPort, flagSshPrivateKey, flagLocalFwPort, flagRemoteFwPort, "")
				}
			case "managed":
				fmt.Println("Not yet implemented")
			default:
				fmt.Println("Not yet implemented")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(bastionCmd)

	homeDir := utils.HomeDir()
	defaultPrivateKeyPath := homeDir + "/.ssh/id_rsa"
	defaultPublicKeyPath := homeDir + "/.ssh/id_rsa.pub"

	// Local flags only exposed to oke command
	bastionCmd.Flags().BoolP("list", "l", false, "List all bastions")

	bastionCmd.Flags().BoolP("create", "r", true, "Create bastion session")

	bastionCmd.Flags().StringP("type", "y", "managed", "The type of bastion session to create (managed/port-forward)")

	bastionCmd.Flags().StringP("bastion-id", "b", "", "ID of the bastion to use")    // TODO: Switch to bastion name
	bastionCmd.Flags().StringP("target-ip", "i", "", "IP of the host to connect to") // TODO: Only require hostname or IP
	bastionCmd.Flags().IntP("ssh-port", "p", 22, "Port: TODO: describe")
	bastionCmd.Flags().IntP("ttl", "m", 10800, "Bastion session TTL")

	bastionCmd.Flags().StringP("private-key", "a", defaultPrivateKeyPath, "Path to SSH private key (identity file)")
	bastionCmd.Flags().StringP("public-key", "e", defaultPublicKeyPath, "Path to SSH public key")

	bastionCmd.Flags().StringP("oke-name", "k", "", "Name of the OKE cluster to connect to")

	bastionCmd.Flags().IntP("local-fw-port", "f", 6443, "Local port to forward")
	bastionCmd.Flags().IntP("remote-fw-port", "w", 6443, "Remote port to forward")
}
