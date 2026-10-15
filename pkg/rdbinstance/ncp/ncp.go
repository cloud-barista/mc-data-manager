// Package ncp implements rdbinstance.Provider for NCP Cloud DB for MySQL.
//
// NCP Cloud DB only supports MySQL. MariaDB is not available.
// DeleteInstance takes the CloudMysqlInstanceNo (numeric string returned by
// ListInstances / CreateInstance) as the instanceID.
package ncp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	vmysql "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vmysql"
	ncpvpc "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	ncpvserver "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	ncpcommon "github.com/cloud-barista/mc-data-manager/pkg/ncp"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
)

// NCPProvider implements rdbinstance.Provider for NCP Cloud DB for MySQL.
type NCPProvider struct {
	api        *vmysql.V2ApiService
	vpcApi     *ncpvpc.V2ApiService
	vserverApi *ncpvserver.V2ApiService
	region     string
	accessKey  string
	secretKey  string
}

// New builds an NCP Cloud DB provider from static credentials and a region.
func New(accessKey, secretKey, region string) (rdbinstance.Provider, error) {
	api, err := config.NewNCPCloudMysqlClient(accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create NCP Cloud MySQL client: %w", err)
	}
	vpcApi, err := config.NewNCPVPCClient(accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create NCP VPC client: %w", err)
	}
	vserverApi, err := config.NewNCPVServerClient(accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create NCP VServer client: %w", err)
	}
	return &NCPProvider{api: api, vpcApi: vpcApi, vserverApi: vserverApi, region: region, accessKey: accessKey, secretKey: secretKey}, nil
}

// normalizeStatus maps NCP CloudMysqlInstanceStatusName to the unified status.
func normalizeStatus(s string) string {
	switch s {
	case "running":
		return "available"
	// case "creating", "init", "setting":
	// 	return "creating"
	case "deleting", "terminating":
		return "deleting"
	default:
		return s
	}
}

func (p *NCPProvider) DeleteInstance(_ context.Context, instanceID string) (models.DBInstance, error) {
	resp, err := p.api.DeleteCloudMysqlInstance(&vmysql.DeleteCloudMysqlInstanceRequest{
		RegionCode:           ncloud.String(p.region),
		CloudMysqlInstanceNo: ncloud.String(instanceID),
	})
	if err != nil {
		if strings.Contains(err.Error(), "5001017") {
			return models.DBInstance{Provider: "ncp", InstanceID: instanceID, Status: "deleted", Region: p.region}, nil
		}
		return models.DBInstance{}, fmt.Errorf("failed to delete NCP Cloud MySQL instance: %w", err)
	}

	name := instanceID
	if resp != nil && len(resp.CloudMysqlInstanceList) > 0 {
		if sn := resp.CloudMysqlInstanceList[0].CloudMysqlServiceName; sn != nil {
			name = *sn
		}
	}

	return models.DBInstance{
		Provider:   "ncp",
		InstanceID: instanceID,
		Name:       name,
		Status:     "deleting",
		Region:     p.region,
	}, nil
}

func toDBInstance(inst *vmysql.CloudMysqlInstance, region string) models.DBInstance {
	instanceNo := ncloud.StringValue(inst.CloudMysqlInstanceNo)
	name := instanceNo
	if inst.CloudMysqlServiceName != nil && *inst.CloudMysqlServiceName != "" {
		name = *inst.CloudMysqlServiceName
	}
	status := ""
	if inst.CloudMysqlInstanceStatusName != nil {
		status = normalizeStatus(*inst.CloudMysqlInstanceStatusName)
	}
	var port int32
	if inst.CloudMysqlPort != nil {
		port = *inst.CloudMysqlPort
	}
	endpoint := "-"
	var instanceClass string
	if len(inst.CloudMysqlServerInstanceList) > 0 {
		srv := inst.CloudMysqlServerInstanceList[0]
		if ncloud.BoolValue(srv.IsPublicSubnet) && srv.PublicDomain != nil {
			endpoint = *srv.PublicDomain
		}
		instanceClass = ncloud.StringValue(srv.CloudMysqlProductCode)
	}
	return models.DBInstance{
		Provider:      "ncp",
		InstanceID:    instanceNo,
		Name:          name,
		Status:        status,
		Region:        region,
		Engine:        "mysql",
		EngineVersion: ncloud.StringValue(inst.EngineVersion),
		Endpoint:      endpoint,
		Port:          port,
		InstanceClass: instanceClass,
	}
}

func (p *NCPProvider) instanceDetail(instanceNo string) (*vmysql.CloudMysqlInstance, error) {
	resp, err := p.api.GetCloudMysqlInstanceDetail(&vmysql.GetCloudMysqlInstanceDetailRequest{
		RegionCode:           ncloud.String(p.region),
		CloudMysqlInstanceNo: ncloud.String(instanceNo),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe NCP Cloud MySQL instance %s: %w", instanceNo, err)
	}
	if resp == nil || len(resp.CloudMysqlInstanceList) == 0 {
		return nil, fmt.Errorf("no detail returned for NCP Cloud MySQL instance %s", instanceNo)
	}
	return resp.CloudMysqlInstanceList[0], nil
}

func (p *NCPProvider) ListInstances(_ context.Context) ([]models.DBInstance, error) {
	resp, err := p.api.GetCloudMysqlInstanceList(&vmysql.GetCloudMysqlInstanceListRequest{
		RegionCode: ncloud.String(p.region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP Cloud MySQL instances: %w", err)
	}
	out := []models.DBInstance{}
	for _, inst := range resp.CloudMysqlInstanceList {
		// Skip instances already being removed — matches NCP console behavior.
		if s := ncloud.StringValue(inst.CloudMysqlInstanceStatusName); s == "terminating" || s == "deleting" {
			continue
		}
		instanceNo := ncloud.StringValue(inst.CloudMysqlInstanceNo)
		detail, err := p.instanceDetail(instanceNo)
		if err != nil {
			// Instance may have been deleted between list and detail calls; skip it.
			continue
		}
		out = append(out, toDBInstance(detail, p.region))
	}
	return out, nil
}

func (p *NCPProvider) CreateInstance(_ context.Context, spec rdbinstance.CreateSpec) (models.DBInstance, error) {
	sub, err := ncpcommon.ResolveSubnet(p.vpcApi, p.vserverApi, p.region)
	if err != nil {
		return models.DBInstance{}, err
	}
	resp, err := p.api.CreateCloudMysqlInstance(&vmysql.CreateCloudMysqlInstanceRequest{
		RegionCode:                 ncloud.String(p.region),
		VpcNo:                      ncloud.String(sub.VpcNo),
		SubnetNo:                   ncloud.String(sub.SubnetNo),
		CloudMysqlServiceName:      ncloud.String(spec.InstanceID),
		CloudMysqlServerNamePrefix: ncloud.String(spec.InstanceID),
		CloudMysqlUserName:         ncloud.String(spec.MasterUsername),
		CloudMysqlUserPassword:     ncloud.String(spec.MasterPassword),
		HostIp:                     ncloud.String("%"),
		CloudMysqlDatabaseName:     ncloud.String("mcmp_default"),
		EngineVersionCode:          ncloud.String(spec.EngineVersion),
		CloudMysqlProductCode:      ncloud.String(spec.InstanceClass),
		IsBackup:                   ncloud.Bool(false),
		IsHa:                       ncloud.Bool(false),
	})
	if err != nil {
		return models.DBInstance{}, fmt.Errorf("failed to create NCP Cloud MySQL instance: %w", err)
	}
	instanceNo := ""
	status := "creating"
	if resp != nil && len(resp.CloudMysqlInstanceList) > 0 {
		inst := resp.CloudMysqlInstanceList[0]
		if inst.CloudMysqlInstanceNo != nil {
			instanceNo = *inst.CloudMysqlInstanceNo
		}
		if inst.CloudMysqlInstanceStatusName != nil {
			status = normalizeStatus(*inst.CloudMysqlInstanceStatusName)
		}
	}
	go p.provisionInBackground(instanceNo, sub.VpcNo)

	return models.DBInstance{
		Provider:      "ncp",
		InstanceID:    instanceNo,
		Name:          spec.InstanceID,
		Status:        status,
		Region:        p.region,
		Engine:        "mysql",
		EngineVersion: spec.EngineVersion,
		InstanceClass: spec.InstanceClass,
	}, nil
}

// ListEngineVersions returns the available MySQL engine versions from the image product list.
// Each unique EngineVersionCode (e.g. "8.0", "5.7") is returned once.
func (p *NCPProvider) ListEngineVersions(_ context.Context) ([]models.DBEngineVersion, error) {
	resp, err := p.api.GetCloudMysqlImageProductList(&vmysql.GetCloudMysqlImageProductListRequest{
		RegionCode:     ncloud.String(p.region),
		GenerationCode: ncloud.String("G3"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP Cloud MySQL image products: %w", err)
	}

	seen := map[string]struct{}{}
	var out []models.DBEngineVersion
	for _, product := range resp.ProductList {
		if product.EngineVersionCode == nil || *product.EngineVersionCode == "" {
			continue
		}
		code := *product.EngineVersionCode
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, models.DBEngineVersion{Engine: "mysql", EngineVersion: code})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EngineVersion < out[j].EngineVersion })
	return out, nil
}

func (p *NCPProvider) ListInstanceClasses(_ context.Context, _, engineVersion string) ([]string, error) {
	imgResp, err := p.api.GetCloudMysqlImageProductList(&vmysql.GetCloudMysqlImageProductListRequest{
		RegionCode:     ncloud.String(p.region),
		GenerationCode: ncloud.String("G3"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP Cloud MySQL image products: %w", err)
	}

	var imageProductCode string
	for _, product := range imgResp.ProductList {
		if product.EngineVersionCode != nil && *product.EngineVersionCode == engineVersion && product.ProductCode != nil {
			imageProductCode = *product.ProductCode
			break
		}
	}
	if imageProductCode == "" {
		return nil, fmt.Errorf("no NCP Cloud MySQL image found for engineVersion %q", engineVersion)
	}

	prodResp, err := p.api.GetCloudMysqlProductList(&vmysql.GetCloudMysqlProductListRequest{
		RegionCode:                 ncloud.String(p.region),
		CloudMysqlImageProductCode: ncloud.String(imageProductCode),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP Cloud MySQL server products: %w", err)
	}

	var out []string
	for _, product := range prodResp.ProductList {
		if product.ProductCode != nil && *product.ProductCode != "" {
			out = append(out, *product.ProductCode)
		}
	}
	sort.Strings(out)
	return out, nil
}
