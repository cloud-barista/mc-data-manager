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

type DatamoldParams struct {
	// credential
	CredentialPath string
	ConfigData     map[string]map[string]map[string]string
	TaskTarget     bool

	//src
	SrcProvider    string
	SrcAccessKey   string
	SrcSecretKey   string
	SrcRegion      string
	SrcBucketName  string
	SrcGcpCredPath string
	SrcProjectID   string
	SrcEndpoint    string
	SrcUsername    string
	SrcPassword    string
	SrcHost        string
	SrcPort        string
	SrcDBName      string

	//dst
	DstProvider    string
	DstAccessKey   string
	DstSecretKey   string
	DstRegion      string
	DstBucketName  string
	DstGcpCredPath string
	DstProjectID   string
	DstEndpoint    string
	DstUsername    string
	DstPassword    string
	DstHost        string
	DstPort        string
	DstDBName      string

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
