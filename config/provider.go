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
package config

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiV1 "github.com/alibabacloud-go/darabonba-openapi/client"
	alibabadds "github.com/alibabacloud-go/dds-20151201/client"
	alibabards "github.com/alibabacloud-go/rds-20140815/v8/client"
	alibabavpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	osscred "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	ncpvpc "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	ncpvserver "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	vmongodb "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vmongodb"
	vmysql "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vmysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	firestorev1 "google.golang.org/api/firestore/v1"
	sqladmin "google.golang.org/api/sqladmin/v1"
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
		config.WithCredentialsProvider(awscred.NewStaticCredentialsProvider(accesskey, secretkey, "")),
		config.WithRegion(region),
		config.WithRetryMaxAttempts(5),
	)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// todo : deprecated 확인 후 변경 필요
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
		config.WithCredentialsProvider(awscred.NewStaticCredentialsProvider(accesskey, secretkey, "")),
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

func newAlibabaMongoDBConfig(username, password, host string, port int) *options.ClientOptions {
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

func NewAlibabaMongoDBClient(username, password, host string, port int) (*mongo.Client, error) {
	if err := validateInputs(&username, &password, &host, &port); err != nil {
		return nil, err
	}
	return mongo.Connect(context.Background(), newAlibabaMongoDBConfig(username, password, host, port))
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

func NewAWSRDBClient(accesskey, secretkey, region string) (*rds.Client, error) {
	cfg, err := newAWSConfig(accesskey, secretkey, region)
	if err != nil {
		return nil, err
	}

	return rds.NewFromConfig(*cfg), nil
}

func NewGCPSQLAdminClient(credJSON []byte) (*sqladmin.Service, error) {
	svc, err := sqladmin.NewService(context.TODO(), option.WithCredentialsJSON(credJSON))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func NewGCPClient(credentialsJson string) (*storage.Client, error) {
	var client *storage.Client
	var err error
	ctx := context.TODO()
	switch {

	case credentialsJson != "":
		client, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(credentialsJson)))

	default:
		return nil, errors.New("either credentialsFile or credentialsJson must be provided")
	}
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewAlibabaClient(region, accessKey, secretKey string) (*oss.Client, error) {

	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := oss.LoadDefaultConfig().
		WithEndpoint("https://oss-" + region + ".aliyuncs.com").
		WithCredentialsProvider(osscred.NewStaticCredentialsProvider(accessKey, secretKey)).
		WithRegion(region).
		WithRetryMaxAttempts(5)

	client := oss.NewClient(cfg)
	return client, nil
}

// NewAlibabaDDSClient builds an ApsaraDB for MongoDB (DDS) client from static credentials and a region.
func NewAlibabaDDSClient(region, accessKey, secretKey string) (*alibabadds.Client, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := &openapiV1.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(secretKey),
		RegionId:        tea.String(region),
		Endpoint:        tea.String("mongodb." + region + ".aliyuncs.com"),
	}
	return alibabadds.NewClient(cfg)
}

// NewAlibabaRDBClient builds an ApsaraDB RDS client from static credentials and a region.
func NewAlibabaRDBClient(region, accessKey, secretKey string) (*alibabards.Client, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(secretKey),
		RegionId:        tea.String(region),
		Endpoint:        tea.String("rds." + region + ".aliyuncs.com"),
	}
	return alibabards.NewClient(cfg)
}

// NewAlibabaVPCClient builds an Alibaba VPC client from static credentials and a region.
func NewAlibabaVPCClient(region, accessKey, secretKey string) (*alibabavpc.Client, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(secretKey),
		RegionId:        tea.String(region),
		Endpoint:        tea.String("vpc." + region + ".aliyuncs.com"),
	}
	return alibabavpc.NewClient(cfg)
}

// NewNCPCloudMongoDbClient builds an NCP Cloud DB for MongoDB V2 API service from static credentials.
func NewNCPCloudMongoDbClient(accessKey, secretKey string) (*vmongodb.V2ApiService, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := vmongodb.NewConfiguration(&ncloud.APIKey{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	return vmongodb.NewAPIClient(cfg).V2Api, nil
}

// NewNCPCloudMysqlClient builds an NCP Cloud DB for MySQL V2 API service from static credentials.
func NewNCPCloudMysqlClient(accessKey, secretKey string) (*vmysql.V2ApiService, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := vmysql.NewConfiguration(&ncloud.APIKey{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	return vmysql.NewAPIClient(cfg).V2Api, nil
}

// NewNCPVPCClient builds an NCP VPC V2 API service from static credentials.
func NewNCPVPCClient(accessKey, secretKey string) (*ncpvpc.V2ApiService, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := ncpvpc.NewConfiguration(&ncloud.APIKey{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	return ncpvpc.NewAPIClient(cfg).V2Api, nil
}

// NewNCPVServerClient builds an NCP VServer V2 API service from static credentials.
func NewNCPVServerClient(accessKey, secretKey string) (*ncpvserver.V2ApiService, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("accessKey and secretKey are required")
	}
	cfg := ncpvserver.NewConfiguration(&ncloud.APIKey{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	return ncpvserver.NewAPIClient(cfg).V2Api, nil
}

// NewGCPFirestoreAdminClient builds a Firestore Admin API client from a service
// account credentials JSON. Used for managing Firestore databases (instances).
func NewGCPFirestoreAdminClient(credJSON []byte) (*firestorev1.Service, error) {
	svc, err := firestorev1.NewService(context.TODO(), option.WithCredentialsJSON(credJSON))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func NewFireStoreClient(credentialsJson, projectID, databaseID string) (*firestore.Client, error) {
	var client *firestore.Client
	var err error

	ctx := context.TODO()
	if databaseID != "" {
		client, err = firestore.NewClientWithDatabase(ctx, projectID, databaseID, option.WithCredentialsJSON([]byte(credentialsJson)))
	} else {
		client, err = firestore.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(credentialsJson)))
	}
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
