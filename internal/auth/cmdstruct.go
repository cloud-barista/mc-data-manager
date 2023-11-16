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
