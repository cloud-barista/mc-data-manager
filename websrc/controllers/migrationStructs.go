package controllers

import "mime/multipart"

type MigrationForm struct {
	Path string `form:"path"`

	AWSRegion    string `form:"awsRegion"`
	AWSAccessKey string `form:"awsAccessKey"`
	AWSSecretKey string `form:"awsSecretKey"`
	AWSBucket    string `form:"awsBucket"`

	ProjectID     string                `form:"projectid"`
	GCPRegion     string                `form:"gcpRegion"`
	GCPBucket     string                `form:"gcpBucket"`
	GCPCredential *multipart.FileHeader `form:"gcpCredential"`

	NCPRegion    string `form:"ncpRegion"`
	NCPAccessKey string `form:"ncpAccessKey"`
	NCPSecretKey string `form:"ncpSecretKey"`
	NCPEndPoint  string `form:"ncpEndpoint"`
	NCPBucket    string `form:"ncpBucket"`

	MongoHost     string `form:"host"`
	MongoPort     string `form:"port"`
	MongoUsername string `form:"username"`
	MongoPassword string `form:"password"`
	MongoDBName   string `form:"databaseName"`
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
	DHost         string `json:"destHost" form:"desttHost"`
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
