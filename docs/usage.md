# Usage Guide

## Prerequisites

### 1. Optional: Install and configure oshell

[oshell](https://github.com/cnopslabs/oshell) is `oshiv`'s companion shell helper. It helps configure OCI CLI auth and refreshes your OCI tokens automatically.

[https://github.com/cnopslabs/oshell](https://github.com/cnopslabs/oshell)

### 2. OCI Authentication and Authorization

`oshiv` relies on the OCI CLI for authentication and authorization. You can follow Oracle's [Installing the CLI guide](https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm#Quickstart) to set up the OCI CLI.

`oshiv` uses the credentials stored in the `$HOME/.oci/config` file and the OCI profile specified by the `OCI_CLI_PROFILE` environment variable. If the `OCI_CLI_PROFILE` variable is not set, it defaults to the `DEFAULT` profile.

To set a custom OCI profile, use the following command:

```bash
export OCI_CLI_PROFILE=MYCUSTOMPROFILE
```

With these steps completed, you're ready to use `oshiv` for managing and connecting to OCI instances.

### 3. OCI Tenancy

`oshiv` will attempt to determine tenancy in this order:

1. Attempt to get tenancy ID from `OCI_CLI_TENANCY` environment variable. (E.g. `export OCI_CLI_TENANCY=ocid1.tenancy.oc2..`)

2. Attempt to get tenancy ID from `-t` flag

3. Attempt to get tenancy ID from OCI config file (`$HOME/.oci/config`)

Patterns `#1` and `#2` above allow you to override your default tenancy.

## Defaults

### SSH keys

By default, `oshiv` looks for the `id_rsa` and `id_rsa.pub` key pair in `$HOME/.ssh`. You can override the directory of your `id_rsa` and `id_rsa.pub` key pair by setting the `OSHIV_SSH_HOME` environment variable. 

Example: 

```
export OSHIV_SSH_HOME=$HOME/.ssh/oshiv
ssh-keygen -t rsa -b 4096 -C "oshiv@oraclecloud.com" -f $OSHIV_SSH_HOME/id_rsa
```

Individual keys can be overwritten by passing the the following flags:

- `-a`, `--private-key` Path to SSH private key (identity file)
- `-e`, `--public-key` Path to SSH public key

See `oshiv bastion -h` for more detail.

*Note: If you use `OSHIV_SSH_HOME` you'll want to add it to your ZSH init file.*

### SSH user

By default, `oshiv` uses the `opc` user. This can be overriden by flags. See `oshiv bastion -h`

### SSH port

By default, `oshiv` uses port `22` user. This can be overriden by flags. See `oshiv -h`

*Note: This is the port used to SSH to the bastion host and subsequently the target host. Not to be confused with the local/remote ports used for tunneling.*

## Info Command

The `info` command displays custom tenancy info that you define in your tenancy info file located at `$HOME/.oci/tenancy-map.yaml`. This is helpful to quickly display the tenancy and compartment info necessary to run most oshiv commands.

[tenancy-map.yaml example](../examples/tenancy-map.yml)

## Help (all options)

```
oshiv -h
```

```
A tool for finding and connecting to OCI resources via the bastion service

Usage:
  oshiv [flags]
  oshiv [command]

Available Commands:
  bastion     Find, list, and connect to resources via the OCI bastion service
  compartment Find and list compartments
  completion  Generate the autocompletion script for the specified shell
  config      Display oshiv configuration
  db          Find and list databases
  help        Help about any command
  image       Find and list OCI compute images
  info        Display your custom OCI tenancy information
  instance    Find and list OCI instances
  oke         Find and list OKE clusters
  policy      Find and list policies by name or statement
  subnet      Find and list subnets
  version     Print the version number of oshiv CLI

Flags:
  -c, --compartment string   The name of the compartment to use
  -h, --help                 help for oshiv
  -t, --tenancy-id string    Override's the default tenancy with this tenancy ID
  -v, --version              Print the version number of oshiv CLI
```