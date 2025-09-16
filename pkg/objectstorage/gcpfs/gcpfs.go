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
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
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
	_, err := f.bktclient.Attrs(f.ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist {
			return f.bktclient.Create(f.ctx, f.projectID, &storage.BucketAttrs{
				Name:     f.bucketName,
				Location: f.region,
			})
		}
		return err
	}
	return nil
}

// Delete Bucket
//
// Check and delete all objects in the bucket and delete the bucket
func (f *GCPfs) DeleteBucket() error {
	iter := f.bktclient.Objects(f.ctx, &storage.Query{})
	for {
		attr, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return err
		}
		if err := f.bktclient.Object(attr.Name).Delete(f.ctx); err != nil {
			return err
		}
	}
	return f.bktclient.Delete(f.ctx)
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

	var query *storage.Query
	if flt != nil && flt.Path != "" {
		pre := strings.TrimPrefix(flt.Path, "/")
		query = &storage.Query{Prefix: pre}
	}

	log.Debug().Str("path", query.Prefix).Msg("[gcp path filtering]")
	it := f.bktclient.Objects(f.ctx, query)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		candidate := filtering.Candidate{
			Key:          objAttrs.Name,
			Size:         objAttrs.Size,
			LastModified: objAttrs.Created,
		}

		log.Debug().
			Str("gcp key", candidate.Key).
			Int64("gcp bytes", candidate.Size).
			Time("gcp date", candidate.LastModified).
			Msg("gcp value")

		// filtering.MatchCandidate() 호출
		if filtering.MatchCandidate(flt, candidate) {
			objList = append(objList, &models.Object{
				ETag:         objAttrs.Etag,
				Key:          objAttrs.Name,
				LastModified: objAttrs.Created,
				Size:         objAttrs.Size,
				StorageClass: objAttrs.StorageClass,
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
