# oshiv structure

Module structure follows the "[Packages and commands in the same repository](https://go.dev/doc/modules/layout#package-or-command-with-supporting-packages)" convention. OCI resources are placed in the `./internal/resources` directory and [Viper commands](https://github.com/spf13/viper) are placed in the `./cmd` directory. 

The `./website` directory does not contain Go code, but provides the location for the download web site content.

## Structure

```
├── cmd
│   ├── bastion.go
│   ├── compartment.go
│   ├── config.go
│   ├── image.go
│   ├── instance.go
│   ├── oke.go
│   ├── policy.go
│   ├── root.go
│   ├── session.go
│   └── subnet.go
├── go.mod
├── go.sum
├── internal
│   ├── resources
│   │   ├── bastion.go
│   │   ├── compartment.go
│   │   ├── image.go
│   │   ├── instance.go
│   │   ├── oke.go
│   │   ├── policy.go
│   │   ├── subnet.go
│   │   └── tenancy.go
│   └── utils
│       ├── config.go
│       ├── errors.go
│       ├── home_dir.go
│       ├── logger.go
│       ├── oci_config.go
│       └── print.go
├── main.go
└── website
    └── oshiv
        ├── index.tmpl
        └── renderhtml.go
```
