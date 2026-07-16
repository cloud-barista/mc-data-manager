// Package alibaba implements nrdbinstance.Provider for Alibaba ApsaraDB for MongoDB (DDS).
package alibaba

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	dds "github.com/alibabacloud-go/dds-20151201/client"
	"github.com/alibabacloud-go/tea/tea"
	alibabavpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	alibabacommon "github.com/cloud-barista/mc-data-manager/pkg/alibaba"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance"
)

const (
	dbType        = "normal"
	chargeType    = "PostPaid"
	engine        = "MongoDB"
	storageEngine = "WiredTiger"
)

// AlibabaProvider implements nrdbinstance.Provider for Alibaba MongoDB.
type AlibabaProvider struct {
	client    *dds.Client
	vpcClient *alibabavpc.Client
	region    string
}

// New builds an Alibaba MongoDB provider from static credentials and a region.
func New(accessKey, secretKey, region string) (nrdbinstance.Provider, error) {
	client, err := config.NewAlibabaDDSClient(region, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Alibaba DDS client: %w", err)
	}
	vpcClient, err := config.NewAlibabaVPCClient(region, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Alibaba VPC client: %w", err)
	}
	return &AlibabaProvider{client: client, vpcClient: vpcClient, region: region}, nil
}

// ListInstances returns all DDS MongoDB instances in the region.
// Endpoint and port require a per-instance DescribeReplicaSetRole call (N+1).
func (p *AlibabaProvider) ListInstances(_ context.Context) ([]models.NRDBInstance, error) {
	const pageSize int32 = 30
	out := []models.NRDBInstance{}
	for pageNumber := int32(1); ; pageNumber++ {
		resp, err := p.client.DescribeDBInstances(&dds.DescribeDBInstancesRequest{
			RegionId:   tea.String(p.region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe MongoDB instances: %w", err)
		}
		if resp == nil || resp.Body == nil || resp.Body.DBInstances == nil {
			break
		}
		items := resp.Body.DBInstances.DBInstance
		for _, item := range items {
			ep, port, err := p.instanceEndpoint(tea.StringValue(item.DBInstanceId))
			if err != nil {
				return nil, fmt.Errorf("failed to describe endpoint for %s: %w", tea.StringValue(item.DBInstanceId), err)
			}
			out = append(out, toNRDBInstance(item, ep, port, p.region))
		}
		if int32(len(items)) < pageSize {
			break
		}
	}
	return out, nil
}

// instanceEndpoint returns the endpoint and port for one instance.
// Private (VPC) is used as fallback; public overrides if present.
func (p *AlibabaProvider) instanceEndpoint(instanceID string) (string, int32, error) {
	resp, err := p.client.DescribeReplicaSetRole(&dds.DescribeReplicaSetRoleRequest{
		DBInstanceId: tea.String(instanceID),
	})
	if err != nil {
		return "", 0, err
	}
	if resp == nil || resp.Body == nil || resp.Body.ReplicaSets == nil {
		return "", 0, nil
	}
	ep, port := pickEndpoint(resp.Body.ReplicaSets.ReplicaSet)
	return ep, port, nil
}

// pickEndpoint returns the best endpoint from a replica set role response.
// VPC endpoint is used as fallback; public endpoint overrides if present.
func pickEndpoint(replicaSets []*dds.DescribeReplicaSetRoleResponseBodyReplicaSetsReplicaSet) (string, int32) {
	var privPort int32
	for _, rs := range replicaSets {
		if tea.StringValue(rs.NetworkType) == "Public" && tea.StringValue(rs.ReplicaSetRole) == "Primary" {
			return tea.StringValue(rs.ConnectionDomain), parsePort(rs.ConnectionPort)
		}
		if tea.StringValue(rs.NetworkType) == "VPC" {
			privPort = parsePort(rs.ConnectionPort)
		}
	}
	return "-", privPort
}

// toNRDBInstance converts a DescribeDBInstances item to the CSP-agnostic model.
func toNRDBInstance(item *dds.DescribeDBInstancesResponseBodyDBInstancesDBInstance, endpoint string, port int32, region string) models.NRDBInstance {
	name := tea.StringValue(item.DBInstanceDescription)
	if name == "" {
		name = tea.StringValue(item.DBInstanceId)
	}
	return models.NRDBInstance{
		Provider:      "alibaba",
		InstanceID:    tea.StringValue(item.DBInstanceId),
		Name:          name,
		Status:        normalizeStatus(tea.StringValue(item.DBInstanceStatus)),
		Region:        region,
		EngineVersion: tea.StringValue(item.EngineVersion),
		Endpoint:      endpoint,
		Port:          port,
	}
}

// normalizeStatus lowercases an Alibaba DBInstanceStatus, mapping "running" to "available".
func normalizeStatus(s string) string {
	lower := strings.ToLower(s)
	if lower == "running" {
		return "available"
	}
	return lower
}

// parsePort converts a port string to int32 (0 if empty or invalid).
func parsePort(port *string) int32 {
	v, err := strconv.Atoi(tea.StringValue(port))
	if err != nil {
		return 0
	}
	return int32(v)
}

// CreateInstance provisions a new Alibaba MongoDB (DDS) replica-set instance.
func (p *AlibabaProvider) CreateInstance(_ context.Context, spec nrdbinstance.CreateSpec) (models.NRDBInstance, error) {
	if spec.MasterUsername != "root" {
		return models.NRDBInstance{}, fmt.Errorf("Alibaba MongoDB only supports 'root' as masterUsername")
	}

	vsw, err := alibabacommon.ResolveVSwitch(p.vpcClient, p.region)
	if err != nil {
		return models.NRDBInstance{}, err
	}

	resp, err := p.client.CreateDBInstance(&dds.CreateDBInstanceRequest{
		RegionId:              tea.String(p.region),
		ZoneId:                tea.String(vsw.ZoneID),
		Engine:                tea.String(engine),
		EngineVersion:         tea.String(spec.EngineVersion),
		DBInstanceClass:       tea.String(spec.InstanceClass),
		DBInstanceStorage:     tea.Int32(spec.AllocatedStorage),
		DBInstanceDescription: tea.String(spec.InstanceID),
		ChargeType:            tea.String(chargeType),
		NetworkType:           tea.String("VPC"),
		VpcId:                 tea.String(vsw.VpcID),
		VSwitchId:             tea.String(vsw.VSwitchID),
		StorageEngine:         tea.String(storageEngine),
		AccountPassword:       tea.String(spec.MasterPassword),
		SecurityIPList:        tea.String(config.OutboundIP),
	})
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("failed to create Alibaba MongoDB instance: %w", err)
	}
	if resp == nil || resp.Body == nil || tea.StringValue(resp.Body.DBInstanceId) == "" {
		return models.NRDBInstance{}, fmt.Errorf("Alibaba MongoDB create returned no instance id")
	}

	instance := models.NRDBInstance{
		Provider:   "alibaba",
		InstanceID: tea.StringValue(resp.Body.DBInstanceId),
		Name:       spec.InstanceID,
		Status:     "creating",
		Region:     p.region,
	}

	go p.provisionInBackground(instance.InstanceID)

	return instance, nil
}

// DeleteInstance deletes a DDS MongoDB instance. The response has no instance
// details, so the returned model is constructed locally.
func (p *AlibabaProvider) DeleteInstance(_ context.Context, instanceID string) (models.NRDBInstance, error) {
	_, err := p.client.DeleteDBInstance(&dds.DeleteDBInstanceRequest{
		DBInstanceId: tea.String(instanceID),
	})
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("failed to delete Alibaba MongoDB instance: %w", err)
	}
	return models.NRDBInstance{
		Provider:   "alibaba",
		InstanceID: instanceID,
		Name:       instanceID,
		Status:     "deleting",
		Region:     p.region,
	}, nil
}

// distinctSorted returns the unique sorted set of non-empty strings.
func distinctSorted(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	for _, s := range items {
		if s != "" {
			seen[s] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ListEngineVersions returns available MongoDB engine versions for the region
func (p *AlibabaProvider) ListEngineVersions(_ context.Context) ([]models.NRDBEngineVersion, error) {
	resp, err := p.client.DescribeAvailableResource(&dds.DescribeAvailableResourceRequest{
		RegionId:           tea.String(p.region),
		DbType:             tea.String(dbType),
		InstanceChargeType: tea.String(chargeType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe available resource: %w", err)
	}
	if resp == nil || resp.Body == nil || resp.Body.SupportedDBTypes == nil {
		return []models.NRDBEngineVersion{}, nil
	}

	seen := make(map[string]struct{})
	var out []models.NRDBEngineVersion

	for _, dbTyp := range resp.Body.SupportedDBTypes.SupportedDBType {
		if dbTyp.AvailableZones == nil {
			continue
		}
		for _, zone := range dbTyp.AvailableZones.AvailableZone {
			if zone.SupportedEngineVersions == nil {
				continue
			}
			for _, ver := range zone.SupportedEngineVersions.SupportedEngineVersion {
				v := tea.StringValue(ver.Version)
				if v == "" {
					continue
				}
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				out = append(out, models.NRDBEngineVersion{EngineVersion: v})
			}
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].EngineVersion < out[j].EngineVersion })
	return out, nil
}

// firstUsableZone extracts the first single-AZ zone ID from a DescribeAvailableResource response.
// Multi-AZ aggregated zones (containing "MAZ" or "(") are skipped, matching the RDB Alibaba pattern.
func firstUsableZone(resp *dds.DescribeAvailableResourceResponse) string {
	if resp == nil || resp.Body == nil || resp.Body.SupportedDBTypes == nil {
		return ""
	}
	for _, dbTyp := range resp.Body.SupportedDBTypes.SupportedDBType {
		if dbTyp.AvailableZones == nil {
			continue
		}
		for _, zone := range dbTyp.AvailableZones.AvailableZone {
			id := tea.StringValue(zone.ZoneId)
			if id != "" && alibabacommon.IsUsableZone(id) {
				return id
			}
		}
	}
	return ""
}

// ListInstanceClasses returns orderable MongoDB instance classes for the given engine version.
// It first resolves the first available single-AZ zone, then re-queries with that zone to get
func (p *AlibabaProvider) ListInstanceClasses(_ context.Context, engineVersion string) ([]string, error) {
	// Step 1: resolve first usable zone for the region.
	regionResp, err := p.client.DescribeAvailableResource(&dds.DescribeAvailableResourceRequest{
		RegionId:           tea.String(p.region),
		DbType:             tea.String(dbType),
		InstanceChargeType: tea.String(chargeType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe available resource: %w", err)
	}
	zoneID := firstUsableZone(regionResp)
	if zoneID == "" {
		return []string{}, nil
	}

	// Step 2: query instance classes for the resolved zone and given engine version.
	resp, err := p.client.DescribeAvailableResource(&dds.DescribeAvailableResourceRequest{
		RegionId:           tea.String(p.region),
		ZoneId:             tea.String(zoneID),
		DbType:             tea.String(dbType),
		InstanceChargeType: tea.String(chargeType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe available resource for zone %s: %w", zoneID, err)
	}
	if resp == nil || resp.Body == nil || resp.Body.SupportedDBTypes == nil {
		return []string{}, nil
	}

	dbTypes := resp.Body.SupportedDBTypes.SupportedDBType
	if len(dbTypes) == 0 || dbTypes[0].AvailableZones == nil {
		return []string{}, nil
	}
	zones := dbTypes[0].AvailableZones.AvailableZone
	if len(zones) == 0 || zones[0].SupportedEngineVersions == nil {
		return []string{}, nil
	}

	var classes []string
	for _, ver := range zones[0].SupportedEngineVersions.SupportedEngineVersion {
		if tea.StringValue(ver.Version) != engineVersion || ver.SupportedEngines == nil {
			continue
		}
		for _, eng := range ver.SupportedEngines.SupportedEngine {
			if eng.SupportedNodeTypes == nil {
				continue
			}
			for _, nt := range eng.SupportedNodeTypes.SupportedNodeType {
				if nt.AvailableResources == nil {
					continue
				}
				for _, res := range nt.AvailableResources.AvailableResource {
					classes = append(classes, tea.StringValue(res.InstanceClass))
				}
			}
		}
	}

	return distinctSorted(classes), nil
}
