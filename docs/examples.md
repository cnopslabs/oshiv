# Usage Examples

## Quick Examples

### Finding and connecting to OCI instances

Search for instances:

```
oshiv inst -f foo-node
```
```
Name: my-foo-node-1
Instance ID: ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.5

Name: my-foo-node-2
Instance ID: ocid1.instance.oc2.us-luke-1.bacdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.6
```

Connect via bastion service:

```
oshiv bastion -i 123.456.789.5 -o ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
```

### Finding and connecting to Kubernetes clusters

Search for clusters:

```
oshiv oke -f foo-cluster
```
```
Name: oke-my-foo-cluster
Cluster ID: ocid1.cluster.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
Private endpoint IP: 123.456.789.7
Private endpoint port: 6443
```

Connect via bastion service:

```
oshiv bastion -y port-forward -k oke-my-foo-cluster -i 123.456.789.7
```

## Common Usage Patterns

### Compartments

List compartments:

```
oshiv compart -l
```

<details>
<summary>Output</summary>

```
COMPARTMENTS:
fakecompartment1
dummycompartment2
mycompartment
```

</details>

### Instances

Find instance:

```
oshiv inst -f foo-app
```

<details>
<summary>Output</summary>

```
Name: my-foo-app-1
Instance ID: ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.5

Name: my-foo-app-2
Instance ID: ocid1.instance.oc2.us-luke-1.bacdefghijklmnopqrstuvwxyz
Private IP: 123.456.789.6
```

</details>

<br>

Create bastion session to connect to instance:

```
oshiv inst -i 123.456.789.5 -o ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
```

Connect to instance:

`oshiv` will produce various SSH commands to connect to your instance

<details>
<summary>Output</summary>

```
Tunnel:
sudo ssh -i "/Users/myuser/.ssh/id_rsa" \
-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i "/Users/myuser/.ssh/id_rsa" -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5 -N -L <LOCAL PORT>:123.456.789.5:<REMOTE PORT>

SCP:
scp -i /Users/myuser/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P 22 \
-o ProxyCommand='ssh -i /Users/myuser/.ssh/id_rsa -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
<SOURCE PATH> opc@123.456.789.5:<TARGET PATH>

SSH:
ssh -i /Users/myuser/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i /Users/myuser/.ssh/id_rsa -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5
```

</details>

### OKE Kubernetes clusters

Find OKE cluster and create bastion session to connect to the Kubernetes API:

```
oshiv oke -f oke-my-foo-cluster
```

```
oshiv bastion -y port-forward -k oke-my-foo-cluster -i 123.456.789.7
```

Connect to cluster:

`oshiv` will produce an SSH command to allow port forwarding connectivity to your cluster. It will also produce an oci cli commands to update your Kubernetes config file with the OKE cluster details (this only needs to be performed once).

```
Update kube config (One time operation):
oci ce cluster create-kubeconfig --cluster-id ocid1.cluster.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz --token-version 2.0.0 --kube-endpoint 123.456.789.7

Port Forwarding command:
ssh -i /Users/myuser/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-p 22 -N -L 6443:123.456.789.7:6443 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com
```

You should now be able to connect to your cluster's API endpoint using tools like `kubectl` and `k9s`.

## Tunneling Examples

### VNC (Linux GUI)

```
sudo ssh -i "/Users/myuser/.ssh/id_rsa" \
-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i "/Users/myuser/.ssh/id_rsa" -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5 -N -L 5902:123.456.789.5:5902
```

Now you should be able to connect (via localhost) using a VNC client. 

For MacOS, I recommend [TigerVNC](https://tigervnc.org/) but the built-in VNC client will work as well.

```
localhost:5902
```

Tiger VNC

<img src="../tiger-vnc.png" width="400"/>

MacOS VNC

<img src="../mac-vnc-connect-to-server.jpg" width="400"/>

### RDP (Windows)

```
sudo ssh -i "/Users/myuser/.ssh/id_rsa" \
-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
-o ProxyCommand='ssh -i "/Users/myuser/.ssh/id_rsa" -W %h:%p -p 22 ocid1.bastionsession.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz@host.bastion.us-luke-1.oci.oraclegovcloud.com' \
-P 22 opc@123.456.789.5 -N -L 3389:123.456.789.5:3389
```

Now you should be able to connect via localhost with an RDP client.