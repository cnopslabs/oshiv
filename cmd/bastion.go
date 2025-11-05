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
	Use:     "bastion",
	Short:   "Find, list, and connect to resources via the OCI bastion service",
	Long:    "Find, list, and connect to resources via the OCI bastion service",
	Aliases: []string{"bast"},
	Run: func(cmd *cobra.Command, args []string) {
		identityClient, identityErr := identity.NewIdentityClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(identityErr)

		// Read tenancy ID flag and calculate tenancy
		FlagTenancyId := rootCmd.Flags().Lookup("tenancy-id")
		utils.SetTenancyConfig(FlagTenancyId, utils.OciConfig())
		tenancyId := viper.GetString("tenancy-id")
		tenancyName := viper.GetString("tenancy-name")

		// Read compartment flag and add to Viper config
		FlagCompartment := rootCmd.Flags().Lookup("compartment")
		compartments := resources.FetchCompartments(tenancyId, identityClient)
		utils.SetCompartmentConfig(FlagCompartment, compartments, tenancyName)
		compartment := viper.GetString("compartment")

		compartmentId := resources.LookupCompartmentId(compartments, tenancyId, tenancyName, compartment)

		containerEngineClient, err := containerengine.NewContainerEngineClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(err)

		bastionClient, err := bastion.NewBastionClientWithConfigurationProvider(utils.OciConfig())
		utils.CheckError(err)

		region, envVarExists := os.LookupEnv("OCI_CLI_REGION")
		if envVarExists {
			containerEngineClient.SetRegion(region)
			bastionClient.SetRegion(region)
		}

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
		flagInstanceId, _ := cmd.Flags().GetString("instance-id")
		flagSshUser, _ := cmd.Flags().GetString("user")

		// Flags applicable to port forward sessions
		flagOkeName, _ := cmd.Flags().GetString("oke-name")
		flagLocalFwPort, _ := cmd.Flags().GetInt("local-fw-port")
		flagHostFwPort, _ := cmd.Flags().GetInt("host-fw-port")

		// if only hostFwPort is passed, set flagLocalFwPort to match
		// This allows a sane default of ports for an SSH address like: -L 443:IP_ADDRESS:443
		if flagLocalFwPort == 0 && flagHostFwPort != 0 {
			flagLocalFwPort = flagHostFwPort
		}

		if flagList {
			resources.ListBastions(bastions, tenancyName, compartment)
			os.Exit(0)
		} else if flagCreate {
			// Check if there's only one bastion, if so use it (no input required)
			var bastionId string
			_, uniqueBastionId := resources.CheckForUniqueBastion(bastions)
			if uniqueBastionId != "" {
				bastionId = uniqueBastionId
			} else {
				if flagBastionId == "" {
					fmt.Println("Multiple bastions found, must pass flag --bastion-id:")
					fmt.Println("Bastions:")

					for k, v := range bastions {
						fmt.Println(" - " + k + ": " + v)
					}

					os.Exit(1)
				}

				bastionId = flagBastionId
			}

			// Get SSH public key
			publicKeyContent, err := os.ReadFile(flagSshPublicKey)
			utils.CheckError(err)

			// Create the bastion session
			utils.FaintMagenta.Println("Tenancy(Compartment): " + tenancyName + "(" + compartment + ")")

			var sessionId *string
			if flagOkeName != "" {
				sessionId = resources.CreateBastionSession(bastionClient, bastionId, flagSessionType, string(publicKeyContent), flagTargetIp, 22, 6443, flagTtl, flagInstanceId, flagSshUser)
			} else {
				sessionId = resources.CreateBastionSession(bastionClient, bastionId, flagSessionType, string(publicKeyContent), flagTargetIp, 22, flagHostFwPort, flagTtl, flagInstanceId, flagSshUser)
			}

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
					fmt.Println("flagOkeId")
					resources.PrintPortFwSshCommands(bastionClient, sessionId, flagTargetIp, 22, flagSshPrivateKey, 6443, 6443, flagOkeId)
				} else {
					resources.PrintPortFwSshCommands(bastionClient, sessionId, flagTargetIp, 22, flagSshPrivateKey, flagLocalFwPort, flagHostFwPort, "")
				}
			} else if flagSessionType == "managed" {
				resources.PrintManagedSshCommands(bastionClient, sessionId, flagTargetIp, flagSshUser, 22, flagSshPrivateKey, flagLocalFwPort, flagHostFwPort)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(bastionCmd)

	homeDir := utils.HomeDir()
	// If OSHIV_SSH_HOME is set, we'll use this location for the SSH keys
	sshKeyHomeEnv := os.Getenv("OSHIV_SSH_HOME")
	var sshKeyHome string

	if sshKeyHomeEnv == "" {
		// Use default SSH keys location
		sshKeyHome = homeDir + "/.ssh"
	} else {
		// Use custom SSH keys location
		sshKeyHome = sshKeyHomeEnv
	}

	defaultPrivateKeyPath := sshKeyHome + "/id_rsa"
	defaultPublicKeyPath := sshKeyHome + "/id_rsa.pub"

	bastionCmd.Flags().BoolP("list", "l", false, "List all bastions")

	bastionCmd.Flags().BoolP("create", "r", true, "Create bastion session")
	bastionCmd.Flags().StringP("type", "y", "managed", "The type of bastion session to create (managed/port-forward)")

	// Flags applicable to all session types
	bastionCmd.Flags().StringP("bastion-id", "b", "", "ID of the bastion to use")    // TODO: Switch to bastion name
	bastionCmd.Flags().StringP("target-ip", "i", "", "IP of the host to connect to") // TODO: Only require one of: (hostname, IP, or instance ID). The lookup the others
	bastionCmd.Flags().IntP("ttl", "m", 10800, "Bastion session TTL")
	bastionCmd.Flags().StringP("private-key", "a", defaultPrivateKeyPath, "Path to SSH private key (identity file)")
	bastionCmd.Flags().StringP("public-key", "e", defaultPublicKeyPath, "Path to SSH public key")

	// Flags applicable to managed sessions
	bastionCmd.Flags().StringP("instance-id", "o", "", "The OCID of the instance to connect to")
	bastionCmd.Flags().StringP("user", "u", "opc", "The SSH username to use to connect to an instance")

	// Flags applicable to port forward sessions
	bastionCmd.Flags().StringP("oke-name", "k", "", "Name of the OKE cluster to connect to")

	bastionCmd.Flags().IntP("local-fw-port", "w", 0, "The port on the local (client) host to forward connections from")
	bastionCmd.Flags().IntP("host-fw-port", "f", 0, "The host port that connections are forwarded to")
}
