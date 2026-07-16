package alibaba

import (
	"strings"
	"time"

	alibabards "github.com/alibabacloud-go/rds-20140815/v8/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/rs/zerolog/log"
)

const (
	// alibabaPollInterval is how often the instance status is polled while waiting
	// for it to become Running before creating the account / public endpoint.
	alibabaPollInterval = 5 * time.Second
	// alibabaPollTimeout caps the total wait for the instance to become Running.
	alibabaPollTimeout = 10 * time.Minute
	// alibabaPublicPort is the fixed port assigned to the public endpoint (1000-5999).
	alibabaPublicPort = "3306"
	statusRunning     = "Running"
)

// provisionInBackground waits for the instance to become Running, then creates
// the master account and allocates a public endpoint. Best-effort: progress and
// failures are logged; nothing is returned to the caller.
func (p *AlibabaProvider) provisionInBackground(instanceID, masterUsername, masterPassword string) {
	logger := log.With().Str("provider", "alibaba").Str("instanceId", instanceID).Logger()

	if !p.waitForRunning(instanceID) {
		return
	}

	if err := p.createAccount(instanceID, masterUsername, masterPassword); err != nil {
		logger.Error().Err(err).Msg("alibaba CreateAccount failed")
		return
	}
	logger.Info().Msg("alibaba account created")

	if !p.waitForRunning(instanceID) {
		return
	}

	if err := p.allocatePublicConnection(instanceID); err != nil {
		logger.Error().Err(err).Msg("alibaba AllocateInstancePublicConnection failed")
		return
	}
	logger.Info().Msg("alibaba public endpoint allocated; provisioning complete")
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

// createAccount creates the master account on the instance.
func (p *AlibabaProvider) createAccount(instanceID, username, password string) error {
	_, err := p.client.CreateAccount(&alibabards.CreateAccountRequest{
		DBInstanceId:    tea.String(instanceID),
		AccountName:     tea.String(username),
		AccountPassword: tea.String(password),
		AccountType:     tea.String("Super"),
	})
	return err
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
