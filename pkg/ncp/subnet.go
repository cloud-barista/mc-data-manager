// Package ncp provides shared helpers for NCP operations used by both the
// rdbinstance and nrdbinstance NCP providers.
package ncp

import (
	"fmt"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	ncpvpc "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	ncpvserver "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
)

const (
	vpcCIDR    = "172.16.0.0/16"
	subnetCIDR = "172.16.0.0/24"
)

// SubnetInfo holds the SubnetNo and VpcNo resolved for instance creation.
type SubnetInfo struct {
	SubnetNo string
	VpcNo    string
}

// ResolveSubnet returns the first available public subnet in the region.
// If none exists, a new VPC and public subnet are created.
func ResolveSubnet(vpcApi *ncpvpc.V2ApiService, vserverApi *ncpvserver.V2ApiService, region string) (SubnetInfo, error) {
	resp, err := vpcApi.GetSubnetList(&ncpvpc.GetSubnetListRequest{
		RegionCode:     ncloud.String(region),
		SubnetTypeCode: ncloud.String("PUBLIC"),
		UsageTypeCode:  ncloud.String("GEN"),
	})
	if err != nil {
		return SubnetInfo{}, fmt.Errorf("failed to list public subnets: %w", err)
	}
	if resp != nil && len(resp.SubnetList) > 0 {
		first := resp.SubnetList[0]
		return SubnetInfo{
			SubnetNo: ncloud.StringValue(first.SubnetNo),
			VpcNo:    ncloud.StringValue(first.VpcNo),
		}, nil
	}
	return createVPCAndSubnet(vpcApi, vserverApi, region)
}

// waitVPCAvailable polls GetVpcList until VpcStatus.Code is "RUN", timing out after 60s.
func waitVPCAvailable(vpcApi *ncpvpc.V2ApiService, region, vpcNo string) error {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := vpcApi.GetVpcList(&ncpvpc.GetVpcListRequest{
			RegionCode: ncloud.String(region),
			VpcNoList:  []*string{ncloud.String(vpcNo)},
		})
		if err != nil {
			return fmt.Errorf("failed to describe VPC status: %w", err)
		}
		if resp != nil && len(resp.VpcList) > 0 && resp.VpcList[0].VpcStatus != nil {
			if ncloud.StringValue(resp.VpcList[0].VpcStatus.Code) == "RUN" {
				return nil
			}
		}
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("VPC %s did not reach RUN status within 60s", vpcNo)
}

// createVPCAndSubnet creates a new VPC and public subnet, returning the IDs.
func createVPCAndSubnet(vpcApi *ncpvpc.V2ApiService, vserverApi *ncpvserver.V2ApiService, region string) (SubnetInfo, error) {
	vpcResp, err := vpcApi.CreateVpc(&ncpvpc.CreateVpcRequest{
		RegionCode:    ncloud.String(region),
		Ipv4CidrBlock: ncloud.String(vpcCIDR),
	})
	if err != nil {
		return SubnetInfo{}, fmt.Errorf("failed to create NCP VPC: %w", err)
	}
	if vpcResp == nil || len(vpcResp.VpcList) == 0 || vpcResp.VpcList[0].VpcNo == nil {
		return SubnetInfo{}, fmt.Errorf("create NCP VPC returned no VPC id")
	}
	vpcNo := ncloud.StringValue(vpcResp.VpcList[0].VpcNo)

	if err := waitVPCAvailable(vpcApi, region, vpcNo); err != nil {
		return SubnetInfo{}, err
	}

	zoneResp, err := vserverApi.GetZoneList(&ncpvserver.GetZoneListRequest{
		RegionCode: ncloud.String(region),
	})
	if err != nil {
		return SubnetInfo{}, fmt.Errorf("failed to list NCP zones: %w", err)
	}
	if zoneResp == nil || len(zoneResp.ZoneList) == 0 {
		return SubnetInfo{}, fmt.Errorf("no zones found in region %s", region)
	}
	zoneCode := ncloud.StringValue(zoneResp.ZoneList[0].ZoneCode)

	aclResp, err := vpcApi.GetNetworkAclList(&ncpvpc.GetNetworkAclListRequest{
		RegionCode: ncloud.String(region),
		VpcNo:      ncloud.String(vpcNo),
	})
	if err != nil {
		return SubnetInfo{}, fmt.Errorf("failed to list NCP network ACLs: %w", err)
	}
	if aclResp == nil || len(aclResp.NetworkAclList) == 0 {
		return SubnetInfo{}, fmt.Errorf("no network ACL found for VPC %s", vpcNo)
	}
	networkAclNo := ncloud.StringValue(aclResp.NetworkAclList[0].NetworkAclNo)

	subnetResp, err := vpcApi.CreateSubnet(&ncpvpc.CreateSubnetRequest{
		RegionCode:     ncloud.String(region),
		VpcNo:          ncloud.String(vpcNo),
		ZoneCode:       ncloud.String(zoneCode),
		NetworkAclNo:   ncloud.String(networkAclNo),
		Subnet:         ncloud.String(subnetCIDR),
		SubnetTypeCode: ncloud.String("PUBLIC"),
	})
	if err != nil {
		return SubnetInfo{}, fmt.Errorf("failed to create NCP subnet: %w", err)
	}
	if subnetResp == nil || len(subnetResp.SubnetList) == 0 || subnetResp.SubnetList[0].SubnetNo == nil {
		return SubnetInfo{}, fmt.Errorf("create NCP subnet returned no subnet id")
	}
	return SubnetInfo{
		SubnetNo: ncloud.StringValue(subnetResp.SubnetList[0].SubnetNo),
		VpcNo:    vpcNo,
	}, nil
}
