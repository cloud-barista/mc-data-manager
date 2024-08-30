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

type BaseParams struct {
	OperationParams
	ProviderParams
	RegionParams
	ProfileParams
}

type OperationParams struct {
	OperationId string `json:"operationId" form:"operationId"`
}

type ProviderParams struct {
	Provider string `json:"provider" form:"provider"`
}

type RegionParams struct {
	Region string `json:"region" form:"region"`
}

type ProfileParams struct {
	ProfileName string `json:"profileName" form:"profileName"`
}

type MySQLParams struct {
	Provider     string `json:"provider" form:"provider"`
	Host         string `json:"host" form:"host"`
	Port         string `json:"port" form:"port"`
	User         string `json:"username" form:"username"`
	Password     string `json:"password" form:"password"`
	DatabaseName string `json:"databaseName" form:"databaseName"`
}

type ObjectStorageParams struct {
	Bucket   string `json:"bucket" form:"bucket"`
	Endpoint string `json:"endpoint" form:"endpoint"`
}

type FileFormatParams struct {
	CheckSQL        string `json:"checkSQL" form:"checkSQL"`
	CheckCSV        string `json:"checkCSV" form:"checkCSV"`
	CheckTXT        string `json:"checkTXT" form:"checkTXT"`
	CheckPNG        string `json:"checkPNG" form:"checkPNG"`
	CheckGIF        string `json:"checkGIF" form:"checkGIF"`
	CheckZIP        string `json:"checkZIP" form:"checkZIP"`
	CheckJSON       string `json:"checkJSON" form:"checkJSON"`
	CheckXML        string `json:"checkXML" form:"checkXML"`
	CheckServerJSON string `json:"checkServerJSON" form:"checkServerJSON"`
	CheckServerSQL  string `json:"checkServerSQL" form:"checkServerSQL"`
}

type FileSizeParams struct {
	SizeSQL        string `json:"sizeSQL" form:"sizeSQL"`
	SizeCSV        string `json:"sizeCSV" form:"sizeCSV"`
	SizeTXT        string `json:"sizeTXT" form:"sizeTXT"`
	SizePNG        string `json:"sizePNG" form:"sizePNG"`
	SizeGIF        string `json:"sizeGIF" form:"sizeGIF"`
	SizeZIP        string `json:"sizeZIP" form:"sizeZIP"`
	SizeJSON       string `json:"sizeJSON" form:"sizeJSON"`
	SizeXML        string `json:"sizeXML" form:"sizeXML"`
	SizeServerJSON string `json:"sizeServerJSON" form:"sizeServerJSON"`
	SizeServerSQL  string `json:"sizeServerSQL" form:"sizeServerSQL"`
}
