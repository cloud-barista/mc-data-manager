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
package alibabamgdb

import (
	"context"

	"github.com/cloud-barista/mc-data-manager/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlibabaMongoDBMS struct {
	provider models.Provider
	dbName   string

	client *mongo.Client
	db     *mongo.Database
	ctx    context.Context
}

type AlibabaMongoDBOption func(*AlibabaMongoDBMS)

func New(client *mongo.Client, databaseName string, opts ...AlibabaMongoDBOption) *AlibabaMongoDBMS {
	dms := &AlibabaMongoDBMS{
		provider: models.ALIBABA,
		dbName:   databaseName,
		client:   client,
		ctx:      context.TODO(),
		db:       client.Database(databaseName),
	}

	for _, opt := range opts {
		opt(dms)
	}

	return dms
}

// list table
func (a *AlibabaMongoDBMS) ListTables() ([]string, error) {
	return a.db.ListCollectionNames(a.ctx, bson.D{})
}

// delete table
func (a *AlibabaMongoDBMS) DeleteTables(tableName string) error {
	return a.client.Database(a.dbName).Collection(tableName).Drop(a.ctx)
}

// create table
func (a *AlibabaMongoDBMS) CreateTable(tableName string) error {
	_, err := a.db.Collection(tableName).InsertOne(a.ctx, map[string]interface{}{})
	if err != nil {
		return err
	}

	_, err = a.db.Collection(tableName).DeleteOne(a.ctx, map[string]interface{}{})
	return err
}

// import table
func (a *AlibabaMongoDBMS) ImportTable(tableName string, srcData *[]map[string]interface{}) error {
	for _, data := range *srcData {
		_, err := a.db.Collection(tableName).InsertOne(a.ctx, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// export table
func (a *AlibabaMongoDBMS) ExportTable(tableName string, dstData *[]map[string]interface{}) error {
	cursor, err := a.db.Collection(tableName).Find(a.ctx, map[string]interface{}{})
	if err != nil {
		return err
	}
	defer cursor.Close(a.ctx)

	for cursor.Next(a.ctx) {
		var result map[string]interface{}
		err := cursor.Decode(&result)
		if err != nil {
			return err
		}

		if oid, ok := result["_id"].(primitive.ObjectID); ok {
			result["_id"] = oid.Hex()
		}

		*dstData = append(*dstData, result)
	}
	return nil
}
