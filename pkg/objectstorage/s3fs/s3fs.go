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
package s3fs

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/rs/zerolog/log"
)

type reader struct {
	r        *io.PipeReader
	ch       chan error
	cancel   context.CancelFunc
	chkClose bool
}

func (p *reader) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

func (p *reader) Close() error {
	if !p.chkClose {
		p.chkClose = true
		return p.r.Close()
	}
	return nil
}

type writer struct {
	w        *io.PipeWriter
	ch       chan error
	cancel   context.CancelFunc
	chkClose bool
}

func (p *writer) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

func (p *writer) Close() error {
	if !p.chkClose {
		p.chkClose = true
		_ = p.w.Close()
		return <-p.ch
	}
	return nil
}

type fakeWriteAt struct {
	W io.Writer
}

func (w *fakeWriteAt) WriteAt(p []byte, off int64) (n int, err error) {
	return w.W.Write(p)
}

type S3FS struct {
	provider   models.Provider
	bucketName string
	region     string

	client     *s3.Client
	ctx        context.Context
	uploader   manager.Uploader
	downloader manager.Downloader
}

// Creating a Bucket
//
// Aws imposes location constraints when creating buckets
func (f *S3FS) CreateBucket() error {
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

// Delete Bucket
// Check and delete all objects in the bucket and delete the bucket
func (f *S3FS) DeleteBucket() error {
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

// deleteObjectBatch deletes a batch of objects
func (f *S3FS) deleteObjectBatch(keys []string) error {
	path := "/tumblebug/resources/objectStorage/" + f.bucketName + "?delete=true"
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
	_, rerr := utils.RequestTumblebug(path, method, connName, []byte(xml.Header+string(output)))
	if rerr != nil {
		return err
	}

	return nil
}

// Open function using pipeline
func (f *S3FS) Open(name string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()
	ch := make(chan error)
	ctx, cancel := context.WithCancel(f.ctx)
	go func() {
		defer cancel()
		_, err := f.downloader.Download(
			ctx,
			&fakeWriteAt{W: pw},
			&s3.GetObjectInput{
				Bucket: aws.String(f.bucketName),
				Key:    aws.String(name),
			}, func(d *manager.Downloader) { d.Concurrency = 1 },
		)
		if cerr := pw.Close(); cerr != nil {
			err = cerr
		}
		ch <- err
	}()

	return &reader{r: pr, ch: ch, cancel: cancel, chkClose: false}, nil
}

// Create function using pipeline
func (f *S3FS) Create(name string) (io.WriteCloser, error) {
	pr, pw := io.Pipe()
	ch := make(chan error)
	ctx, cancel := context.WithCancel(f.ctx)
	go func() {
		defer cancel()
		_, err := f.uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(f.bucketName),
			Key:    aws.String(name),
			Body:   pr,
		})
		ch <- err
	}()

	return &writer{w: pw, ch: ch, cancel: cancel, chkClose: false}, nil
}

// Look up the list of objects in your bucket
// func (f *S3FS) ObjectList() ([]*models.Object, error) {
// 	var objlist []*models.Object
// 	var ContinuationToken *string

// 	for {
// 		LOut, err := f.client.ListObjectsV2(
// 			f.ctx,
// 			&s3.ListObjectsV2Input{
// 				Bucket:            aws.String(f.bucketName),
// 				ContinuationToken: ContinuationToken,
// 			},
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		for _, obj := range LOut.Contents {
// 			objlist = append(objlist, &models.Object{
// 				ETag:         *obj.ETag,
// 				Key:          *obj.Key,
// 				LastModified: *obj.LastModified,
// 				Size:         *obj.Size,
// 				StorageClass: string(obj.StorageClass),
// 			})
// 		}

// 		if LOut.NextContinuationToken == nil {
// 			break
// 		}

// 		ContinuationToken = LOut.NextContinuationToken
// 	}

// 	return objlist, nil
// }

func New(provider models.Provider, client *s3.Client, bucketName, region string) *S3FS {
	sfs := &S3FS{
		ctx:        context.TODO(),
		provider:   provider,
		bucketName: bucketName,
		region:     region,
		client:     client,
	}

	sfs.uploader = *manager.NewUploader(client, func(u *manager.Uploader) { u.Concurrency = 1; u.PartSize = 128 * 1024 * 1024 })
	sfs.downloader = *manager.NewDownloader(client, func(d *manager.Downloader) { d.Concurrency = 1; d.PartSize = 128 * 1024 * 1024 })

	return sfs
}

func (f *S3FS) ObjectListWithFilter(flt *filtering.ObjectFilter) ([]*models.Object, error) {
	log.Debug().Msg("[S3FS] filtering")
	var out []*models.Object
	// var token *string

	var prefix *string
	if flt != nil && flt.Path != "" {
		pre := strings.TrimPrefix(flt.Path, "/")
		prefix = aws.String(pre)
	}

	for {
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
			c := filtering.Candidate{
				Key:          o.Key,
				Size:         o.Size,
				LastModified: o.LastModified,
			}

			log.Debug().Str("key", c.Key).Int64("size", c.Size).
				Msg("[S3FS] candidate")

			matched := filtering.MatchCandidate(flt, c)
			if !matched {
				if flt != nil {
					log.Debug().
						Str("key", c.Key).
						Str("prefix", aws.ToString(prefix)).
						Strs("exact", flt.Exact).
						Str("modifiedDate", c.LastModified.String()).
						Msg("[S3FS] filtered out")
				}
				continue
			}

			out = append(out, &models.Object{
				ETag: o.ETag,
				// ETag:         aws.ToString(o.ETag),
				Key:          c.Key,
				LastModified: c.LastModified,
				Size:         c.Size,
				StorageClass: o.StorageClass,
				Provider:     f.provider,
			})
		}

		break
	}
	return out, nil
}

func (f *S3FS) ObjectList() ([]*models.Object, error) {
	return f.ObjectListWithFilter(nil)
}

func (f *S3FS) BucketList() ([]models.Bucket, error) {
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
