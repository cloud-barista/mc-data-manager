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
package awsdnmdb

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cloud-barista/mc-data-manager/models"
	"golang.org/x/sync/errgroup"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

type DynamoDBMS struct {
	provider models.Provider
	region   string

	client *dynamodb.Client
	ctx    context.Context

	partitionKey          *string
	sortKey               *[]string
	billingMode           *string
	deleteProtection      *bool
	provisionedThroughput *types.ProvisionedThroughput
	sSESpecification      *types.SSESpecification
	tableClass            *string
	tags                  *[]types.Tag
}

type DynamoDBOption func(*DynamoDBMS)

func New(client *dynamodb.Client, region string, opts ...DynamoDBOption) *DynamoDBMS {
	dms := &DynamoDBMS{
		provider:         models.AWS,
		region:           region,
		client:           client,
		ctx:              context.TODO(),
		partitionKey:     aws.String("_id"),
		sortKey:          &[]string{},
		billingMode:      aws.String("PAY_PER_REQUEST"),
		deleteProtection: aws.Bool(false),
		provisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		sSESpecification: &types.SSESpecification{},
		tableClass:       aws.String("STANDARD"),
		tags:             &[]types.Tag{},
	}

	for _, opt := range opts {
		opt(dms)
	}

	return dms
}

// Get table list
func (d *DynamoDBMS) ListTables() ([]string, error) {
	tables, err := d.client.ListTables(d.ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return []string{}, err
	}
	return tables.TableNames, nil
}

// DynamoDB has no concept of multiple databases per instance, so this
// returns an empty list.
func (d *DynamoDBMS) ListDatabases() ([]string, error) {
	return []string{}, nil
}

// Delete table
func (d *DynamoDBMS) DeleteTables(tableName string) error {
	_, err := d.client.DeleteTable(d.ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	return err
}

// Create table
func (d *DynamoDBMS) CreateTable(tableName string) error {
	AD, KS := getAttrNSchema(*d.partitionKey, *d.sortKey...)

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions:      AD,
		KeySchema:                 KS,
		TableName:                 aws.String(tableName),
		BillingMode:               types.BillingMode(*d.billingMode),
		DeletionProtectionEnabled: d.deleteProtection,
		SSESpecification:          d.sSESpecification,
		TableClass:                types.TableClass(*d.tableClass),
		Tags:                      *d.tags,
	}
	// DynamoDB rejects ProvisionedThroughput on PAY_PER_REQUEST tables.
	if types.BillingMode(*d.billingMode) == types.BillingModeProvisioned {
		input.ProvisionedThroughput = d.provisionedThroughput
	}

	_, err := d.client.CreateTable(d.ctx, input)

	for {
		describeResp, err := d.client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return err
		}

		tableStatus := describeResp.Table.TableStatus

		if tableStatus == types.TableStatusActive {
			break
		}

		time.Sleep(5 * time.Second)
	}

	return err
}

// partition & sort
func getAttrNSchema(partition string, sort ...string) ([]types.AttributeDefinition, []types.KeySchemaElement) {
	var AD []types.AttributeDefinition
	var KS []types.KeySchemaElement

	AD = append(AD, types.AttributeDefinition{
		AttributeName: aws.String(partition),
		AttributeType: types.ScalarAttributeTypeS,
	})

	KS = append(KS, types.KeySchemaElement{
		AttributeName: aws.String(partition),
		KeyType:       types.KeyTypeHash,
	})

	for _, key := range sort {
		AD = append(AD, types.AttributeDefinition{
			AttributeName: aws.String(partition),
			AttributeType: types.ScalarAttributeTypeS,
		})

		KS = append(KS, types.KeySchemaElement{
			AttributeName: aws.String(key),
			KeyType:       types.KeyTypeRange,
		})
	}

	return AD, KS
}

// import table
func (d *DynamoDBMS) ImportTable(tableName string, srcData *[]map[string]interface{}) error {
	const batchSize = 25
	const maxConcurrency = 8

	writeRequests := make([]types.WriteRequest, 0, len(*srcData))
	for _, data := range *srcData {
		if _, ok := data["_id"]; !ok {
			data["_id"] = generateRandomString(10)
		}

		item, err := attributevalue.MarshalMap(data)
		if err != nil {
			return err
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
	}

	g, ctx := errgroup.WithContext(d.ctx)
	g.SetLimit(maxConcurrency)

	for i := 0; i < len(writeRequests); i += batchSize {
		batch := writeRequests[i:min(i+batchSize, len(writeRequests))]
		g.Go(func() error {
			return d.writeBatchWithRetry(ctx, tableName, batch)
		})
	}

	return g.Wait()
}

// clear table: removes all items but keeps the table, by scanning every
// item's key and issuing batched delete requests.
func (d *DynamoDBMS) ClearTable(tableName string) error {
	const batchSize = 25
	const maxConcurrency = 8

	keyAttrs := append([]string{*d.partitionKey}, *d.sortKey...)

	// Attribute names like "_id" aren't valid bare tokens in a
	// ProjectionExpression, so they must go through placeholders.
	exprNames := make(map[string]string, len(keyAttrs))
	placeholders := make([]string, len(keyAttrs))
	for i, attr := range keyAttrs {
		ph := fmt.Sprintf("#k%d", i)
		exprNames[ph] = attr
		placeholders[i] = ph
	}
	projExpr := strings.Join(placeholders, ", ")

	var writeRequests []types.WriteRequest
	var exclusiveStartKey map[string]types.AttributeValue
	for {
		out, err := d.client.Scan(d.ctx, &dynamodb.ScanInput{
			TableName:                aws.String(tableName),
			ProjectionExpression:     aws.String(projExpr),
			ExpressionAttributeNames: exprNames,
			ExclusiveStartKey:        exclusiveStartKey,
		})
		if err != nil {
			return err
		}

		for _, item := range out.Items {
			key := make(map[string]types.AttributeValue, len(keyAttrs))
			for _, attr := range keyAttrs {
				if v, ok := item[attr]; ok {
					key[attr] = v
				}
			}
			writeRequests = append(writeRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{Key: key},
			})
		}

		if out.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = out.LastEvaluatedKey
	}

	g, ctx := errgroup.WithContext(d.ctx)
	g.SetLimit(maxConcurrency)

	for i := 0; i < len(writeRequests); i += batchSize {
		batch := writeRequests[i:min(i+batchSize, len(writeRequests))]
		g.Go(func() error {
			return d.writeBatchWithRetry(ctx, tableName, batch)
		})
	}

	return g.Wait()
}

// writeBatchWithRetry sends a single (<=25 item) BatchWriteItem request,
// retrying items DynamoDB reports as unprocessed (e.g. due to throttling)
// with linear backoff, up to maxRetries.
func (d *DynamoDBMS) writeBatchWithRetry(ctx context.Context, tableName string, batch []types.WriteRequest) error {
	const maxRetries = 8

	requestItems := map[string][]types.WriteRequest{tableName: batch}
	for attempt := 0; len(requestItems) > 0; attempt++ {
		if attempt >= maxRetries {
			return fmt.Errorf("batch write to %s: %d items still unprocessed after %d retries (check the table's provisioned throughput)", tableName, len(requestItems[tableName]), maxRetries)
		}

		output, err := d.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: requestItems,
		})
		if err != nil {
			return err
		}

		requestItems = output.UnprocessedItems
		if len(requestItems) > 0 {
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
		}
	}

	return nil
}

func generateRandomString(length int) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// export table
func (d *DynamoDBMS) ExportTable(tableName string, dstData *[]map[string]interface{}) error {
	param := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	scan, err := d.client.Scan(context.TODO(), param)
	if err != nil {
		return err
	}

	err = attributevalue.UnmarshalListOfMaps(scan.Items, dstData)
	if err != nil {
		return err
	}

	removeIDKey(dstData)

	return nil
}

func removeIDKey(data *[]map[string]interface{}) {
	for _, item := range *data {
		delete(item, "_id")
	}
}
