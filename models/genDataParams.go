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
	BaseParams
	AccessKey string `json:"accessKey" form:"accessKey"`
	SecretKey string `json:"secretKey" form:"secretKey"`

	ObjectStorageParams

	GenFileParams

	MySQLParams

	FirestoreParams
}

type GenFileParams struct {
	DummyPath string `json:"path" form:"path"`
	FileFormatParams
	FileSizeParams
}

type GenMySQLParams struct {
	BaseParams
	MySQLParams
}
