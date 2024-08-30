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
package auth

import "github.com/cloud-barista/mc-data-manager/models"

type DatamoldParams struct {
	models.BaseProfile
	// credential
	CredentialPath string
	ConfigData     map[string]map[string]map[string]string
	TaskTarget     bool

	//src
	SrcProvider ProviderConfig

	//dst
	DstProvider ProviderConfig

	// dummy
	DstPath  string
	SqlSize  int
	CsvSize  int
	JsonSize int
	XmlSize  int
	TxtSize  int
	PngSize  int
	GifSize  int
	ZipSize  int

	DeleteDBList    []string
	DeleteTableList []string
}

type ProviderConfig struct {
	// common
	Provider string
	Region   string
	ObjectStorageParams
	MySQLParams
	GcpNosqlParams
	CredParams
}

type ProviderConfigSet struct {
	// common
	Provider string
	Region   string
	ObjectStorageParams
	MySQLParams
	GcpNosqlParams
}

type CredParams struct {
	AccessKey   string
	SecretKey   string
	GcpCredPath string
	GcpCredJson string
}

type ObjectStorageParams struct {
	BucketName string
	Endpoint   string
}

type MySQLParams struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
}

type GcpNosqlParams struct {
	ProjectID  string
	DatabaseID string
}
