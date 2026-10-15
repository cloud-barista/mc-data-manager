package alibaba

import (
	"strings"
	"time"

	dds "github.com/alibabacloud-go/dds-20151201/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/rs/zerolog/log"
)

const (
	alibabaPollInterval = 30 * time.Second
	alibabaPollTimeout  = 20 * time.Minute
	statusRunning       = "Running"
)

// provisionInBackground waits for the instance to become Running, then allocates
// a public network address. Best-effort: progress and failures are logged only.
func (p *AlibabaProvider) provisionInBackground(instanceID string) {
	logger := log.With().Str("provider", "alibaba-mongodb").Str("instanceId", instanceID).Logger()

	deadline := time.Now().Add(alibabaPollTimeout)
	for {
		time.Sleep(alibabaPollInterval)

		status, err := p.instanceStatus(instanceID)
		if err != nil {
			logger.Warn().Err(err).Msg("status poll failed")
		} else if status == statusRunning {
			logger.Info().Msg("instance is Running")
			break
		} else {
			logger.Info().Str("status", status).Msg("waiting for Running")
		}

		if time.Now().After(deadline) {
			logger.Error().Msg("timed out waiting for Running")
			return
		}
	}

	if err := p.allocatePublicNetworkAddress(instanceID); err != nil {
		logger.Error().Err(err).Msg("AllocatePublicNetworkAddress failed")
		return
	}
	logger.Info().Msg("public network address allocated; provisioning complete")
}

// instanceStatus returns the current DBInstanceStatus of the instance.
// Retries up to 3 times on EOF (server-side idle connection closure).
func (p *AlibabaProvider) instanceStatus(instanceID string) (string, error) {
	var lastErr error
	for range 3 {
		resp, err := p.client.DescribeDBInstanceAttribute(&dds.DescribeDBInstanceAttributeRequest{
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
		if resp == nil || resp.Body == nil || resp.Body.DBInstances == nil || len(resp.Body.DBInstances.DBInstance) == 0 {
			return "", nil
		}
		return tea.StringValue(resp.Body.DBInstances.DBInstance[0].DBInstanceStatus), nil
	}
	return "", lastErr
}

// allocatePublicNetworkAddress allocates a public endpoint for the replica-set instance.
func (p *AlibabaProvider) allocatePublicNetworkAddress(instanceID string) error {
	_, err := p.client.AllocatePublicNetworkAddress(&dds.AllocatePublicNetworkAddressRequest{
		DBInstanceId: tea.String(instanceID),
	})
	return err
}
