package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/openbao"
	"github.com/rs/zerolog/log"
)

// VaultAddr and VaultToken are package-level vars initialized by InitOpenBao.
var VaultAddr string
var VaultToken string

// InitOpenBao loads VAULT_ADDR and VAULT_TOKEN from environment variables.
// Must be called during application startup (config.Init).
func InitOpenBao() {
	VaultAddr = os.Getenv("VAULT_ADDR")
	if VaultAddr == "" {
		VaultAddr = "http://localhost:8200"
	}
	VaultToken = os.Getenv("VAULT_TOKEN")

	log.Debug().Str("vaultAddr", VaultAddr).Bool("vaultTokenSet", VaultToken != "").Msg("[OpenBao] initialized")
}

// LoadCredentialsByProvider reads CSP credentials from OpenBao by provider name.
// Path convention matches cb-tumblebug: secret/data/csp/{provider}
func (cred *CredentialManager) LoadCredentialsByProvider(ctx context.Context, provider string) (interface{}, error) {
	path := fmt.Sprintf("secret/data/csp/%s", strings.ToLower(provider))
	data, err := openbao.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(provider) {
	case "aws":
		return models.AWSCredentials{
			AccessKey: openbao.GetString(data, "AWS_ACCESS_KEY_ID"),
			SecretKey: openbao.GetString(data, "AWS_SECRET_ACCESS_KEY"),
		}, nil
	case "ncp":
		return models.NCPCredentials{
			AccessKey: openbao.GetString(data, "NCLOUD_ACCESS_KEY"),
			SecretKey: openbao.GetString(data, "NCLOUD_SECRET_KEY"),
		}, nil
	case "gcp":
		// OpenBao stores the private key with literal "\n" — convert to actual newlines.
		privateKey := strings.ReplaceAll(openbao.GetString(data, "private_key"), `\n`, "\n")
		return models.GCPCredentials{
			Type:         "service_account",
			ProjectID:    openbao.GetString(data, "project_id"),
			ClientEmail:  openbao.GetString(data, "client_email"),
			PrivateKey:   privateKey,
			PrivateKeyID: openbao.GetString(data, "private_key_id"),
			ClientID:     openbao.GetString(data, "client_id"),
		}, nil
	case "alibaba":
		return models.AlibabaCredentials{
			AccessKey: openbao.GetString(data, "ALIBABA_CLOUD_ACCESS_KEY_ID"),
			SecretKey: openbao.GetString(data, "ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
		}, nil
	case "ibm":
		return models.IBMCredentials{
			ApiKey:      openbao.GetString(data, "IC_API_KEY"),
			S3AccessKey: openbao.GetString(data, "S3_ACCESS_KEY"),
			S3SecretKey: openbao.GetString(data, "S3_SECRET_KEY"),
		}, nil
	case "kt":
		return models.KTCredentials{
			Username:    openbao.GetString(data, "KT_USERNAME"),
			Password:    openbao.GetString(data, "KT_PASSWORD"),
			DomainName:  openbao.GetString(data, "KT_DOMAIN_NAME"),
			ProjectID:   openbao.GetString(data, "KT_PROJECT_ID"),
			S3AccessKey: openbao.GetString(data, "KT_S3_ACCESS_KEY"),
			S3SecretKey: openbao.GetString(data, "KT_S3_SECRET_KEY"),
		}, nil
	case "tencent":
		return models.TencentCredentials{
			SecretId:  openbao.GetString(data, "TENCENTCLOUD_SECRET_ID"),
			SecretKey: openbao.GetString(data, "TENCENTCLOUD_SECRET_KEY"),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
