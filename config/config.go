package config

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

func validateInputs(username, password, host *string, port *int) error {
	if username == nil || password == nil || host == nil || port == nil {
		return errors.New("The input is invalid")
	}
	return nil
}

func newAWSConfig(accesskey, secretkey, region string) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accesskey, secretkey, "")),
		config.WithRegion(region),
		config.WithRetryMaxAttempts(5),
	)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func newAWSConfigWithEndpoint(serviceID, accesskey, secretkey, region, endpoint string) (*aws.Config, error) {
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == serviceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		}

		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accesskey, secretkey, "")),
		config.WithRegion(region),
		config.WithRetryMaxAttempts(5),
		config.WithEndpointResolver(customResolver),
	)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func newNCPMongoDBConfig(username, password, host string, port int) *options.ClientOptions {
	dc := true
	return &options.ClientOptions{
		Auth: &options.Credential{
			Username: username,
			Password: password,
		},
		Direct: &dc,
		Hosts:  []string{fmt.Sprintf("%s:%d", host, port)},
	}
}

func NewNCPMongoDBClient(username, password, host string, port int) (*mongo.Client, error) {
	if err := validateInputs(&username, &password, &host, &port); err != nil {
		return nil, err
	}
	return mongo.Connect(context.Background(), newNCPMongoDBConfig(username, password, host, port))
}

func NewS3Client(accesskey, secretkey, region string) (*s3.Client, error) {
	cfg, err := newAWSConfig(accesskey, secretkey, region)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func NewS3ClientWithEndpoint(accesskey, secretkey, region string, endpoint string) (*s3.Client, error) {
	cfg, err := newAWSConfigWithEndpoint(s3.ServiceID, accesskey, secretkey, region, endpoint)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func NewDynamoDBClient(accesskey, secretkey, region string) (*dynamodb.Client, error) {
	cfg, err := newAWSConfig(accesskey, secretkey, region)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(*cfg), nil
}

func NewDynamoDBClientWithEndpoint(accesskey, secretkey, region string, endpoint string) (*dynamodb.Client, error) {
	cfg, err := newAWSConfigWithEndpoint(dynamodb.ServiceID, accesskey, secretkey, region, endpoint)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(*cfg), nil
}

func NewGCSClient(credentialsFile string) (*storage.Client, error) {
	client, err := storage.NewClient(
		context.TODO(),
		option.WithCredentialsFile(credentialsFile),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewFireStoreClient(credentialsFile, projectID string) (*firestore.Client, error) {
	client, err := firestore.NewClient(
		context.TODO(),
		projectID,
		option.WithCredentialsFile(credentialsFile),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewFireStoreClientWithDatabase(credentialsFile, projectID, databaseID string) (*firestore.Client, error) {
	client, err := firestore.NewClientWithDatabase(
		context.TODO(),
		projectID,
		databaseID,
		option.WithCredentialsFile(credentialsFile),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
