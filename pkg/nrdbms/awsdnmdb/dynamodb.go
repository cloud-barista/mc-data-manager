package awsdnmdb

import (
	"context"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

type DynamoDBMS struct {
	provider utils.Provider
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
		provider:         utils.AWS,
		region:           region,
		client:           client,
		ctx:              context.TODO(),
		partitionKey:     aws.String("_id"),
		sortKey:          &[]string{},
		billingMode:      aws.String("PROVISIONED"),
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

	_, err := d.client.CreateTable(d.ctx,
		&dynamodb.CreateTableInput{
			AttributeDefinitions:      AD,
			KeySchema:                 KS,
			TableName:                 aws.String(tableName),
			BillingMode:               types.BillingMode(*d.billingMode),
			DeletionProtectionEnabled: d.deleteProtection,
			ProvisionedThroughput:     d.provisionedThroughput,
			SSESpecification:          d.sSESpecification,
			TableClass:                types.TableClass(*d.tableClass),
			Tags:                      *d.tags,
		},
	)

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
	for _, data := range *srcData {
		if _, ok := data["_id"]; !ok {
			data["_id"] = generateRandomString(10)
		}

		item, err := attributevalue.MarshalMap(data)
		if err != nil {
			return err
		}

		input := &dynamodb.PutItemInput{
			Item:      item,
			TableName: aws.String(tableName),
		}

		_, err = d.client.PutItem(d.ctx, input)
		if err != nil {
			return err
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
