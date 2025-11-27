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
package alibabafs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/rs/zerolog/log"
)

// ErrNotImplemented is returned by stub methods until the Alibaba integration is complete.
var ErrNotImplemented = errors.New("alibaba object storage: not implemented")

// AlibabaFS is a placeholder that will back the OSC interface for Alibaba Cloud.
type AlibabaFS struct {
	provider   models.Provider
	endpoint   string
	region     string
	bucketName string

	ctx    context.Context
	client *oss.Client
}

// ossWriter bridges an io.PipeWriter with the goroutine uploading to OSS.
type ossWriter struct {
	w      *io.PipeWriter
	ch     chan error
	closed bool
}

func (w *ossWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *ossWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	_ = w.w.Close()
	return <-w.ch
}

// CreateBucket will provision a bucket if it is not already present.
func (f *AlibabaFS) CreateBucket() error {
	path := "/tumblebug/resources/objectStorage/" + f.bucketName
	method := http.MethodHead
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	_, err := utils.RequestTumblebug(path, method, connName, nil)
	if err != nil {
		path = "/tumblebug/resources/objectStorage/" + f.bucketName
		method = http.MethodPut

		_, err := utils.RequestTumblebug(path, method, connName, nil)
		if err != nil {
			fmt.Println("create error: ", err.Error())
			return err
		}

		return nil
	}
	return nil
	// return err
}

// DeleteBucket removes all objects in a bucket and deletes the bucket itself.
func (f *AlibabaFS) DeleteBucket() error {
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
	path := "/tumblebug/resources/objectStorage/" + f.bucketName
	method := http.MethodDelete
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	_, err = utils.RequestTumblebug(path, method, connName, nil)
	if err != nil {
		return err
	}
	log.Info().Msg("DeleteDone")
	return nil
}

// deleteObjectBatch deletes objects in manageable chunks.
func (f *AlibabaFS) deleteObjectBatch(keys []string) error {
	// TODO: marshal keys into the request body and invoke the OpenAPI client.
	return ErrNotImplemented
}

// ObjectList yields the objects contained within the configured bucket.
func (f *AlibabaFS) ObjectList() ([]*models.Object, error) {
	return f.ObjectListWithFilter(nil)
}

// ObjectListWithFilter filters the Alibaba objects according to the supplied matcher.
func (f *AlibabaFS) ObjectListWithFilter(flt *filtering.ObjectFilter) ([]*models.Object, error) {
	log.Debug().Msg("[ALIBABA] filtering")
	var objList []*models.Object

	// var query *storage.Query
	// if flt != nil && flt.Path != "" {
	// 	pre := strings.TrimPrefix(flt.Path, "/")
	// 	query = &storage.Query{Prefix: pre}
	// }

	path := "/tumblebug/resources/objectStorage/" + f.bucketName
	method := http.MethodGet
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	result, err := utils.RequestTumblebug(path, method, connName, nil)
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
			Str("alibaba key", candidate.Key).
			Int64("alibaba bytes", candidate.Size).
			Time("alibaba date", candidate.LastModified).
			Msg("alibaba value")

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

// BucketList returns all buckets that are available for the configured account.
func (f *AlibabaFS) BucketList() ([]models.Bucket, error) {
	path := "/tumblebug/resources/objectStorage"
	method := http.MethodGet
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	body, err := utils.RequestTumblebug(path, method, connName, nil)
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

// Open streams a single object from Alibaba Cloud OSS.
func (f *AlibabaFS) Open(name string) (io.ReadCloser, error) {
	// TODO: wire Alibaba download stream into an io.Reader.
	return nil, ErrNotImplemented
}

// Create opens a writer that uploads an object to the configured bucket.
func (f *AlibabaFS) Create(name string) (io.WriteCloser, error) {
	if f.client == nil {
		return nil, fmt.Errorf("alibaba oss client is not configured")
	}

	ctx := f.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	pr, pw := io.Pipe()
	ch := make(chan error, 1)

	go func() {
		_, err := f.client.PutObject(ctx, &oss.PutObjectRequest{
			Bucket: oss.Ptr(f.bucketName),
			Key:    oss.Ptr(name),
			Body:   pr,
		})
		if cerr := pr.Close(); cerr != nil && err == nil {
			err = cerr
		}
		ch <- err
	}()

	return &ossWriter{w: pw, ch: ch}, nil
}

// New builds a controller-compatible filesystem instance for Alibaba Cloud.
func New(provider models.Provider, client *oss.Client, endpoint, bucketName, region string) *AlibabaFS {
	alibabafs := &AlibabaFS{
		provider:   provider,
		endpoint:   endpoint,
		bucketName: bucketName,
		region:     region,
		ctx:        context.TODO(),
		client:     client,
	}

	return alibabafs
}
