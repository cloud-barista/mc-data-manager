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
package ncpmgdb

import (
	"context"

	"github.com/cloud-barista/mc-data-manager/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NCPMongoDBMS struct {
	provider models.Provider
	dbName   string

	client *mongo.Client
	db     *mongo.Database
	ctx    context.Context
}

type NCPMongoDBOption func(*NCPMongoDBMS)

func New(client *mongo.Client, databaseName string, opts ...NCPMongoDBOption) *NCPMongoDBMS {
	dms := &NCPMongoDBMS{
		provider: models.NCP,
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
func (n *NCPMongoDBMS) ListTables() ([]string, error) {
	return n.db.ListCollectionNames(n.ctx, bson.D{})
}

// list database
func (n *NCPMongoDBMS) ListDatabases() ([]string, error) {
	names, err := n.client.ListDatabaseNames(n.ctx, bson.D{})
	if err != nil {
		return names, err
	}

	dbList := []string{}
	for _, name := range names {
		if name != "admin" && name != "config" && name != "local" {
			dbList = append(dbList, name)
		}
	}
	return dbList, nil
}

// delete table
func (n *NCPMongoDBMS) DeleteTables(tableName string) error {
	return n.client.Database(n.dbName).Collection(tableName).Drop(n.ctx)
}

// create table
func (n *NCPMongoDBMS) CreateTable(tableName string) error {
	_, err := n.db.Collection(tableName).InsertOne(n.ctx, map[string]interface{}{})
	if err != nil {
		return err
	}

	_, err = n.db.Collection(tableName).DeleteOne(n.ctx, map[string]interface{}{})
	return err
}

// import table
func (n *NCPMongoDBMS) ImportTable(tableName string, srcData *[]map[string]interface{}) error {
	if len(*srcData) == 0 {
		return nil
	}

	docs := make([]interface{}, len(*srcData))
	for i, data := range *srcData {
		docs[i] = data
	}

	_, err := n.db.Collection(tableName).InsertMany(n.ctx, docs)
	return err
}

// export table
func (n *NCPMongoDBMS) ExportTable(tableName string, dstData *[]map[string]interface{}) error {
	cursor, err := n.db.Collection(tableName).Find(n.ctx, map[string]interface{}{})
	if err != nil {
		return err
	}
	defer cursor.Close(n.ctx)

	for cursor.Next(n.ctx) {
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
