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

import "time"

type BaseParams struct {
	ProviderParams
	RegionParams
	// ProfileParams
	CredentialParams
}

type ProviderParams struct {
	Provider string `json:"provider" form:"provider"`
}

type RegionParams struct {
	Region string `json:"region" form:"region"`
}

// type ProfileParams struct {
// 	ProfileName string `json:"profileName" form:"profileName"`
// }

type CredentialParams struct {
	CredentialId int64 `json:"credentialId" form:"credentialId"`
}

type MySQLParams struct {
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
	CheckSQL        bool `json:"checkSQL" form:"checkSQL"`
	CheckCSV        bool `json:"checkCSV" form:"checkCSV"`
	CheckTXT        bool `json:"checkTXT" form:"checkTXT"`
	CheckPNG        bool `json:"checkPNG" form:"checkPNG"`
	CheckGIF        bool `json:"checkGIF" form:"checkGIF"`
	CheckZIP        bool `json:"checkZIP" form:"checkZIP"`
	CheckJSON       bool `json:"checkJSON" form:"checkJSON"`
	CheckXML        bool `json:"checkXML" form:"checkXML"`
	CheckServerJSON bool `json:"checkServerJSON" form:"checkServerJSON"`
	CheckServerSQL  bool `json:"checkServerSQL" form:"checkServerSQL"`
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

type MongoMigrationParams struct {
	MongoHost     string `form:"host" json:"host"`
	MongoPort     string `form:"port" json:"port"`
	MongoUsername string `form:"username" json:"username"`
	MongoPassword string `form:"password" json:"password"`
	MongoDBName   string `form:"databaseName" json:"databaseName"`
}

type MigrationMySQLForm struct {
	SProvider     string `json:"srcProvider" form:"srcProvider"`
	SHost         string `json:"srcHost" form:"srcHost"`
	SPort         string `json:"srcPort" form:"srcPort"`
	SUsername     string `json:"srcUsername" form:"srcUsername"`
	SPassword     string `json:"srcPassword" form:"srcPassword"`
	SDatabaseName string `json:"srcDatabaseName" form:"srcDatabaseName"`

	DProvider     string `json:"destProvider" form:"destProvider"`
	DHost         string `json:"destHost" form:"destHost"`
	DPort         string `json:"destPort" form:"destPort"`
	DUsername     string `json:"destUsername" form:"destUsername"`
	DPassword     string `json:"destPassword" form:"destPassword"`
	DDatabaseName string `json:"destDatabaseName" form:"destDatabaseName"`
}

type MigrationMySQLParams struct {
	SourcePoint MySQLParams
	TargetPoint MySQLParams
}

type NoSQLParams struct {
	GcpNosqlParams
}

type GcpNosqlParams struct {
	DatabaseID string `json:"databaseId" form:"databaseId"`
	ProjectID  string `json:"projectId" form:"projectId"`
}

type CredParams struct {
	AccessKey   string
	SecretKey   string
	GcpCredPath string
	GcpCredJson string
}

type FirestoreParams struct {
	DatabaseID string `json:"databaseId" form:"databaseId"`
	ProjectID  string `json:"projectId" form:"projectId"`
}

// object
type Object struct {
	ChecksumAlgorithm []string
	ETag              string
	Key               string
	LastModified      time.Time
	Size              int64
	StorageClass      string
	Provider
}

type ServiceType struct {
	Type CloudServiceType `json:"type" form:"type"` // The type of cloud service
}
