// Home dir is used for reading OCI config files and SSH keys
package utils

import "os"

func HomeDir() string {
	homeDir, err := os.UserHomeDir()
	CheckError(err)

	return homeDir
}
