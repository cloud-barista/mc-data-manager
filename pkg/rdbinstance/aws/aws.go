package aws

import (
	"context"
	"errors"
	"fmt"
	"sort"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
)

// AWSProvider implements rdbinstance.Provider for AWS RDS.
type AWSProvider struct {
	client *rds.Client
	region string
}

// New builds an AWS RDS provider from static credentials and a region.
func New(accessKey, secretKey, region string) (rdbinstance.Provider, error) {
	client, err := config.NewAWSRDBClient(accessKey, secretKey, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create RDS client: %w", err)
	}
	return &AWSProvider{client: client, region: region}, nil
}

// buildCreateInput maps a CSP-agnostic CreateSpec to an RDS CreateDBInstanceInput.
// PubliclyAccessible is fixed to true so the instance is reachable for testing.
func buildCreateInput(spec rdbinstance.CreateSpec) *rds.CreateDBInstanceInput {
	return &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: awssdk.String(spec.InstanceID),
		DBInstanceClass:      awssdk.String(spec.InstanceClass),
		Engine:               awssdk.String(spec.Engine),
		EngineVersion:        awssdk.String(spec.EngineVersion),
		MasterUsername:       awssdk.String(spec.MasterUsername),
		MasterUserPassword:   awssdk.String(spec.MasterPassword),
		AllocatedStorage:     awssdk.Int32(spec.AllocatedStorage),
		PubliclyAccessible:   awssdk.Bool(true),
	}
}

// CreateInstance provisions a new RDS instance and returns its initial state.
func (p *AWSProvider) CreateInstance(ctx context.Context, spec rdbinstance.CreateSpec) (models.DBInstance, error) {
	out, err := p.client.CreateDBInstance(ctx, buildCreateInput(spec))
	if err != nil {
		return models.DBInstance{}, fmt.Errorf("failed to create RDS instance: %w", err)
	}
	if out.DBInstance == nil {
		return models.DBInstance{}, fmt.Errorf("RDS create returned no instance")
	}
	return toDBInstance(*out.DBInstance, p.region), nil
}

// supportedEngines is the set of DB engines exposed by the metadata APIs.
var supportedEngines = []string{"mysql", "mariadb"}

// toEngineVersions extracts engine versions from a DescribeDBEngineVersions
// response, tagging each with the engine it was queried for.
func toEngineVersions(in []types.DBEngineVersion, engine string) []models.DBEngineVersion {
	out := make([]models.DBEngineVersion, 0, len(in))
	for _, v := range in {
		out = append(out, models.DBEngineVersion{
			Engine:        engine,
			EngineVersion: awssdk.ToString(v.EngineVersion),
		})
	}
	return out
}

// distinctInstanceClasses returns the unique, sorted set of DB instance class
// names from orderable option results.
func distinctInstanceClasses(in []types.OrderableDBInstanceOption) []string {
	seen := make(map[string]struct{}, len(in))
	for _, opt := range in {
		if c := awssdk.ToString(opt.DBInstanceClass); c != "" {
			seen[c] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// ListEngineVersions returns available versions for the supported engines, merged.
func (p *AWSProvider) ListEngineVersions(ctx context.Context) ([]models.DBEngineVersion, error) {
	var out []models.DBEngineVersion
	for _, engine := range supportedEngines {
		resp, err := p.client.DescribeDBEngineVersions(ctx, &rds.DescribeDBEngineVersionsInput{
			Engine: awssdk.String(engine),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe %s engine versions: %w", engine, err)
		}
		out = append(out, toEngineVersions(resp.DBEngineVersions, engine)...)
	}
	return out, nil
}

// ListInstanceClasses returns the orderable instance classes for engine+version.
func (p *AWSProvider) ListInstanceClasses(ctx context.Context, engine, engineVersion string) ([]string, error) {
	var options []types.OrderableDBInstanceOption
	paginator := rds.NewDescribeOrderableDBInstanceOptionsPaginator(p.client, &rds.DescribeOrderableDBInstanceOptionsInput{
		Engine:        awssdk.String(engine),
		EngineVersion: awssdk.String(engineVersion),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe orderable instance options: %w", err)
		}
		options = append(options, page.OrderableDBInstanceOptions...)
	}
	return distinctInstanceClasses(options), nil
}

// buildDeleteInput maps an instance identifier to an RDS DeleteDBInstanceInput.
// SkipFinalSnapshot is fixed to true so deletion does not require a snapshot id.
func buildDeleteInput(instanceID string) *rds.DeleteDBInstanceInput {
	return &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: awssdk.String(instanceID),
		SkipFinalSnapshot:    awssdk.Bool(true),
	}
}

// DeleteInstance deletes an RDS instance and returns its (deleting) state.
func (p *AWSProvider) DeleteInstance(ctx context.Context, instanceID string) (models.DBInstance, error) {
	out, err := p.client.DeleteDBInstance(ctx, buildDeleteInput(instanceID))
	if err != nil {
		if errors.As(err, new(*types.DBInstanceNotFoundFault)) {
			return models.DBInstance{Provider: "aws", InstanceID: instanceID, Status: "deleted", Region: p.region}, nil
		}
		return models.DBInstance{}, fmt.Errorf("failed to delete RDS instance: %w", err)
	}
	if out.DBInstance == nil {
		return models.DBInstance{}, fmt.Errorf("RDS delete returned no instance")
	}
	return toDBInstance(*out.DBInstance, p.region), nil
}

// ListInstances returns all RDS instances in the provider's region.
func (p *AWSProvider) ListInstances(ctx context.Context) ([]models.DBInstance, error) {
	var instances []types.DBInstance
	paginator := rds.NewDescribeDBInstancesPaginator(p.client, &rds.DescribeDBInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
		}
		instances = append(instances, page.DBInstances...)
	}
	return toDBInstances(instances, p.region), nil
}

// toDBInstance converts a single AWS RDS DBInstance into the CSP-agnostic model.
func toDBInstance(db types.DBInstance, region string) models.DBInstance {
	inst := models.DBInstance{
		Provider:      "aws",
		InstanceID:    awssdk.ToString(db.DBInstanceIdentifier),
		Name:          awssdk.ToString(db.DBInstanceIdentifier),
		Engine:        awssdk.ToString(db.Engine),
		EngineVersion: awssdk.ToString(db.EngineVersion),
		Status:        awssdk.ToString(db.DBInstanceStatus),
		InstanceClass: awssdk.ToString(db.DBInstanceClass),
		Region:        region,
	}
	if db.Endpoint != nil {
		if awssdk.ToBool(db.PubliclyAccessible) {
			inst.Endpoint = awssdk.ToString(db.Endpoint.Address)
		} else {
			inst.Endpoint = "-"
		}
		inst.Port = awssdk.ToInt32(db.Endpoint.Port)
	}
	return inst
}

// toDBInstances converts AWS RDS DBInstance descriptions into the CSP-agnostic
// models.DBInstance representation.
func toDBInstances(in []types.DBInstance, region string) []models.DBInstance {
	out := make([]models.DBInstance, 0, len(in))
	for _, db := range in {
		out = append(out, toDBInstance(db, region))
	}
	return out
}
