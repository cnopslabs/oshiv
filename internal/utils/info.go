package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"gopkg.in/yaml.v2"
)

type OciTenancyEnvironment struct {
	Environment  string `yaml:"environment"`
	Tenancy      string `yaml:"tenancy"`
	TenancyId    string `yaml:"tenancy_id"`
	Realm        string `yaml:"realm"`
	Compartments string `yaml:"compartments"`
	Regions      string `yaml:"regions"`
}

var tenancyMapPath string = filepath.Join(HomeDir(), ".oci", "tenancy-map.yaml")

func PrintTenancyMap() {
	var ociTenancyEnvironments []OciTenancyEnvironment
	_, err_stat := os.Stat(tenancyMapPath)

	if err_stat == nil {
		yamlFile, err := os.ReadFile(tenancyMapPath)
		CheckError(err)

		err = yaml.Unmarshal(yamlFile, &ociTenancyEnvironments)
		CheckError(err)

		tbl := table.New("ENVIRONMENT", "TENANCY", "REALM", "COMPARTMENTS", "REGIONS")
		tbl.WithHeaderFormatter(HeaderFmt).WithFirstColumnFormatter(ColumnFmt)

		for _, env := range ociTenancyEnvironments {
			tbl.AddRow(env.Environment, env.Tenancy, env.Realm, env.Compartments, env.Regions)
		}

		tbl.Print()

		fmt.Println("\nTo set Tenancy, Compartment, or Region export the " + color.YellowString("OCI_TENANCY_NAME") + ", " + color.YellowString("OCI_COMPARTMENT") + ", or " + color.YellowString("OCI_CLI_REGION") + " environment variables.")
		fmt.Println("\nOr if using oshell, run:")
		fmt.Print("oci_set_tenancy ")
		Yellow.Println("TENANCY_NAME ")
		fmt.Print("oci_set_tenancy ")
		Yellow.Println("TENANCY_NAME COMPARTMENT_NAME")
	} else {
		fmt.Println("No tenancy info file found.")
	}
}

func LookUpTenancyID(tenancyName string) (string, error) {
	var ociTenancyEnvironments []OciTenancyEnvironment
	_, err_stat := os.Stat(tenancyMapPath)

	if err_stat == nil {
		yamlFile, err := os.ReadFile(tenancyMapPath)
		CheckError(err)

		err = yaml.Unmarshal(yamlFile, &ociTenancyEnvironments)
		CheckError(err)

		for _, env := range ociTenancyEnvironments {
			if tenancyName == env.Tenancy {
				return env.TenancyId, nil
			}
		}

		return "", errors.New("tenancy not found")
	} else {
		return "", errors.New("no tenancy info file found")
	}
}
