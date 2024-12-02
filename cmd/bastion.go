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
		// Lookup tenancy ID and compartment flags, add to Viper config if passed
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
		flagSessionType, _ := cmd.Flags().GetString("type")

		// Flags applicable to all session types
		flagBastionId, _ := cmd.Flags().GetString("bastion-id")
		flagTargetIp, _ := cmd.Flags().GetString("target-ip")
		flagTtl, _ := cmd.Flags().GetInt("ttl")
		flagSshPrivateKey, _ := cmd.Flags().GetString("private-key")
		flagSshPublicKey, _ := cmd.Flags().GetString("public-key")

		// Flags applicable to managed sessions
		flagSshPort, _ := cmd.Flags().GetInt("ssh-port")
		flagInstanceId, _ := cmd.Flags().GetString("instance-id")
		flagSshUser, _ := cmd.Flags().GetString("user")

		// Flags applicable to port forward sessions
		flagOkeName, _ := cmd.Flags().GetString("oke-name")
		flagLocalFwPort, _ := cmd.Flags().GetInt("local-fw-port")
		flagHostFwPort, _ := cmd.Flags().GetInt("host-fw-port")

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

			// Get SSH public key
			publicKeyContent, err := os.ReadFile(flagSshPublicKey)
			utils.CheckError(err)

			// Create the bastion session
			utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartmentName + ")")
			sessionId := resources.CreateBastionSession(bastionClient, bastionId, flagSessionType, string(publicKeyContent), flagTargetIp, flagSshPort, flagHostFwPort, flagTtl, flagInstanceId, flagSshUser)
			session := resources.FetchSession(bastionClient, sessionId, flagSessionType)

			// Wait until session is active
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
					session = resources.FetchSession(bastionClient, sessionId, flagSessionType)
				}
			}

			// Flex print commands between port forward and managed type
			if flagSessionType == "port-forward" {
				if flagOkeName != "" {
					// If creating bastion session to an OKE cluster, lookup cluster ID and set ports to 6443
					flagOkeId := resources.FetchClusterId(containerEngineClient, compartmentId, flagOkeName)
					resources.PrintPortFwSshCommands(bastionClient, sessionId, flagTargetIp, flagSshPort, flagSshPrivateKey, 6443, 6443, flagOkeId)
				} else {
					resources.PrintPortFwSshCommands(bastionClient, sessionId, flagTargetIp, flagSshPort, flagSshPrivateKey, flagLocalFwPort, flagHostFwPort, "")
				}
			} else if flagSessionType == "managed" {
				resources.PrintManagedSshCommands(bastionClient, sessionId, flagTargetIp, flagSshUser, flagSshPort, flagSshPrivateKey, flagLocalFwPort, flagHostFwPort)
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

	// Flags applicable to all session types
	bastionCmd.Flags().StringP("bastion-id", "b", "", "ID of the bastion to use")    // TODO: Switch to bastion name
	bastionCmd.Flags().StringP("target-ip", "i", "", "IP of the host to connect to") // TODO: Only require hostname or IP
	bastionCmd.Flags().IntP("ttl", "m", 10800, "Bastion session TTL")
	bastionCmd.Flags().StringP("private-key", "a", defaultPrivateKeyPath, "Path to SSH private key (identity file)")
	bastionCmd.Flags().StringP("public-key", "e", defaultPublicKeyPath, "Path to SSH public key")
	bastionCmd.Flags().IntP("ssh-port", "p", 22, "Port to connect to on the remote host")

	// Flags applicable to managed sessions
	bastionCmd.Flags().StringP("instance-id", "o", "", "The OCID of the instance to connect to") // TODO: Only require hostname or IP
	bastionCmd.Flags().StringP("user", "u", "opc", "The SSH username to use to connect to an instance")

	// Flags applicable to port forward sessions
	// Set both the local and host forwarding ports to 0 but if only hostFwPort is passed, use that for both unless localFwPort is explicitly passed
	hostFwPort := 0
	localFwPort := hostFwPort
	bastionCmd.Flags().StringP("oke-name", "k", "", "Name of the OKE cluster to connect to")
	bastionCmd.Flags().IntP("local-fw-port", "w", localFwPort, "The port on the local (client) host to forward connections from")
	bastionCmd.Flags().IntP("host-fw-port", "f", hostFwPort, "The host port that connections are forwarded to")
}
