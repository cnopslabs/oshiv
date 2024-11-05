# Refactor by OCI API - Use packages

Generally follow OCI API to break components into packages. Create new packages for non-API components (E.g. ssh).

## Structure

```
├── README.md
├── bastion
│   └── bastion.go
├── cluster
│   └── cluster.go
├── cmd
│   ├── bastion.go
│   ├── compartment.go
│   ├── image.go
│   ├── instance.go
│   ├── policy.go
│   ├── root.go
│   ├── session.go (a sub-command of bastion)
│   └── subnet.go
├── compute
│   ├── image.go
│   └── instance.go
├── go.mod
├── go.sum
├── identity
│   ├── compartment.go
│   └── policy.go
├── main.go
├── network
│   ├── subnets.go
│   └── vnic.go
└── ssh
    └── commands.go
```
