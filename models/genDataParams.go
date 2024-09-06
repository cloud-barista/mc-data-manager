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
package models

type GenDataParams struct {
	OperationParams
	BaseParams

	ObjectStorageParams

	GenFileParams

	MySQLParams

	FirestoreParams
}

type GenFileParams struct {
	Directory string `json:"Directory,omitempty" swaggerignore:"true"`
	DummyPath string `json:"dummyPath,omitempty" swaggerignore:"true"`
	FileFormatParams
	FileSizeParams
}

type GenMySQLParams struct {
	BaseParams
	MySQLParams
}

type APICredentialParams struct {
	GCPCredentialJson string
	GCPCredentialPath string
	AWSAccessKey      string
	AWSSecretKey      string
	NCPAccessKey      string
	NCPSecretKey      string
}
