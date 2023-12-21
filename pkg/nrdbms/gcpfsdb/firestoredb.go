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
package gcpfsdb

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"google.golang.org/api/iterator"
)

type FirestoreDBMS struct {
	provider utils.Provider
	region   string

	client *firestore.Client
	ctx    context.Context
}

type FirestoreDBOption func(*FirestoreDBMS)

func New(client *firestore.Client, region string, opts ...FirestoreDBOption) *FirestoreDBMS {
	dms := &FirestoreDBMS{
		provider: utils.GCP,
		region:   region,
		client:   client,
		ctx:      context.TODO(),
	}

	for _, opt := range opts {
		opt(dms)
	}

	return dms
}

// list table
func (f *FirestoreDBMS) ListTables() ([]string, error) {
	tableList := []string{}
	collIter := f.client.Collections(f.ctx)
	for {
		coll, err := collIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []string{}, err
		}
		tableList = append(tableList, coll.ID)
	}
	return tableList, nil
}

// delete table
func (f *FirestoreDBMS) DeleteTables(tableName string) error {
	iter := f.client.Collection(tableName).Documents(f.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		if _, err := doc.Ref.Delete(f.ctx); err != nil {
			return err
		}
	}
	return nil
}

// create table
func (f *FirestoreDBMS) CreateTable(tableName string) error {
	_, err := f.client.Collection(tableName).NewDoc().Set(f.ctx, map[string]interface{}{})
	return err
}

// import table
func (f *FirestoreDBMS) ImportTable(tableName string, srcData *[]map[string]interface{}) error {
	collRef := f.client.Collection(tableName)
	for _, dd := range *srcData {
		_, err := collRef.NewDoc().Set(f.ctx, dd)
		if err != nil {
			return err
		}
	}

	docIter := collRef.Documents(f.ctx)
	for {
		doc, err := docIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if len(doc.Data()) == 0 {
			_, err := doc.Ref.Delete(f.ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// export table
func (f *FirestoreDBMS) ExportTable(tableName string, dstData *[]map[string]interface{}) error {
	collRef := f.client.Collection(tableName)
	docIter := collRef.Documents(f.ctx)
	for {
		doc, err := docIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if len(doc.Data()) != 0 {
			*dstData = append(*dstData, doc.Data())
		}
	}
	return nil
}
