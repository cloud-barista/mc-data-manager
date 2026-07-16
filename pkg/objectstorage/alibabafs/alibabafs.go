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
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/rs/zerolog/log"
)

// AlibabaFS backs the OSC interface for Alibaba Cloud.
type AlibabaFS struct {
	provider   models.Provider
	endpoint   string
	region     string
	bucketName string

	ctx    context.Context
	client *oss.Client
}

// CreateBucket will provision a bucket if it is not already present.
func (f *AlibabaFS) CreateBucket() error {
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

// deleteObjectBatch deletes objects in manageable chunks.
func (f *AlibabaFS) deleteObjectBatch(keys []string) error {
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
func (f *AlibabaFS) BucketList(filterKey, filterVal string) ([]models.ObjectStorage, error) {
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

// presignedURLResponse는 Tumblebug Presigned URL API의 응답 구조체입니다.
type presignedURLResponse struct {
	PresignedURL string `json:"presignedURL"`
	Expires      int64  `json:"expires"`
	Method       string `json:"method"`
}

// Tumblebug의 Presigned URL API를 통해 오브젝트를 다운로드합니다.
//
// 기존 Open()이 Alibaba SDK를 직접 사용하는 것과 달리,
// 이 함수는 Tumblebug에 Presigned URL 발급을 요청한 뒤
// 해당 URL로 HTTP GET을 수행하여 스트림을 반환합니다.
//
// POST /ns/{nsId}/resources/objectStorage/{osId}/object/{objectKey}/presignedUrl?operation=download
func (f *AlibabaFS) Open(name string) (io.ReadCloser, error) {
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
		Msg("[ALIBABA] openWithTumblebug: downloading via presigned URL")

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
// Alibaba OSS Presigned URL은 Transfer-Encoding: chunked를 지원하지 않으므로
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
		Msg("[ALIBABA] createWithTumblebug: sending PUT request")

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
		Msg("[ALIBABA] createWithTumblebug: PUT response")

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("createWithTumblebug: unexpected status %d for %q, body: %s",
			httpResp.StatusCode, w.name, string(respBody))
	}

	log.Info().Str("key", w.name).Int("statusCode", httpResp.StatusCode).
		Msg("[ALIBABA] createWithTumblebug: upload succeeded")
	return nil
}

// createWithTumblebug은 Tumblebug의 Presigned URL API를 통해 오브젝트를 업로드합니다.
//
// 기존 Create()가 Alibaba SDK를 직접 사용하는 것과 달리,
// 이 함수는 Tumblebug에 Presigned URL 발급을 요청한 뒤
// 데이터를 버퍼링하여 Close() 시점에 Content-Length와 함께 HTTP PUT으로 전송합니다.
//
// POST /ns/{nsId}/resources/objectStorage/{osId}/object/{objectKey}/presignedUrl?operation=upload
func (f *AlibabaFS) Create(name string) (io.WriteCloser, error) {
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
		Msg("[ALIBABA] createWithTumblebug: presigned URL acquired")

	return &tumblebugWriter{
		presignedURL: resp.PresignedURL,
		name:         name,
		ctx:          f.ctx,
	}, nil
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
