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
package gcpfs

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/rs/zerolog/log"
)

type GCPfs struct {
	provider   models.Provider
	projectID  string
	bucketName string
	region     string

	ctx       context.Context
	client    *storage.Client
	bktclient *storage.BucketHandle
}

// Creating a Bucket
func (f *GCPfs) CreateBucket() error {
	url := "http://localhost:1323/tumblebug/resources/objectStorage/" + f.bucketName
	// url := "http://mc-infra-manager:1323/tumblebug/resources/objectStorage/" + f.bucketName
	method := http.MethodHead
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	_, err := utils.RequestTumblebug(url, method, connName, nil)
	if err != nil {
		url = "http://localhost:1323/tumblebug/resources/objectStorage/" + f.bucketName
		// url := "http://mc-infra-manager:1323/tumblebug/resources/objectStorage/" + f.bucketName
		method = http.MethodPut

		_, err := utils.RequestTumblebug(url, method, connName, nil)
		if err != nil {
			fmt.Println("create error: ", err.Error())
			return err
		}

		return nil
	}
	return nil
}

// Delete Bucket
//
// Check and delete all objects in the bucket and delete the bucket
func (f *GCPfs) DeleteBucket() error {
	objList, err := f.ObjectList()
	if err != nil {
		return err
	}

	if len(objList) != 0 {
		// Divide objectIds into batches of 1000
		const batchSize = 1000
		var objectIds []string

		for _, object := range objList {
			objectIds = append(objectIds, object.Key)

			// When we reach batch size, delete objects
			if len(objectIds) == batchSize {
				if err := f.deleteObjectBatch(objectIds); err != nil {
					return err
				}
				// Reset objectIds for the next batch
				objectIds = []string{}
			}
		}

		// Delete any remaining objects
		if len(objectIds) > 0 {
			if err := f.deleteObjectBatch(objectIds); err != nil {
				return err
			}
		}
	}

	// Delete the bucket
	url := "http://localhost:1323/tumblebug/resources/objectStorage/" + f.bucketName
	// url := "http://mc-infra-manager:1323/tumblebug/resources/objectStorage/" + f.bucketName
	method := http.MethodDelete
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	_, err = utils.RequestTumblebug(url, method, connName, nil)
	if err != nil {
		return err
	}
	log.Info().Msg("DeleteDone")
	return nil
}

// deleteObjectBatch deletes a batch of objects
func (f *GCPfs) deleteObjectBatch(keys []string) error {
	url := "http://localhost:1323/tumblebug/resources/objectStorage/" + f.bucketName + "?delete=true"
	// url := "http://mc-infra-manager:1323/tumblebug/resources/objectStorage/" + f.bucketName
	method := http.MethodPost
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	deleteReq := models.DeleteRequest{
		XMLNS: "http://s3.amazonaws.com/doc/2006-03-01/",
	}
	for _, key := range keys {
		deleteReq.Objects = append(deleteReq.Objects, models.S3Object{Key: key})
	}
	// 보기 좋게 들여쓰기된 XML 생성
	output, err := xml.MarshalIndent(deleteReq, "", "    ")
	if err != nil {
		return err
	}

	// XML 헤더 추가
	_, rerr := utils.RequestTumblebug(url, method, connName, []byte(xml.Header+string(output)))
	if rerr != nil {
		return err
	}

	return nil
}

// Open function
func (f *GCPfs) Open(name string) (io.ReadCloser, error) {
	r, err := f.bktclient.Object(name).NewReader(f.ctx)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Create function
func (f *GCPfs) Create(name string) (io.WriteCloser, error) {
	return f.bktclient.Object(name).NewWriter(f.ctx), nil
}

// Look up the list of objects in your bucket
func (f *GCPfs) ObjectList() ([]*models.Object, error) {
	return f.ObjectListWithFilter(nil)
}

func (f *GCPfs) ObjectListWithFilter(flt *filtering.ObjectFilter) ([]*models.Object, error) {
	log.Debug().Msg("[GCP] filtering")
	var objList []*models.Object

	// var query *storage.Query
	// if flt != nil && flt.Path != "" {
	// 	pre := strings.TrimPrefix(flt.Path, "/")
	// 	query = &storage.Query{Prefix: pre}
	// }

	url := fmt.Sprintf("%s%s", "http://localhost:1323/tumblebug/resources/objectStorage/", f.bucketName)
	method := http.MethodGet
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	result, err := utils.RequestTumblebug(url, method, connName, nil)
	if err != nil {
		return nil, err
	}

	var resp models.ListBucketResult
	if err := json.Unmarshal(result, &resp); err != nil {
		fmt.Println("error: ", err.Error())
		return []*models.Object{}, fmt.Errorf("failed to get objects: %w", err)
	}

	for _, o := range resp.Contents {

		candidate := filtering.Candidate{
			Key:          o.Key,
			Size:         o.Size,
			LastModified: o.LastModified,
		}

		log.Debug().
			Str("gcp key", candidate.Key).
			Int64("gcp bytes", candidate.Size).
			Time("gcp date", candidate.LastModified).
			Msg("gcp value")

		// filtering.MatchCandidate() 호출
		if filtering.MatchCandidate(flt, candidate) {
			objList = append(objList, &models.Object{
				// ETag:         o.Etag,
				Key:          o.Key,
				LastModified: o.LastModified,
				Size:         o.Size,
				StorageClass: o.StorageClass,
				Provider:     f.provider,
			})
		}
	}
	return objList, nil
}

func New(client *storage.Client, projectID, bucketName string, region string) *GCPfs {
	gfs := &GCPfs{
		ctx:        context.TODO(),
		bucketName: bucketName,
		client:     client,
		bktclient:  client.Bucket(bucketName),
		provider:   models.GCP,
		region:     region,
		projectID:  projectID,
	}

	return gfs
}

func (f *GCPfs) BucketList() ([]models.Bucket, error) {
	url := "http://localhost:1323/tumblebug/resources/objectStorage"
	// url := "http://mc-infra-manager:1323/tumblebug/resources/objectStorage"
	method := http.MethodGet
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	body, err := utils.RequestTumblebug(url, method, connName, nil)
	if err != nil {
		return []models.Bucket{}, fmt.Errorf("failed to get buckets: %w", err)
	}

	// Parse the response to extract public key and token ID
	var res models.ListAllMyBucketsResult
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Println("error: ", err.Error())
		return []models.Bucket{}, fmt.Errorf("failed to get buckets: %w", err)
	}

	// 버킷이 비어 있으면 빈 리스트 반환
	if res.Buckets.Bucket == nil {
		return []models.Bucket{}, nil
	}

	return res.Buckets.Bucket, nil
}
