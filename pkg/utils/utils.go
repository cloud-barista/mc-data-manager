/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"os"
	"time"
)

type Object struct {
	ChecksumAlgorithm []string
	ETag              string
	Key               string
	LastModified      time.Time
	Size              int64
	StorageClass      string
}

type Provider string

const (
	AWS Provider = "aws"
	GCP Provider = "gcp"
	NCP Provider = "ncp"
	OPM Provider = "on-premise"
)

// Distinguish between directory and file or directory
func IsDir(path string) error {
	fInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if !fInfo.IsDir() {
			return err
		}
	}
	return nil
}

func FileExists(filePath string) bool {
	if fi, err := os.Stat(filePath); os.IsExist(err) {
		return !fi.IsDir()
	}
	return false
}
