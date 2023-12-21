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
package osc_test

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/gcpfs"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/s3fs"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/cloud-barista/cm-data-mold/service/osc"
	"google.golang.org/api/option"
)

// dynamo to firestore example
func TestMain(m *testing.M) {
	awsosc, err := AWSInfo("your-aws-accessKey", "your-aws-secretKey", "your-aws-reigon", "your-aws-bucket-name")
	if err != nil {
		panic(err)
	}

	gcposc, err := GCPInfo("your-gcp-projectID", "your-gcp-credentialsFile", "your-gcp-reigon", "your-gcp-bucket-name")
	if err != nil {
		panic(err)
	}

	// aws import
	if err := awsosc.MPut("your-upload-directory-path"); err != nil {
		panic(err)
	}

	// aws export
	if err := awsosc.MGet("your-upload-directory-path"); err != nil {
		panic(err)
	}

	// s3 to gcp
	if err := awsosc.Copy(gcposc); err != nil {
		panic(err)
	}
}

func AWSInfo(accessKey, secretKey, region, bucketName string) (*osc.OSController, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
		config.WithRetryMaxAttempts(5),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })

	return osc.New(s3fs.New(utils.AWS, client, bucketName, region))
}

func NCPInfo(accessKey, secretKey, endpoint, region, bucketName string) (*osc.OSController, error) {
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == "S3" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		}

		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
		config.WithRetryMaxAttempts(5),
		config.WithEndpointResolver(customResolver),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })

	return osc.New(s3fs.New(utils.AWS, client, bucketName, region))
}

func GCPInfo(projectID, credentialsFile, region, bucketName string) (*osc.OSController, error) {
	client, err := storage.NewClient(
		context.TODO(),
		option.WithCredentialsFile(credentialsFile),
	)

	if err != nil {
		return nil, err
	}

	return osc.New(gcpfs.New(client, projectID, bucketName, region))
}
