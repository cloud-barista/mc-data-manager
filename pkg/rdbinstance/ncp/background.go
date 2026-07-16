package ncp

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/hmac"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	ncpvserver "github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/rs/zerolog/log"
)

const (
	ncpPollInterval = 5 * time.Second
	ncpPollTimeout  = 10 * time.Minute
	ncpBasePath     = "https://vmysql.apigw.ntruss.com/api/v2"
	// ncpBasePath = "https://console.ncloud.com/vpcCloudMysql/api/v1"
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

// serverInstanceNo returns the CloudMysqlServerInstanceNo of the primary server.
func (p *NCPProvider) serverInstanceNo(instanceNo string) (string, error) {
	inst, err := p.instanceDetail(instanceNo)
	if err != nil {
		return "", err
	}
	if len(inst.CloudMysqlServerInstanceList) == 0 {
		return "", fmt.Errorf("no server instances found for %s", instanceNo)
	}
	no := ncloud.StringValue(inst.CloudMysqlServerInstanceList[0].CloudMysqlServerInstanceNo)
	if no == "" {
		return "", fmt.Errorf("empty server instance no for %s", instanceNo)
	}
	return no, nil
}

func (p *NCPProvider) createPublicDomain(instanceNo, serverInstanceNo string) error {
	path := "/applyPublicDomainRequest"
	body := map[string]interface{}{
		"serviceNo":   instanceNo,
		"computeNo":   serverInstanceNo,
		"isPublicUse": true,
	}
	return p.callRawAPI(http.MethodPost, path, body)
}

func (p *NCPProvider) callRawAPI(method, path string, body interface{}) error {
	fullURL := ncpBasePath + path

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	signer := hmac.NewSigner(p.secretKey, crypto.SHA256)
	signature, err := signer.Sign(method, fullURL, p.accessKey, timestamp)
	if err != nil {
		return fmt.Errorf("hmac sign failed: %w", err)
	}

	req, err := http.NewRequest(method, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ncp-apigw-timestamp", timestamp)
	req.Header.Set("x-ncp-iam-access-key", p.accessKey)
	req.Header.Set("x-ncp-apigw-signature-v1", signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("NCP API returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
