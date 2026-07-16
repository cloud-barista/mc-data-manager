// Package alibaba provides shared helpers for Alibaba Cloud operations used by
// both the rdbinstance and nrdbinstance Alibaba providers.
package alibaba

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	alibabavpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
)

const (
	vpcCIDR     = "172.16.0.0/16"
	vSwitchCIDR = "172.16.0.0/24"
)

// VSwitchInfo holds the VSwitchId, VpcId, and ZoneId resolved for instance creation.
// ZoneId is required when creating DDS (MongoDB) instances; it must match the VSwitch's zone.
type VSwitchInfo struct {
	VSwitchID string
	VpcID     string
	ZoneID    string
}

// IsUsableZone reports whether zoneID is a single-AZ zone.
// Multi-AZ aggregated zones (containing "MAZ" or "(") are not usable for
// single-resource placement such as VSwitches or instance classes.
func IsUsableZone(zoneID string) bool {
	return !strings.Contains(zoneID, "MAZ") && !strings.Contains(zoneID, "(")
}

// ResolveVSwitch returns the first available VSwitch in the region.
// If none exists, it creates a new VPC and VSwitch and returns those IDs.
func ResolveVSwitch(vpcClient *alibabavpc.Client, region string) (VSwitchInfo, error) {
	resp, err := vpcClient.DescribeVSwitches(&alibabavpc.DescribeVSwitchesRequest{
		RegionId: tea.String(region),
	})
	if err != nil {
		return VSwitchInfo{}, fmt.Errorf("failed to describe VSwitches: %w", err)
	}
	if resp != nil && resp.Body != nil && resp.Body.VSwitches != nil {
		for _, vsw := range resp.Body.VSwitches.VSwitch {
			if tea.Int64Value(vsw.AvailableIpAddressCount) > 0 {
				return VSwitchInfo{
					VSwitchID: tea.StringValue(vsw.VSwitchId),
					VpcID:     tea.StringValue(vsw.VpcId),
					ZoneID:    tea.StringValue(vsw.ZoneId),
				}, nil
			}
		}
	}
	return createVPCAndVSwitch(vpcClient, region)
}

// waitVPCAvailable polls DescribeVpcs until the VPC reaches "Available" status,
// timing out after 60 seconds. VPC creation is asynchronous; creating a VSwitch
// before the VPC is available returns IncorrectVpcStatus.
func waitVPCAvailable(vpcClient *alibabavpc.Client, region, vpcID string) error {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := vpcClient.DescribeVpcs(&alibabavpc.DescribeVpcsRequest{
			RegionId: tea.String(region),
			VpcId:    tea.String(vpcID),
		})
		if err != nil {
			return fmt.Errorf("failed to describe VPC status: %w", err)
		}
		if resp != nil && resp.Body != nil && resp.Body.Vpcs != nil && len(resp.Body.Vpcs.Vpc) > 0 {
			if tea.StringValue(resp.Body.Vpcs.Vpc[0].Status) == "Available" {
				return nil
			}
		}
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("VPC %s did not reach Available status within 60s", vpcID)
}

// createVPCAndVSwitch creates a new VPC and a VSwitch inside it, returning the IDs.
// The ZoneId is resolved via VPC DescribeZones so no engine context is required.
func createVPCAndVSwitch(vpcClient *alibabavpc.Client, region string) (VSwitchInfo, error) {
	vpcResp, err := vpcClient.CreateVpc(&alibabavpc.CreateVpcRequest{
		RegionId:  tea.String(region),
		CidrBlock: tea.String(vpcCIDR),
	})
	if err != nil {
		return VSwitchInfo{}, fmt.Errorf("failed to create VPC: %w", err)
	}
	if vpcResp == nil || vpcResp.Body == nil || tea.StringValue(vpcResp.Body.VpcId) == "" {
		return VSwitchInfo{}, fmt.Errorf("create VPC returned no VPC id")
	}
	vpcID := tea.StringValue(vpcResp.Body.VpcId)

	if err := waitVPCAvailable(vpcClient, region, vpcID); err != nil {
		return VSwitchInfo{}, err
	}

	zonesResp, err := vpcClient.DescribeZones(&alibabavpc.DescribeZonesRequest{
		RegionId: tea.String(region),
	})
	if err != nil {
		return VSwitchInfo{}, fmt.Errorf("failed to describe zones: %w", err)
	}
	if zonesResp == nil || zonesResp.Body == nil || zonesResp.Body.Zones == nil || len(zonesResp.Body.Zones.Zone) == 0 {
		return VSwitchInfo{}, fmt.Errorf("no zones found in region %s", region)
	}
	var zoneID string
	for _, z := range zonesResp.Body.Zones.Zone {
		if id := tea.StringValue(z.ZoneId); IsUsableZone(id) {
			zoneID = id
			break
		}
	}
	if zoneID == "" {
		return VSwitchInfo{}, fmt.Errorf("no usable single-AZ zone found in region %s", region)
	}

	vswResp, err := vpcClient.CreateVSwitch(&alibabavpc.CreateVSwitchRequest{
		VpcId:     tea.String(vpcID),
		ZoneId:    tea.String(zoneID),
		CidrBlock: tea.String(vSwitchCIDR),
	})
	if err != nil {
		return VSwitchInfo{}, fmt.Errorf("failed to create VSwitch: %w", err)
	}
	if vswResp == nil || vswResp.Body == nil || tea.StringValue(vswResp.Body.VSwitchId) == "" {
		return VSwitchInfo{}, fmt.Errorf("create VSwitch returned no VSwitch id")
	}
	return VSwitchInfo{
		VSwitchID: tea.StringValue(vswResp.Body.VSwitchId),
		VpcID:     vpcID,
		ZoneID:    zoneID,
	}, nil
}
