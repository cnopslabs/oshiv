# oshiv
A tool for finding and connecting to OCI instances via the OCI bastion service.

**Quick example**

Find instance(s)

```
oshiv -f foo-node
```
```
Name: my-foo-node-1
Instance ID: ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.5

Name: my-foo-node-2
Instance ID: ocid1.instance.oc2.us-luke-1.bacdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.6
```

Connect via bastion service

```
oshiv -i 123.456.789.5 -o ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
```

## Install

### Download

Download binary from [https://www.daniel-lloyd.net/oshiv/index.html](https://www.daniel-lloyd.net/oshiv/index.html).

### Place in PATH

*MacOS*

Place binary in `/usr/local/bin` or other location in your `PATH`.

Example "other"

```
echo $PATH

/Users/fakeuser/.pyenv/shims:/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/Users/fakeuser/.local/bin
```

```
mv ~/Downloads/oshiv  /Users/fakeuser/.local/bin

oshiv -h
```

*Windows*

Go to Control Panel -> System -> System settings -> Environment Variables.

Scroll down in system variables until you find `PATH`.

Click edit and the location of your binary. For example, `c:\oshiv`.

*Note: Be sure to include a semicolon at the end of the previous as that is the delimiter, i.e. `c:\path;c:\oshiv`.*

Launch a new console for the settings to take effect.

### Verify

```
oshiv -h
```

## Usage

### Prerequisites

#### 1. OCI Authentication

This tool will use the credentials set in `$HOME/.oci/config`
This tool will use the profile set by the `OCI_CLI_PROFILE` environment variable
If the `OCI_CLI_PROFILE` environment variable is not set it will use the DEFAULT profile

*(optional)*
```
export OCI_CLI_PROFILE=MYCUSTOMPROFILE
```

#### 2. OCI Tenancy
OCI Tenancy must be set by environment variable (below) or by passing it in with the flag: `-t ocid1.tenancy.oc2..aaaaaaaaabcdefghijklmnopqrstuvwxyz` 

```
export OCI_CLI_TENANCY=ocid1.tenancy.oc2..aaaaaaaaabcdefghijklmnopqrstuvwxyz
```

### Defaults

#### 1. SSH keys

By default, `oshiv` uses the following keys: 

- `$HOME/.ssh/id_rsa`
- `$HOME/.ssh/id_rsa.pub`

These can be overriden by flags. See `oshiv -h`

#### 2. SSH user

By default, `oshiv` uses the `opc` user. This can be overriden by flags. See `oshiv -h`

#### 3. SSH port

By default, `oshiv` uses port `22` user. This can be overriden by flags. See `oshiv -h`

*Note: This is the port used to SSH to the bastion host and subsequently the target host. Not to be confused with the local/remote ports used for tunneling.*

### Common usage pattern

1. List compartments for tenancy

```
oshiv -lc

COMPARTMENTS:
fakecompartment1
dummycompartment2
mycompartment

To set compartment, you can export OCI_COMPARTMENT_NAME:
   export OCI_COMPARTMENT_NAME=
```

2. Set compartment

```
export OCI_COMPARTMENT_NAME=mycompartment
```

3. Find instance(s)

```
oshiv -f mydatabase

Name: mydatabase-1
Instance ID: ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.5

Name: mydatabase-2
Instance ID: ocid1.instance.oc2.us-luke-1.bacdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.6
```

4. List bastions

```
oshiv -lb

Bastions in compartment mycompartment
mybastion-1

To set bastion name, you can export OCI_BASTION_NAME:
   export OCI_BASTION_NAME=
```

5. Set bastion

```
export OCI_BASTION_NAME=mybastion-1
```

6. Create bastion session

```
oshiv -i 123.456.789.5 -o ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
```

7. Connect to instance

`oshiv` will produce various SSH commands to connect to your instance

```
Tunnel:
sudo ssh -i "/Users/myuser/.ssh/id_rsa" \
-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i "/Users/myuser/.ssh/id_rsa" -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5 -N -L<LOCAL PORT>:123.456.789.5:<REMOTE PORT>

SCP:
scp -i /Users/myuser/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P 22 \
-o ProxyCommand='ssh -i /Users/myuser/.ssh/id_rsa -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
<SOURCE PATH> opc@123.456.789.5:<TARGET PATH>

SSH:
ssh -i /Users/myuser/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i /Users/myuser/.ssh/id_rsa -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5
```

### Tunneling examples

#### VNC (Linux GUI)

```
sudo ssh -i "/Users/myuser/.ssh/id_rsa" \
-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i "/Users/myuser/.ssh/id_rsa" -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5 -N -L 5902:123.456.789.5:5902
```

Now you should be able to connect via localhost with a VNC client

```
vnc://localhost:5902
```

![mac-vnc-connect-to-server.jpg](mac-vnc-connect-to-server.jpg)
#### RDP (Windows)

```
sudo ssh -i "/Users/myuser/.ssh/id_rsa" \
-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i "/Users/myuser/.ssh/id_rsa" -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5 -N -L 3389:123.456.789.5:3389
```

Now you should be able to connect via localhost with an RDP client


### Help (all options)

```
oshiv -h
```

```
OCI authentication:
This tool will use the credentials set in $HOME/.oci/config
This tool will use the profile set by the OCI_CLI_PROFILE environment variable
If the OCI_CLI_PROFILE environment variable is not set it will use the DEFAULT profile

Environment variables:
The following environment variables will override their flag counterparts
   OCI_CLI_TENANCY
   OCI_COMPARTMENT_NAME
   OCI_BASTION_NAME

Defaults:
   SSH private key (-k): $HOME/.ssh/id_rsa
   SSH public key (-e): $HOME/.ssh/id_rsa.pub
   SSH user (-u): opc

Common command patterns:
List compartments
   oshiv -lc

List bastions
   oshiv -lb

Create bastion session
   oshiv -i ip_address -o instance_id

List active sessions
   oshiv -ls

Connect to an active session
   oshiv -s session_ocd

Create bastion session (all flags)
   oshiv -t tenant_id -c compartment_name -b bastion_name -i ip_address -o instance_id -k path_to_ssh_private_key -e path_to_ssh_public_key -u cloud-user

All flags for oshiv:
  -b string
    	bastion name
  -c string
    	compartment name
  -e string
    	path to SSH public key
  -f string
    	search string to search for instance
  -i string
    	instance IP address of host to connect to
  -k string
    	path to SSH private key (identity file)
  -l int
    	Session TTL (seconds) (default 10800)
  -lb
    	list bastions
  -lc
    	list compartments
  -ls
    	list sessions
  -o string
    	instance ID of host to connect to
  -p int
    	SSH port (default 22)
  -s string
    	Session ID to check for
  -t string
    	tenancy ID name
  -tp int
    	SSH Tunnel port
  -u string
    	SSH user (default "opc")
  -w	Create an SSH port forward session
```

## Contribute

Style guide: https://go.dev/doc/effective_go

```
git clone https://github.com/dnlloyd/oshiv
```

### Build local OS/Arch

```
go build
```

### Build OS/Arch specific

```
GOOS=darwin GOARCH=amd64 go build -o executables/mac/intel/oshiv
GOOS=darwin GOARCH=arm64 go build -o executables/mac/arm/oshiv
GOOS=windows GOARCH=amd64 go build -o executables/windows/intel/oshiv
GOOS=windows GOARCH=arm64 go build -o executables/windows/arm/oshiv
GOOS=linux GOARCH=amd64 go build -o executables/linux/intel/oshiv
GOOS=linux GOARCH=arm64 go build -o executables/linux/arm/oshiv
```

### Local install

```
go install
```

### Test and push

Test/validate changes, push to your fork, make PR

## Future enhancements and updates

- Add tests!
- Add search capability for NSG rules
- Generate and use ephemeral SSH keys
- Switch to more mature cmd line flag parsing library
- Use logging library
- When creating a bastion session, only require IP address or instance ID (and lookup the other)
- During session creation, find bastion automatically, if only one exists use it, else prompt user
- Manage SSH client
  - https://pkg.go.dev/golang.org/x/crypto/ssh
- Manage SSH keys
  - https://pkg.go.dev/crypto#PrivateKey

## Troubleshooting

```
xattr -d com.apple.quarantine /Users/dan/.local/bin/oshiv
```