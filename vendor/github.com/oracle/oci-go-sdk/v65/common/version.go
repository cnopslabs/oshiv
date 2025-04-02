// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated by go generate; DO NOT EDIT

package common

import (
	"bytes"
	"fmt"
	"sync"
)

const (
	major = "65"
	minor = "88"
	patch = "1"
	tag   = ""
)

var once sync.Once
var version string

// Version returns semantic version of the sdk
func Version() string {
	once.Do(func() {
		ver := fmt.Sprintf("%s.%s.%s", major, minor, patch)
		verBuilder := bytes.NewBufferString(ver)
		if tag != "" && tag != "-" {
			_, err := verBuilder.WriteString(tag)
			if err != nil {
				verBuilder = bytes.NewBufferString(ver)
			}
		}
		version = verBuilder.String()
	})
	return version
}
