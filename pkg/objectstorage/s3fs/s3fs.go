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
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	nsId := utils.GetNsId()
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	headPath := "/tumblebug/ns/" + nsId + "/resources/objectStorage/" + f.bucketName
	_, err := utils.RequestTumblebug(headPath, http.MethodHead, connName, nil)
	if err == nil {
		return nil
	}

	createBody := []byte(fmt.Sprintf(`{"bucketName":"%s","connectionName":"%s"}`, f.bucketName, connName))
	createPath := "/tumblebug/ns/" + nsId + "/resources/objectStorage"
	_, err = utils.RequestTumblebug(createPath, http.MethodPut, connName, createBody)
	if err != nil {
		fmt.Println("create error: ", err.Error())
		return err
	}
	return nil
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
	nsId := utils.GetNsId()
	path := "/tumblebug/ns/" + nsId + "/resources/objectStorage/" + f.bucketName
	method := http.MethodDelete
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	_, err = utils.RequestTumblebug(path, method, connName, nil)
	if err != nil {
		return err
	}
	log.Info().Msg("DeleteDone")
	return nil
}

func (f *S3FS) DeleteObject(name string) error {
	return f.deleteObjectBatch([]string{name})
}

// deleteObjectBatch deletes a batch of objects
func (f *S3FS) deleteObjectBatch(keys []string) error {
	nsId := utils.GetNsId()
	path := "/tumblebug/ns/" + nsId + "/resources/objectStorage/" + f.bucketName + "?delete=true"
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

// presignedURLResponse는 Tumblebug Presigned URL API의 응답 구조체입니다.
type presignedURLResponse struct {
	PresignedURL string `json:"presignedURL"`
	Expires      int64  `json:"expires"`
	Method       string `json:"method"`
}

// openWithTumblebug은 Tumblebug의 Presigned URL API를 통해 오브젝트를 다운로드합니다.
//
// 기존 Open()이 AWS SDK를 직접 사용하는 것과 달리,
// 이 함수는 Tumblebug에 Presigned URL 발급을 요청한 뒤
// 해당 URL로 HTTP GET을 수행하여 스트림을 반환합니다.
//
// POST /ns/{nsId}/resources/objectStorage/{osId}/object/{objectKey}/presignedUrl?operation=download
func (f *S3FS) Open(name string) (io.ReadCloser, error) {
	nsId := utils.GetNsId()
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	// objectKey에 슬래시 등 특수문자가 포함될 수 있으므로 path 세그먼트 단위로 인코딩합니다.
	// url.PathEscape는 '/'를 인코딩하지 않으므로, 키 전체를 하나의 세그먼트로 처리하기 위해
	// url.QueryEscape 후 '+'를 '%20'으로 변환하는 방식을 사용합니다.
	encodedKey := strings.NewReplacer("+", "%20").Replace(url.QueryEscape(name))

	path := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s/object/%s/presignedUrl?operation=download&expires=3600",
		nsId, f.bucketName, encodedKey)

	body, err := utils.RequestTumblebug(path, http.MethodPost, connName, nil)
	if err != nil {
		return nil, fmt.Errorf("openWithTumblebug: failed to generate presigned URL for %q: %w", name, err)
	}

	var resp presignedURLResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("openWithTumblebug: failed to parse presigned URL response: %w", err)
	}
	if resp.PresignedURL == "" {
		return nil, fmt.Errorf("openWithTumblebug: empty presigned URL returned for %q", name)
	}

	log.Debug().Str("key", name).Str("presignedURL", resp.PresignedURL).
		Msg("[S3FS] openWithTumblebug: downloading via presigned URL")

	httpResp, err := http.Get(resp.PresignedURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("openWithTumblebug: HTTP GET failed: %w", err)
	}
	if httpResp.StatusCode != http.StatusOK {
		_ = httpResp.Body.Close()
		return nil, fmt.Errorf("openWithTumblebug: unexpected status %d for %q", httpResp.StatusCode, name)
	}

	return httpResp.Body, nil
}

// tumblebugWriter는 데이터를 메모리에 버퍼링한 뒤 Close() 시점에
// Content-Length를 명시하여 Presigned URL로 한 번에 업로드합니다.
//
// AWS S3 Presigned URL은 Transfer-Encoding: chunked를 지원하지 않으므로
// io.Pipe 스트리밍 방식 대신 버퍼링 후 전송 방식을 사용합니다.
type tumblebugWriter struct {
	buf          bytes.Buffer
	presignedURL string
	name         string
	ctx          context.Context
	chkClose     bool
}

func (w *tumblebugWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *tumblebugWriter) Close() error {
	if w.chkClose {
		return nil
	}
	w.chkClose = true

	data := w.buf.Bytes()

	req, err := http.NewRequestWithContext(w.ctx, http.MethodPut, w.presignedURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("createWithTumblebug: failed to create PUT request: %w", err)
	}
	req.ContentLength = int64(len(data))

	log.Debug().
		Str("key", w.name).
		Str("method", http.MethodPut).
		Int64("contentLength", req.ContentLength).
		Msg("[S3FS] createWithTumblebug: sending PUT request")

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("createWithTumblebug: PUT request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)

	log.Debug().
		Str("key", w.name).
		Int("statusCode", httpResp.StatusCode).
		Str("responseBody", string(respBody)).
		Msg("[S3FS] createWithTumblebug: PUT response")

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("createWithTumblebug: unexpected status %d for %q, body: %s",
			httpResp.StatusCode, w.name, string(respBody))
	}

	log.Info().Str("key", w.name).Int("statusCode", httpResp.StatusCode).
		Msg("[S3FS] createWithTumblebug: upload succeeded")
	return nil
}

// createWithTumblebug은 Tumblebug의 Presigned URL API를 통해 오브젝트를 업로드합니다.
//
// 기존 Create()가 AWS SDK uploader를 직접 사용하는 것과 달리,
// 이 함수는 Tumblebug에 Presigned URL 발급을 요청한 뒤
// 데이터를 버퍼링하여 Close() 시점에 Content-Length와 함께 HTTP PUT으로 전송합니다.
//
// POST /ns/{nsId}/resources/objectStorage/{osId}/object/{objectKey}/presignedUrl?operation=upload
func (f *S3FS) Create(name string) (io.WriteCloser, error) {
	nsId := utils.GetNsId()
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	encodedKey := strings.NewReplacer("+", "%20").Replace(url.QueryEscape(name))
	path := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s/object/%s/presignedUrl?operation=upload&expires=3600",
		nsId, f.bucketName, encodedKey)

	body, err := utils.RequestTumblebug(path, http.MethodPost, connName, nil)
	if err != nil {
		return nil, fmt.Errorf("createWithTumblebug: failed to generate presigned URL for %q: %w", name, err)
	}

	var resp presignedURLResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("createWithTumblebug: failed to parse presigned URL response: %w", err)
	}
	if resp.PresignedURL == "" {
		return nil, fmt.Errorf("createWithTumblebug: empty presigned URL returned for %q", name)
	}

	log.Debug().Str("key", name).Str("presignedURL", resp.PresignedURL).
		Msg("[S3FS] createWithTumblebug: presigned URL acquired")

	return &tumblebugWriter{
		presignedURL: resp.PresignedURL,
		name:         name,
		ctx:          f.ctx,
	}, nil
}

// Open function using pipeline
func (f *S3FS) OpenDeprecated(name string) (io.ReadCloser, error) {
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
func (f *S3FS) CreateDeprecated(name string) (io.WriteCloser, error) {
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
		nsId := utils.GetNsId()
		path := "/tumblebug/ns/" + nsId + "/resources/objectStorage/" + f.bucketName
		method := http.MethodGet
		connName := fmt.Sprintf("%s-%s", f.provider, f.region)

		result, err := utils.RequestTumblebug(path, method, connName, nil)
		if err != nil {
			return nil, err
		}

		var resp models.ObjectStorage
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

func (f *S3FS) BucketList(filterKey, filterVal string) ([]models.ObjectStorage, error) {
	nsId := utils.GetNsId()
	path := "/tumblebug/ns/" + nsId + "/resources/objectStorage"
	if filterKey != "" && filterVal != "" {
		path += "?filterKey=" + url.QueryEscape(filterKey) + "&filterVal=" + url.QueryEscape(filterVal)
	}
	method := http.MethodGet
	connName := fmt.Sprintf("%s-%s", f.provider, f.region)

	body, err := utils.RequestTumblebug(path, method, connName, nil)
	if err != nil {
		return []models.ObjectStorage{}, fmt.Errorf("failed to get buckets: %w", err)
	}

	var res models.ObjectStorageListResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return []models.ObjectStorage{}, fmt.Errorf("failed to parse bucket list response: %w", err)
	}

	return res.ObjectStorage, nil
}
