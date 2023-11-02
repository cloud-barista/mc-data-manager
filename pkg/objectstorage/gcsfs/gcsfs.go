package gcsfs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"google.golang.org/api/iterator"
)

type GCSfs struct {
	provider   utils.Provider
	projectID  string
	bucketName string
	region     string

	ctx       context.Context
	client    *storage.Client
	bktclient *storage.BucketHandle
}

// Creating a Bucket
func (f *GCSfs) CreateBucket() error {
	return f.bktclient.Create(f.ctx, f.projectID, &storage.BucketAttrs{
		Name:     f.bucketName,
		Location: f.region,
	})
}

// Delete Bucket
//
// Check and delete all objects in the bucket and delete the bucket
func (f *GCSfs) DeleteBucket() error {
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
func (f *GCSfs) Open(name string) (io.ReadCloser, error) {
	r, err := f.bktclient.Object(name).NewReader(f.ctx)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Create function
func (f *GCSfs) Create(name string) (io.WriteCloser, error) {
	return f.bktclient.Object(name).NewWriter(f.ctx), nil
}

// Look up the list of objects in your bucket
func (f *GCSfs) ObjectList() ([]*utils.Object, error) {
	var objList []*utils.Object
	it := f.bktclient.Objects(f.ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		objList = append(objList, &utils.Object{
			ETag:         objAttrs.Etag,
			Key:          objAttrs.Name,
			LastModified: objAttrs.Created,
			Size:         objAttrs.Size,
			StorageClass: objAttrs.StorageClass,
		})
	}
	return objList, nil
}

func New(client *storage.Client, projectID, bucketName string, region string) *GCSfs {
	gfs := &GCSfs{
		ctx:        context.TODO(),
		bucketName: bucketName,
		client:     client,
		bktclient:  client.Bucket(bucketName),
		provider:   utils.GCP,
		region:     region,
		projectID:  projectID,
	}

	return gfs
}