// Package ncp implements nrdbinstance.Provider for NCP Cloud DB for MongoDB.
package ncp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	ncpvpc "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	ncpvserver "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	vmongodb "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vmongodb"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	ncpcommon "github.com/cloud-barista/mc-data-manager/pkg/ncp"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance"
)

// NCPProvider implements nrdbinstance.Provider for NCP Cloud DB for MongoDB.
type NCPProvider struct {
	api        *vmongodb.V2ApiService
	vpcApi     *ncpvpc.V2ApiService
	vserverApi *ncpvserver.V2ApiService
	region     string
}

// New builds an NCP MongoDB provider from static credentials and a region.
func New(accessKey, secretKey, region string) (nrdbinstance.Provider, error) {
	api, err := config.NewNCPCloudMongoDbClient(accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create NCP Cloud MongoDB client: %w", err)
	}
	vpcApi, err := config.NewNCPVPCClient(accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create NCP VPC client: %w", err)
	}
	vserverApi, err := config.NewNCPVServerClient(accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create NCP VServer client: %w", err)
	}
	return &NCPProvider{api: api, vpcApi: vpcApi, vserverApi: vserverApi, region: region}, nil
}

func (p *NCPProvider) ListInstances(_ context.Context) ([]models.NRDBInstance, error) {
	resp, err := p.api.GetCloudMongoDbInstanceList(&vmongodb.GetCloudMongoDbInstanceListRequest{
		RegionCode: ncloud.String(p.region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP MongoDB instances: %w", err)
	}

	out := []models.NRDBInstance{}
	for _, inst := range resp.CloudMongoDbInstanceList {
		instanceNo := ncloud.StringValue(inst.CloudMongoDbInstanceNo)
		detail, err := p.instanceDetail(instanceNo)
		if err != nil {
			return nil, err
		}
		out = append(out, toNRDBInstance(detail, p.region))
	}
	return out, nil
}

func (p *NCPProvider) instanceDetail(instanceNo string) (*vmongodb.CloudMongoDbInstance, error) {
	resp, err := p.api.GetCloudMongoDbInstanceDetail(&vmongodb.GetCloudMongoDbInstanceDetailRequest{
		RegionCode:             ncloud.String(p.region),
		CloudMongoDbInstanceNo: ncloud.String(instanceNo),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe NCP MongoDB instance %s: %w", instanceNo, err)
	}
	if resp == nil || len(resp.CloudMongoDbInstanceList) == 0 {
		return nil, fmt.Errorf("no detail returned for NCP MongoDB instance %s", instanceNo)
	}
	return resp.CloudMongoDbInstanceList[0], nil
}

func normalizeStatus(s string) string {
	if s == "running" {
		return "available"
	}
	return s
}

func toNRDBInstance(inst *vmongodb.CloudMongoDbInstance, region string) models.NRDBInstance {
	instanceNo := ncloud.StringValue(inst.CloudMongoDbInstanceNo)
	name := instanceNo
	if inst.CloudMongoDbServiceName != nil && *inst.CloudMongoDbServiceName != "" {
		name = *inst.CloudMongoDbServiceName
	}
	status := ""
	if inst.CloudMongoDbInstanceStatusName != nil {
		status = normalizeStatus(*inst.CloudMongoDbInstanceStatusName)
	}
	var port int32
	if inst.MemberPort != nil {
		port = *inst.MemberPort
	}

	endpoint := "-"
	for _, srv := range inst.CloudMongoDbServerInstanceList {
		if srv.PublicDomain != nil && *srv.PublicDomain != "" {
			endpoint = *srv.PublicDomain
			break
		}
	}

	return models.NRDBInstance{
		Provider:      "ncp",
		InstanceID:    instanceNo,
		Name:          name,
		Status:        status,
		Region:        region,
		EngineVersion: ncloud.StringValue(inst.EngineVersion),
		Endpoint:      endpoint,
		Port:          port,
	}
}

func (p *NCPProvider) CreateInstance(_ context.Context, spec nrdbinstance.CreateSpec) (models.NRDBInstance, error) {
	sub, err := ncpcommon.ResolveSubnet(p.vpcApi, p.vserverApi, p.region)
	if err != nil {
		return models.NRDBInstance{}, err
	}

	resp, err := p.api.CreateCloudMongoDbInstance(&vmongodb.CreateCloudMongoDbInstanceRequest{
		RegionCode:                  ncloud.String(p.region),
		VpcNo:                       ncloud.String(sub.VpcNo),
		SubnetNo:                    ncloud.String(sub.SubnetNo),
		CloudMongoDbServiceName:     ncloud.String(spec.InstanceID),
		CloudMongoDbServerNamePrefix: ncloud.String(spec.InstanceID),
		CloudMongoDbUserName:        ncloud.String(spec.MasterUsername),
		CloudMongoDbUserPassword:    ncloud.String(spec.MasterPassword),
		EngineVersionCode:           ncloud.String(spec.EngineVersion),
		MemberProductCode:           ncloud.String(spec.InstanceClass),
		ClusterTypeCode:             ncloud.String("STAND_ALONE"),
	})
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("failed to create NCP MongoDB instance: %w", err)
	}

	instanceNo := ""
	status := "creating"
	if resp != nil && len(resp.CloudMongoDbInstanceList) > 0 {
		inst := resp.CloudMongoDbInstanceList[0]
		if inst.CloudMongoDbInstanceNo != nil {
			instanceNo = *inst.CloudMongoDbInstanceNo
		}
		if inst.CloudMongoDbInstanceStatusName != nil {
			status = normalizeStatus(*inst.CloudMongoDbInstanceStatusName)
		}
	}

	go p.provisionInBackground(instanceNo, sub.VpcNo)

	return models.NRDBInstance{
		Provider:      "ncp",
		InstanceID:    instanceNo,
		Name:          spec.InstanceID,
		Status:        status,
		Region:        p.region,
		EngineVersion: spec.EngineVersion,
	}, nil
}

func (p *NCPProvider) DeleteInstance(_ context.Context, instanceID string) (models.NRDBInstance, error) {
	resp, err := p.api.DeleteCloudMongoDbInstance(&vmongodb.DeleteCloudMongoDbInstanceRequest{
		RegionCode:              ncloud.String(p.region),
		CloudMongoDbInstanceNo: ncloud.String(instanceID),
	})
	if err != nil {
		if strings.Contains(err.Error(), "5001022") {
			return models.NRDBInstance{Provider: "ncp", InstanceID: instanceID, Status: "deleted", Region: p.region}, nil
		}
		return models.NRDBInstance{}, fmt.Errorf("failed to delete NCP MongoDB instance: %w", err)
	}

	name := instanceID
	if resp != nil && len(resp.CloudMongoDbInstanceList) > 0 {
		if sn := resp.CloudMongoDbInstanceList[0].CloudMongoDbServiceName; sn != nil {
			name = *sn
		}
	}

	return models.NRDBInstance{
		Provider:   "ncp",
		InstanceID: instanceID,
		Name:       name,
		Status:     "deleting",
		Region:     p.region,
	}, nil
}

// ListEngineVersions returns available MongoDB engine versions from the image product list.
func (p *NCPProvider) ListEngineVersions(_ context.Context) ([]models.NRDBEngineVersion, error) {
	resp, err := p.api.GetCloudMongoDbImageProductList(&vmongodb.GetCloudMongoDbImageProductListRequest{
		RegionCode:     ncloud.String(p.region),
		GenerationCode: ncloud.String("G3"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP MongoDB image products: %w", err)
	}

	seen := map[string]struct{}{}
	var out []models.NRDBEngineVersion
	for _, p := range resp.ProductList {
		if p.EngineVersionCode == nil || *p.EngineVersionCode == "" {
			continue
		}
		code := *p.EngineVersionCode
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, models.NRDBEngineVersion{EngineVersion: code})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EngineVersion < out[j].EngineVersion })
	return out, nil
}

func (p *NCPProvider) ListInstanceClasses(_ context.Context, engineVersion string) ([]string, error) {
	imgResp, err := p.api.GetCloudMongoDbImageProductList(&vmongodb.GetCloudMongoDbImageProductListRequest{
		RegionCode:     ncloud.String(p.region),
		GenerationCode: ncloud.String("G3"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP MongoDB image products: %w", err)
	}

	var imageProductCode string
	for _, prod := range imgResp.ProductList {
		if prod.EngineVersionCode != nil && *prod.EngineVersionCode == engineVersion && prod.ProductCode != nil {
			imageProductCode = *prod.ProductCode
			break
		}
	}
	if imageProductCode == "" {
		return nil, fmt.Errorf("no NCP MongoDB image found for engineVersion %q", engineVersion)
	}

	prodResp, err := p.api.GetCloudMongoDbProductList(&vmongodb.GetCloudMongoDbProductListRequest{
		RegionCode:                   ncloud.String(p.region),
		CloudMongoDbImageProductCode: ncloud.String(imageProductCode),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list NCP MongoDB server products: %w", err)
	}

	seen := map[string]struct{}{}
	var out []string
	for _, prod := range prodResp.ProductList {
		if prod.ProductCode == nil || *prod.ProductCode == "" {
			continue
		}
		if prod.InfraResourceDetailType != nil {
			code := ncloud.StringValue(prod.InfraResourceDetailType.Code)
			if code != "MNGOD" {
				continue
			}
		}
		pc := *prod.ProductCode
		if _, ok := seen[pc]; ok {
			continue
		}
		seen[pc] = struct{}{}
		out = append(out, pc)
	}
	sort.Strings(out)
	return out, nil
}
