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
	ncpPollInterval = 5 * time.Second
	ncpPollTimeout  = 20 * time.Minute
	ncpStatusRunning = "running"
)

// provisionInBackground waits for the instance to reach running state, then
// adds an ACG inbound rule for the MySQL port and creates a public domain.
func (p *NCPProvider) provisionInBackground(instanceNo, vpcNo string) {
	logger := log.With().Str("provider", "ncp").Str("instanceNo", instanceNo).Logger()

	if !p.waitForRunning(instanceNo) {
		return
	}

	inst, err := p.instanceDetail(instanceNo)
	if err != nil {
		logger.Error().Err(err).Msg("ncp failed to get instance detail")
		return
	}

	if len(inst.AccessControlGroupNoList) > 0 {
		acgNo := ncloud.StringValue(inst.AccessControlGroupNoList[0])
		port := "3306"
		if inst.CloudMysqlPort != nil {
			port = strconv.Itoa(int(*inst.CloudMysqlPort))
		}
		if err := p.addACGInboundRule(acgNo, vpcNo, port); err != nil {
			logger.Error().Err(err).Str("acgNo", acgNo).Msg("ncp addACGInboundRule failed")
		} else {
			logger.Info().Str("acgNo", acgNo).Str("port", port).Msg("ncp ACG inbound rule added")
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

// waitForRunning polls CloudMysqlInstanceStatusName until "running" or timeout.
func (p *NCPProvider) waitForRunning(instanceNo string) bool {
	deadline := time.Now().Add(ncpPollTimeout)
	for {
		time.Sleep(ncpPollInterval)

		status, err := p.instanceStatus(instanceNo)
		if err != nil {
			log.Warn().Err(err).Str("instanceNo", instanceNo).Msg("ncp status poll failed")
		} else if status == ncpStatusRunning {
			log.Info().Str("instanceNo", instanceNo).Msg("ncp instance is running")
			return true
		} else {
			log.Info().Str("instanceNo", instanceNo).Str("status", status).Msg("ncp waiting for running")
		}

		if time.Now().After(deadline) {
			log.Error().Str("instanceNo", instanceNo).Msg("ncp timed out waiting for running")
			return false
		}
	}
}

// instanceStatus returns CloudMysqlInstanceStatusName for the given instance.
func (p *NCPProvider) instanceStatus(instanceNo string) (string, error) {
	inst, err := p.instanceDetail(instanceNo)
	if err != nil {
		return "", err
	}
	if inst.CloudMysqlInstanceStatusName == nil {
		return "", nil
	}
	return *inst.CloudMysqlInstanceStatusName, nil
}