package alibaba

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	alibabards "github.com/alibabacloud-go/rds-20140815/v8/client"
	"github.com/alibabacloud-go/tea/tea"
	alibabavpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	alibabacommon "github.com/cloud-barista/mc-data-manager/pkg/alibaba"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
	"github.com/rs/zerolog/log"
)

// supportedEngines is the set of DB engines exposed by this provider, in
// Alibaba's native casing (also used as-is in responses — no lowercase mapping).
var supportedEngines = []string{"MySQL", "MariaDB"}

// instanceCategory is fixed to single-node Basic edition, matching CreateInstance.
const instanceCategory = "Basic"

// AlibabaProvider implements rdbinstance.Provider for Alibaba RDS.
type AlibabaProvider struct {
	client    *alibabards.Client
	vpcClient *alibabavpc.Client
	region    string
}

// New builds an Alibaba RDS provider from static credentials and a region.
func New(accessKey, secretKey, region string) (rdbinstance.Provider, error) {
	client, err := config.NewAlibabaRDBClient(region, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Alibaba RDS client: %w", err)
	}
	vpcClient, err := config.NewAlibabaVPCClient(region, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Alibaba VPC client: %w", err)
	}
	return &AlibabaProvider{client: client, vpcClient: vpcClient, region: region} , nil
}

func buildCreateRequest(spec rdbinstance.CreateSpec, region string, vsw alibabacommon.VSwitchInfo) *alibabards.CreateDBInstanceRequest {
	return &alibabards.CreateDBInstanceRequest{
		Engine:                tea.String(spec.Engine),
		EngineVersion:         tea.String(spec.EngineVersion),
		DBInstanceClass:       tea.String(spec.InstanceClass),
		DBInstanceStorage:     tea.Int32(spec.AllocatedStorage),
		RegionId:              tea.String(region),
		DBInstanceDescription: tea.String(spec.InstanceID),
		InstanceNetworkType:   tea.String("VPC"),
		VPCId:                 tea.String(vsw.VpcID),
		VSwitchId:             tea.String(vsw.VSwitchID),
		DBInstanceNetType:     tea.String("Internet"),
		PayType:               tea.String("Postpaid"),
		SecurityIPList:        tea.String(config.OutboundIP),
		Category:              tea.String("Basic"),
	}
}

func toCreatedDBInstance(spec rdbinstance.CreateSpec, body *alibabards.CreateDBInstanceResponseBody, region string) models.DBInstance {
	return models.DBInstance{
		Provider:      "alibaba",
		InstanceID:    tea.StringValue(body.DBInstanceId),
		Name:          spec.InstanceID,
		Engine:        spec.Engine,
		EngineVersion: spec.EngineVersion,
		Status:        "creating",
		Endpoint:      tea.StringValue(body.ConnectionString),
		Port:          parsePort(body.Port),
		Region:        region,
		InstanceClass: spec.InstanceClass,
	}
}

func (p *AlibabaProvider) CreateInstance(ctx context.Context, spec rdbinstance.CreateSpec) (models.DBInstance, error) {
	vsw, err := alibabacommon.ResolveVSwitch(p.vpcClient, p.region)
	if err != nil {
		return models.DBInstance{}, err
	}

	resp, err := p.client.CreateDBInstance(buildCreateRequest(spec, p.region, vsw))
	if err != nil {
		return models.DBInstance{}, fmt.Errorf("failed to create Alibaba RDS instance: %w", err)
	}
	if resp == nil || resp.Body == nil || tea.StringValue(resp.Body.DBInstanceId) == "" {
		return models.DBInstance{}, fmt.Errorf("Alibaba RDS create returned no instance id")
	}

	instance := toCreatedDBInstance(spec, resp.Body, p.region)

	// 생성 직후 조회 시 List에 존재하지 않는 지연 문제 예방
	p.waitForInstanceVisible(ctx, instance.InstanceID, 30*time.Second, 2*time.Second)

	// Account + public endpoint require the instance to be Running (minutes away).
	go p.provisionInBackground(instance.InstanceID, spec.MasterUsername, spec.MasterPassword)

	return instance, nil
}

func (p *AlibabaProvider) waitForInstanceVisible(ctx context.Context, instanceID string, timeout, interval time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		if exists, err := p.InstanceExists(ctx, instanceID); err == nil && exists {
			return
		}
		if time.Now().After(deadline) {
			log.Warn().Str("instanceID", instanceID).Msg("instance not visible via DescribeDBInstances within timeout after creation")
			return
		}
		time.Sleep(interval)
	}
}

// parsePort converts an Alibaba port string into int32 (0 if empty/invalid).
func parsePort(port *string) int32 {
	v, err := strconv.Atoi(tea.StringValue(port))
	if err != nil {
		return 0
	}
	return int32(v)
}

func toSnakeCase(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := runes[i-1]
				prevLowerOrDigit := (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9')
				prevUpper := prev >= 'A' && prev <= 'Z'
				nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
				if prevLowerOrDigit || (prevUpper && nextLower) {
					b.WriteByte('_')
				}
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normalizeStatus converts an Alibaba DBInstanceStatus to snake_case, mapping
// "Running" to AWS's "available".
func normalizeStatus(s string) string {
	out := toSnakeCase(s)
	if out == "running" {
		return "available"
	}
	return out
}

// pickEndpoint returns the connection string + port of the public endpoint if one
// exists, otherwise the first non-public (private/inner) endpoint.
func pickEndpoint(netInfos []*alibabards.DescribeDBInstanceNetInfoResponseBodyDBInstanceNetInfosDBInstanceNetInfo) (string, int32) {
	var fallbackPort int32
	var haveFallback bool
	for _, ni := range netInfos {
		if tea.StringValue(ni.IPType) == "Public" {
			return tea.StringValue(ni.ConnectionString), parsePort(ni.Port)
		}
		if !haveFallback {
			fallbackPort = parsePort(ni.Port)
			haveFallback = true
		}
	}
	if haveFallback {
		return "-", fallbackPort
	}
	return "-", 0
}

func toListedDBInstance(item *alibabards.DescribeDBInstancesResponseBodyItemsDBInstance, endpoint string, port int32, region string) models.DBInstance {
	name := tea.StringValue(item.DBInstanceDescription)
	if name == "" {
		name = tea.StringValue(item.DBInstanceId)
	}
	return models.DBInstance{
		Provider:      "alibaba",
		InstanceID:    tea.StringValue(item.DBInstanceId),
		Name:          name,
		Engine:        tea.StringValue(item.Engine),
		EngineVersion: tea.StringValue(item.EngineVersion),
		Status:        normalizeStatus(tea.StringValue(item.DBInstanceStatus)),
		Endpoint:      endpoint,
		Port:          port,
		Region:        region,
		InstanceClass: tea.StringValue(item.DBInstanceClass),
	}
}

// instanceEndpoint returns the public-or-private endpoint+port for one instance.
func (p *AlibabaProvider) instanceEndpoint(instanceID string) (string, int32, error) {
	resp, err := p.client.DescribeDBInstanceNetInfo(&alibabards.DescribeDBInstanceNetInfoRequest{
		DBInstanceId: tea.String(instanceID),
	})
	if err != nil {
		return "-", 0, err
	}
	if resp == nil || resp.Body == nil || resp.Body.DBInstanceNetInfos == nil {
		return "-", 0, nil
	}
	ep, port := pickEndpoint(resp.Body.DBInstanceNetInfos.DBInstanceNetInfo)
	return ep, port, nil
}

// ListInstances returns all RDS instances in the region. endpoint+port require a
// per-instance DescribeDBInstanceNetInfo call (N+1).
func (p *AlibabaProvider) ListInstances(ctx context.Context) ([]models.DBInstance, error) {
	const pageSize int32 = 100
	out := []models.DBInstance{}
	for pageNumber := int32(1); ; pageNumber++ {
		resp, err := p.client.DescribeDBInstances(&alibabards.DescribeDBInstancesRequest{
			RegionId:   tea.String(p.region),
			PageNumber: tea.Int32(pageNumber),
			PageSize:   tea.Int32(pageSize),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe DB instances: %w", err)
		}
		if resp == nil || resp.Body == nil || resp.Body.Items == nil {
			break
		}
		items := resp.Body.Items.DBInstance
		for _, item := range items {
			id := tea.StringValue(item.DBInstanceId)
			ep, port, err := p.instanceEndpoint(id)
			if err != nil {
				return nil, fmt.Errorf("failed to describe net info for %s: %w", id, err)
			}
			inst := toListedDBInstance(item, ep, port, p.region)
			out = append(out, inst)
		}
		if int32(len(items)) < pageSize {
			break
		}
	}
	return out, nil
}

func (p *AlibabaProvider) InstanceExists(ctx context.Context, instanceID string) (bool, error) {
	resp, err := p.client.DescribeDBInstances(&alibabards.DescribeDBInstancesRequest{
		RegionId:     tea.String(p.region),
		DBInstanceId: tea.String(instanceID),
	})
	if err != nil {
		return false, fmt.Errorf("failed to describe DB instance: %w", err)
	}
	if resp == nil || resp.Body == nil || resp.Body.Items == nil {
		return false, nil
	}
	return len(resp.Body.Items.DBInstance) > 0, nil
}

func (p *AlibabaProvider) DeleteInstance(ctx context.Context, instanceID string) (models.DBInstance, error) {
	_, err := p.client.DeleteDBInstance(&alibabards.DeleteDBInstanceRequest{
		DBInstanceId:       tea.String(instanceID),
		ReleasedKeepPolicy: tea.String("None"),
	})
	if err != nil {
		var sdkErr *tea.SDKError
		if errors.As(err, &sdkErr) && sdkErr.StatusCode != nil && *sdkErr.StatusCode == 404 {
			return models.DBInstance{Provider: "alibaba", InstanceID: instanceID, Status: "deleted", Region: p.region}, nil
		}
		return models.DBInstance{}, fmt.Errorf("failed to delete Alibaba RDS instance: %w", err)
	}
	return models.DBInstance{
		Provider:   "alibaba",
		InstanceID: instanceID,
		Name:       instanceID,
		Status:     "deleting",
		Region:     p.region,
	}, nil
}

// distinctSorted returns the unique, sorted set of non-empty strings.
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

// extractEngineVersions collects the engine versions for the given Alibaba engine
// from a DescribeAvailableZones response (deduplicated, engine tag kept as-is).
func extractEngineVersions(body *alibabards.DescribeAvailableZonesResponseBody, engine string) []models.DBEngineVersion {
	seen := make(map[string]struct{})
	var out []models.DBEngineVersion
	for _, zone := range body.AvailableZones {
		for _, eng := range zone.SupportedEngines {
			if tea.StringValue(eng.Engine) != engine {
				continue
			}
			for _, ver := range eng.SupportedEngineVersions {
				v := tea.StringValue(ver.Version)
				if v == "" {
					continue
				}
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				out = append(out, models.DBEngineVersion{Engine: engine, EngineVersion: v})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EngineVersion < out[j].EngineVersion })
	return out
}

// zoneIDs returns the distinct, sorted single-AZ zone ids from a DescribeAvailableZones response.
func zoneIDs(body *alibabards.DescribeAvailableZonesResponseBody) []string {
	ids := make([]string, 0, len(body.AvailableZones))
	for _, zone := range body.AvailableZones {
		id := tea.StringValue(zone.ZoneId)
		if alibabacommon.IsUsableZone(id) {
			ids = append(ids, id)
		}
	}
	return distinctSorted(ids)
}

func (p *AlibabaProvider) ListEngineVersions(ctx context.Context) ([]models.DBEngineVersion, error) {
	var out []models.DBEngineVersion
	for _, engine := range supportedEngines {
		resp, err := p.client.DescribeAvailableZones(&alibabards.DescribeAvailableZonesRequest{
			RegionId: tea.String(p.region),
			Engine:   tea.String(engine),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe available zones for %s: %w", engine, err)
		}
		if resp != nil && resp.Body != nil {
			out = append(out, extractEngineVersions(resp.Body, engine)...)
		}
	}
	return out, nil
}

// storageTypes is the fixed set of storage types probed for instance classes.
// Alibaba supports only these three; we brute-force (zone × storageType).
var storageTypes = []string{"cloud_ssd", "cloud_essd"}

func (p *AlibabaProvider) ListInstanceClasses(ctx context.Context, engine, engineVersion string) ([]string, error) {
	zonesResp, err := p.client.DescribeAvailableZones(&alibabards.DescribeAvailableZonesRequest{
		RegionId:      tea.String(p.region),
		Engine:        tea.String(engine),
		EngineVersion: tea.String(engineVersion),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe available zones: %w", err)
	}
	if zonesResp == nil || zonesResp.Body == nil {
		return nil, nil
	}
	zones := zoneIDs(zonesResp.Body)
	if len(zones) == 0 {
		return []string{}, nil
	}
	zone := zones[0]

	var classes []string
	for _, storageType := range storageTypes {
		clsResp, err := p.client.DescribeAvailableClasses(&alibabards.DescribeAvailableClassesRequest{
			RegionId:              tea.String(p.region),
			ZoneId:                tea.String(zone),
			Engine:                tea.String(engine),
			EngineVersion:         tea.String(engineVersion),
			Category:              tea.String(instanceCategory),
			DBInstanceStorageType: tea.String(storageType),
			InstanceChargeType:    tea.String("Postpaid"),
		})
		if err != nil {
			if strings.Contains(err.Error(), "InvalidCondition.NotFound") {
				continue
			}
			return nil, fmt.Errorf("failed to describe available classes (zone=%s, storage=%s): %w", zone, storageType, err)
		}
		if clsResp == nil || clsResp.Body == nil {
			continue
		}
		for _, c := range clsResp.Body.DBInstanceClasses {
			classes = append(classes, tea.StringValue(c.DBInstanceClass))
		}
	}

	return distinctSorted(classes), nil
}

func (p *AlibabaProvider) CreateAccount(instanceID, username, password string) error {
	_, err := p.client.CreateAccount(&alibabards.CreateAccountRequest{
		DBInstanceId:    tea.String(instanceID),
		AccountName:     tea.String(username),
		AccountPassword: tea.String(password),
		AccountType:     tea.String("Super"),
	})
	return err
}