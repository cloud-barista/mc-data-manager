package alibaba

import (
	"strings"
	"time"

	alibabards "github.com/alibabacloud-go/rds-20140815/v8/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/repository"
	"github.com/rs/zerolog/log"
)

const (
	alibabaPollInterval = 5 * time.Second
	alibabaPollTimeout = 15 * time.Minute
	alibabaPublicPort = "3306"
	statusRunning     = "Running"
)

func (p *AlibabaProvider) provisionInBackground(instanceID, masterUsername, masterPassword string) {
	logger := log.With().Str("provider", "alibaba").Str("instanceId", instanceID).Logger()

	if !p.waitForRunning(instanceID) {
		return
	}

	// 퍼블릭 도메인 생성
	if err := p.allocatePublicConnection(instanceID); err != nil {
		logger.Error().Err(err).Msgf("alibaba public endpoint allocation failed: %v", err)
	} else if p.waitForPublicEndpoint(instanceID) {
		logger.Info().Msg("alibaba public endpoint allocated")
	} else {
		logger.Error().Msg("alibaba public endpoint never became ready")
	}

	if err := repository.NewRDBInstanceRepository(config.DB).
		UpdatePublicNetPending("alibaba", p.region, instanceID, false); err != nil {
		logger.Error().Err(err).Msg("failed to clear public_net_pending")
	}

	if !p.waitForRunning(instanceID) {
		return
	}

	// 계정 생성
	if err := p.CreateAccount(instanceID, masterUsername, masterPassword); err != nil {
		logger.Error().Err(err).Msg("alibaba CreateAccount failed")
		if err := repository.NewRDBInstanceRepository(config.DB).
			UpdateAccountCreateFailed("alibaba", p.region, instanceID, true); err != nil {
			logger.Error().Err(err).Msg("failed to flag account_create_failed")
		}
		return
	}
	logger.Info().Msg("alibaba account created; provisioning complete")
}

// waitForRunning polls the instance status until Running or the timeout elapses.
func (p *AlibabaProvider) waitForRunning(instanceID string) bool {
	deadline := time.Now().Add(alibabaPollTimeout)
	for {
		time.Sleep(alibabaPollInterval)

		status, err := p.instanceStatus(instanceID)
		if err != nil {
			log.Warn().Err(err).Str("instanceId", instanceID).Msg("alibaba status poll failed")
		} else if status == statusRunning {
			log.Info().Str("instanceId", instanceID).Msg("alibaba instance is Running")
			return true
		} else {
			log.Info().Str("instanceId", instanceID).Str("status", status).Msg("alibaba waiting for Running")
		}

		if time.Now().After(deadline) {
			log.Error().Str("instanceId", instanceID).Msg("alibaba timed out waiting for Running")
			return false
		}
	}
}

// waitForPublicEndpoint polls the instance's network info until the public
// endpoint's connection string is populated, or the timeout elapses.
func (p *AlibabaProvider) waitForPublicEndpoint(instanceID string) bool {
	deadline := time.Now().Add(alibabaPollTimeout)
	for {
		time.Sleep(alibabaPollInterval)

		endpoint, _, err := p.instanceEndpoint(instanceID)
		if err != nil {
			log.Warn().Err(err).Str("instanceId", instanceID).Msg("alibaba public endpoint poll failed")
		} else if endpoint != "-" {
			log.Info().Str("instanceId", instanceID).Str("endpoint", endpoint).Msg("alibaba public endpoint ready")
			return true
		} else {
			log.Info().Str("instanceId", instanceID).Msg("alibaba waiting for public endpoint")
		}

		if time.Now().After(deadline) {
			log.Error().Str("instanceId", instanceID).Msg("alibaba timed out waiting for public endpoint")
			return false
		}
	}
}

// instanceStatus returns the current DBInstanceStatus of the instance.
// Retries up to 3 times on EOF (server-side idle connection closure).
func (p *AlibabaProvider) instanceStatus(instanceID string) (string, error) {
	var lastErr error
	for range 3 {
		resp, err := p.client.DescribeDBInstanceAttribute(&alibabards.DescribeDBInstanceAttributeRequest{
			DBInstanceId: tea.String(instanceID),
		})
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				lastErr = err
				time.Sleep(2 * time.Second)
				continue
			}
			return "", err
		}
		if resp == nil || resp.Body == nil || resp.Body.Items == nil || len(resp.Body.Items.DBInstanceAttribute) == 0 {
			return "", nil
		}
		return tea.StringValue(resp.Body.Items.DBInstanceAttribute[0].DBInstanceStatus), nil
	}
	return "", lastErr
}

// allocatePublicConnection allocates a public endpoint for the instance.
func (p *AlibabaProvider) allocatePublicConnection(instanceID string) error {
	_, err := p.client.AllocateInstancePublicConnection(&alibabards.AllocateInstancePublicConnectionRequest{
		DBInstanceId:           tea.String(instanceID),
		ConnectionStringPrefix: tea.String(strings.ToLower(instanceID) + "-pub"),
		Port:                   tea.String(alibabaPublicPort),
	})
	return err
}
