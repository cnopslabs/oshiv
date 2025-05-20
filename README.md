# OCI Shiv
[![CI Build](https://github.com/cnopslabs/oshiv/actions/workflows/build.yml/badge.svg)](https://github.com/cnopslabs/oshiv/actions/workflows/build.yml)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cnopslabs/oshiv?sort=semver)
[![Version](https://img.shields.io/badge/goversion-1.23.x-blue.svg)](https://golang.org)
<a href="https://golang.org"><img src="https://img.shields.io/badge/powered_by-Go-3362c2.svg?style=flat-square" alt="Built with GoLang"></a>
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/cnopslabs/oshiv/main/LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/cnopslabs/oshiv)](https://goreportcard.com/report/github.com/cnopslabs/oshiv)

A tool for quickly finding and connecting to OCI instances, OKE clusters, and autonomous databases via the bastion service.

## Overview

oshiv is a command-line tool that simplifies working with Oracle Cloud Infrastructure (OCI) resources. It helps you:

- Find and list OCI resources (instances, clusters, databases, etc.)
- Connect to resources via the OCI bastion service
- Manage SSH connections and tunneling to OCI resources

## Quick Examples

**Finding and connecting to OCI instances**

```
# Search for instances
oshiv inst -f foo-node

# Connect via bastion service
oshiv bastion -i 123.456.789.5 -o ocid1.instance.oc2.us-luke-1.abcdefghijklmnopqrstuvwxyz
```

**Finding and connecting to Kubernetes clusters**

```
# Search for clusters
oshiv oke -f foo-cluster

# Connect via bastion service
oshiv bastion -y port-forward -k oke-my-foo-cluster -i 123.456.789.7
```

## Documentation

For detailed documentation, please refer to the following guides:

- [Installation Guide](docs/installation.md) - How to install oshiv
- [Usage Guide](docs/usage.md) - Prerequisites, authentication, and basic usage
- [Examples](docs/examples.md) - Common usage patterns and tunneling examples
- [Contributing](docs/contributing.md) - How to contribute to the project
- [Future Enhancements](docs/future.md) - Planned features and improvements
- [Reference](docs/reference.md) - Additional resources and links

## Contributing

Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on submitting Enhancement Proposals and Pull Requests.

## License

This project is licensed under the MIT License - see the [LICENSE.md](./LICENSE.md) file for details.
