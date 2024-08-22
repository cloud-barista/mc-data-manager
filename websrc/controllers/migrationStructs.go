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
package controllers

import "mime/multipart"

// MigrationForm represents the form data required for migration processes.
// @Description MigrationForm contains all the necessary fields for migrating data between different services.
type MigrationForm struct {
	Path string `form:"path"`

	AWSRegion    string `form:"awsRegion" json:"awsRegion"`
	AWSAccessKey string `form:"awsAccessKey" json:"awsAccessKey"`
	AWSSecretKey string `form:"awsSecretKey" json:"awsSecretKey"`
	AWSBucket    string `form:"awsBucket" json:"awsBucket"`

	ProjectID     string                `form:"projectid" json:"projectid"`
	GCPRegion     string                `form:"gcpRegion" json:"gcpRegion"`
	GCPBucket     string                `form:"gcpBucket" json:"gcpBucket"`
	GCPCredential *multipart.FileHeader `form:"gcpCredential" json:"-" swaggerignore:"true"`

	NCPRegion    string `form:"ncpRegion" json:"ncpRegion"`
	NCPAccessKey string `form:"ncpAccessKey" json:"ncpAccessKey"`
	NCPSecretKey string `form:"ncpSecretKey" json:"ncpSecretKey"`
	NCPEndPoint  string `form:"ncpEndpoint" json:"ncpEndpoint"`
	NCPBucket    string `form:"ncpBucket" json:"ncpBucket"`

	MongoHost     string `form:"host" json:"host"`
	MongoPort     string `form:"port" json:"port"`
	MongoUsername string `form:"username" json:"username"`
	MongoPassword string `form:"password" json:"password"`
	MongoDBName   string `form:"databaseName" json:"databaseName"`
}
type GCPMigrationParams struct {
	ProjectID     string                `form:"projectid" json:"projectid"`
	GCPRegion     string                `form:"gcpRegion" json:"gcpRegion"`
	GCPBucket     string                `form:"gcpBucket" json:"gcpBucket"`
	GCPCredential *multipart.FileHeader `form:"gcpCredential" json:"-" swaggerignore:"true"`
}

type MigrationMySQLParams struct {
	Source MySQLParams
	Dest   MySQLParams
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

type MySQLParams struct {
	Provider     string `json:"Provider"`
	Host         string `json:"Host"`
	Port         string `json:"Port"`
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	DatabaseName string `json:"DatabaseName"`
}

type MongoDBParams struct {
	Host         string `json:"Host"`
	Port         string `json:"Port"`
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	DatabaseName string `json:"DatabaseName"`
}

func GetMigrationParamsFormFormData(form MigrationMySQLForm) MigrationMySQLParams {
	src := MySQLParams{
		Provider:     form.SProvider,
		Host:         form.SHost,
		Port:         form.SPort,
		Username:     form.SUsername,
		Password:     form.SPassword,
		DatabaseName: form.SDatabaseName,
	}
	dest := MySQLParams{
		Provider:     form.DProvider,
		Host:         form.DHost,
		Port:         form.DPort,
		Username:     form.DUsername,
		Password:     form.DPassword,
		DatabaseName: form.DDatabaseName,
	}
	return MigrationMySQLParams{Source: src, Dest: dest}
}
