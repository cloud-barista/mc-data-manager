package nrdbc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/awsdnmdb"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/gcpfsdb"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/ncpmgdb"
	"github.com/cloud-barista/cm-data-mold/service/nrdbc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

// dynamo to firestore example
func TestMain(m *testing.M) {
	awsnrdbc, err := AWSInfo("your-aws-accessKey", "your-aws-secretKey", "your-aws-reigon")
	if err != nil {
		panic(err)
	}

	gcpnrdc, err := GCPInfo("your-gcp-projectID", "your-gcp-credentialsFile")
	if err != nil {
		panic(err)
	}

	var srcData []map[string]interface{}
	err = json.Unmarshal([]byte(exJSON), &srcData)
	if err != nil {
		panic(err)
	}

	if err := awsnrdbc.Put("address", &srcData); err != nil {
		panic(err)
	}

	var dstData []map[string]interface{}
	if err := awsnrdbc.Get("address", &dstData); err != nil {
		panic(err)
	}

	// dynamo to firestore
	if err := awsnrdbc.Copy(gcpnrdc); err != nil {
		panic(err)
	}
}

func AWSInfo(accessKey, secretKey, region string) (*nrdbc.NRDBController, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
		config.WithRetryMaxAttempts(5),
	)

	if err != nil {
		return nil, err
	}

	return nrdbc.New(awsdnmdb.New(dynamodb.NewFromConfig(cfg), region))
}

func GCPInfo(projectID, credentialsFile string) (*nrdbc.NRDBController, error) {
	client, err := firestore.NewClient(
		context.TODO(),
		projectID,
		option.WithCredentialsFile(credentialsFile),
	)
	if err != nil {
		return nil, err
	}

	return nrdbc.New(gcpfsdb.New(client, "your-aws-region"))
}

func NCPInfo(username, password, host, port, dbName string) (*nrdbc.NRDBController, error) {
	dc := true
	mongoClient, err := mongo.Connect(context.Background(), &options.ClientOptions{
		Auth: &options.Credential{
			Username: username,
			Password: password,
		},
		Direct: &dc,
		Hosts:  []string{fmt.Sprintf("%s:%s", host, port)},
	})

	if err != nil {
		return nil, err
	}

	return nrdbc.New(ncpmgdb.New(mongoClient, dbName))
}

const exJSON string = `
[
    {
        "addr_id": "DhiFnPGUcQUL8wgSkx1ky1hb",
        "countryabr": "SE",
        "street": "8350 West Loopbury",
        "city": "St. Petersburg",
        "state": "Idaho",
        "zip": "58295",
        "country": "Brazil",
        "latitude": 43,
        "longitude": -11
    },
    {
        "addr_id": "Hf1ct5b8OCySK3OxiUET1nQ4",
        "countryabr": "IN",
        "street": "6258 New Viaberg",
        "city": "Boston",
        "state": "Mississippi",
        "zip": "70833",
        "country": "Lebanon",
        "latitude": -42,
        "longitude": 70
    },
    {
        "addr_id": "s3suVD7d1hFUUz4ci6Z1TMAb",
        "countryabr": "PL",
        "street": "2497 Squarestad",
        "city": "Fort Wayne",
        "state": "Missouri",
        "zip": "50560",
        "country": "Hong Kong",
        "latitude": 37,
        "longitude": 123
    }
]`
