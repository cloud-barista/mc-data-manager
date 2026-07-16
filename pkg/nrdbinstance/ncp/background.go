package ncp

import (
	"fmt"
	"strconv"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	ncpvserver "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/rs/zerolog/log"
)

const (
	ncpMongoPollInterval  = 30 * time.Second
	ncpMongoPollTimeout   = 20 * time.Minute
	ncpMongoStatusRunning = "running"
)

// provisionInBackground waits for the instance to reach running state, then
// adds an ACG inbound rule for the MongoDB member port.
func (p *NCPProvider) provisionInBackground(instanceNo, vpcNo string) {
	logger := log.With().Str("provider", "ncp-mongodb").Str("instanceNo", instanceNo).Logger()

	if !p.waitForRunning(instanceNo) {
		return
	}

	inst, err := p.instanceDetail(instanceNo)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get instance detail")
		return
	}

	if len(inst.AccessControlGroupNoList) > 0 {
		acgNo := ncloud.StringValue(inst.AccessControlGroupNoList[0])
		port := "27017"
		if inst.MemberPort != nil {
			port = strconv.Itoa(int(*inst.MemberPort))
		}
		if err := p.addACGInboundRule(acgNo, vpcNo, port); err != nil {
			logger.Error().Err(err).Str("acgNo", acgNo).Msg("addACGInboundRule failed")
		} else {
			logger.Info().Str("acgNo", acgNo).Str("port", port).Msg("ACG inbound rule added; provisioning complete")
		}
	}
}

func (p *NCPProvider) waitForRunning(instanceNo string) bool {
	deadline := time.Now().Add(ncpMongoPollTimeout)
	for {
		time.Sleep(ncpMongoPollInterval)

		inst, err := p.instanceDetail(instanceNo)
		if err != nil {
			log.Warn().Err(err).Str("instanceNo", instanceNo).Msg("ncp-mongodb status poll failed")
		} else {
			status := ncloud.StringValue(inst.CloudMongoDbInstanceStatusName)
			if status == ncpMongoStatusRunning {
				log.Info().Str("instanceNo", instanceNo).Msg("ncp-mongodb instance is running")
				return true
			}
			log.Info().Str("instanceNo", instanceNo).Str("status", status).Msg("ncp-mongodb waiting for running")
		}

		if time.Now().After(deadline) {
			log.Error().Str("instanceNo", instanceNo).Msg("ncp-mongodb timed out waiting for running")
			return false
		}
	}
}

// addACGInboundRule adds a TCP inbound rule for the given port to the ACG.
func (p *NCPProvider) addACGInboundRule(acgNo, vpcNo, port string) error {
	_, err := p.vserverApi.AddAccessControlGroupInboundRule(&ncpvserver.AddAccessControlGroupInboundRuleRequest{
		RegionCode:           ncloud.String(p.region),
		AccessControlGroupNo: ncloud.String(acgNo),
		VpcNo:                ncloud.String(vpcNo),
		AccessControlGroupRuleList: []*ncpvserver.AddAccessControlGroupRuleParameter{
			{
				ProtocolTypeCode: ncloud.String("TCP"),
				IpBlock:          ncloud.String(config.OutboundIP),
				PortRange:        ncloud.String(port),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add ACG inbound rule: %w", err)
	}
	return nil
}
