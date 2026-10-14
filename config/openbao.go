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

// spiderKeyMap maps env-var format keys (stored by Tumblebug's credentialKeyMap) to
// Spider-format keys (stored by admin-cli, which reads field names from Spider meta API).
// Used as a fallback when the env-var key is absent — e.g. credentials registered via
// admin-cli bypass Tumblebug's credentialKeyMap conversion and land in OpenBao as-is.
var spiderKeyMap = map[string]map[string]string{
	"aws": {
		"AWS_ACCESS_KEY_ID":     "ClientId",
		"AWS_SECRET_ACCESS_KEY": "ClientSecret",
	},
	"ncp": {
		"NCLOUD_ACCESS_KEY": "ClientId",
		"NCLOUD_SECRET_KEY": "ClientSecret",
	},
	"nhn": { // NHN has no Spider format, so map keys to themselves
		"NHN_USERNAME":          "NHN_USERNAME",
		"NHN_PASSWORD":          "NHN_PASSWORD",
		"NHN_TENANT_ID":         "NHN_TENANT_ID",
		"NHN_DOMAIN_NAME":       "NHN_DOMAIN_NAME",
		"NHN_IDENTITY_ENDPOINT": "NHN_IDENTITY_ENDPOINT",
	},
	"gcp": {
		"private_key":  "PrivateKey",
		"project_id":   "ProjectID",
		"client_email": "ClientEmail",
	},
	"alibaba": {
		"ALIBABA_CLOUD_ACCESS_KEY_ID":     "ClientId",
		"ALIBABA_CLOUD_ACCESS_KEY_SECRET": "ClientSecret",
	},
	"ibm": { //IBM Credentials does not have a Spider format, but we include it here for completeness
		"IC_API_KEY":        "IC_API_KEY",
		"IBM_S3_ACCESS_KEY": "IBM_S3_ACCESS_KEY",
		"IBM_S3_SECRET_KEY": "IBM_S3_SECRET_KEY",
	},
	"kt": {
		"KT_USERNAME":      "Username",
		"KT_PASSWORD":      "Password",
		"KT_DOMAIN_NAME":   "DomainName",
		"KT_PROJECT_ID":    "ProjectID",
		"KT_S3_ACCESS_KEY": "S3AccessKey",
		"KT_S3_SECRET_KEY": "S3SecretKey",
	},
	"tencent": {
		"TENCENTCLOUD_SECRET_ID":  "TENCENTCLOUD_SECRET_ID",
		"TENCENTCLOUD_SECRET_KEY": "TENCENTCLOUD_SECRET_KEY",
	},
}

// getString reads key from data, falling back to the Spider-format key for the given provider
// when the primary key is absent.
func getString(data map[string]interface{}, provider, key string) string {
	if v := openbao.GetString(data, key); v != "" {
		return v
	}
	if m, ok := spiderKeyMap[provider]; ok {
		if fallback, ok := m[key]; ok {
			return openbao.GetString(data, fallback)
		}
	}
	return ""
}

// LoadCredentialsByProvider reads CSP credentials from OpenBao by provider name.
// Path convention matches cb-tumblebug: secret/data/csp/{provider}
func (cred *CredentialManager) LoadCredentialsByProvider(ctx context.Context, provider string) (interface{}, error) {
	path := fmt.Sprintf("secret/data/csp/%s", strings.ToLower(provider))
	data, err := openbao.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	p := strings.ToLower(provider)
	switch p {
	case "aws":
		return models.AWSCredentials{
			AccessKey: getString(data, p, "AWS_ACCESS_KEY_ID"),
			SecretKey: getString(data, p, "AWS_SECRET_ACCESS_KEY"),
		}, nil
	case "ncp":
		return models.NCPCredentials{
			AccessKey: getString(data, p, "NCLOUD_ACCESS_KEY"),
			SecretKey: getString(data, p, "NCLOUD_SECRET_KEY"),
		}, nil
	case "nhn":
		return models.NHNCredentials{
			Username:         getString(data, p, "NHN_USERNAME"),
			Password:         getString(data, p, "NHN_PASSWORD"),
			TenantID:         getString(data, p, "NHN_TENANT_ID"),
			DomainName:       getString(data, p, "NHN_DOMAIN_NAME"),
			IdentityEndpoint: getString(data, p, "NHN_IDENTITY_ENDPOINT"),
			S3AccessKey:      getString(data, p, "NHN_S3_ACCESS_KEY"),
			S3SecretKey:      getString(data, p, "NHN_S3_SECRET_KEY"),
		}, nil
	case "gcp":
		// OpenBao stores the private key with literal "\n" — convert to actual newlines.
		privateKey := strings.ReplaceAll(getString(data, p, "private_key"), `\n`, "\n")
		return models.GCPCredentials{
			Type:         "service_account",
			ProjectID:    getString(data, p, "project_id"),
			ClientEmail:  getString(data, p, "client_email"),
			PrivateKey:   privateKey,
			PrivateKeyID: getString(data, p, "private_key_id"),
			ClientID:     getString(data, p, "client_id"),
			S3AccessKey:  getString(data, p, "GCP_S3_ACCESS_KEY"),
			S3SecretKey:  getString(data, p, "GCP_S3_SECRET_KEY"),
		}, nil
	case "alibaba":
		return models.AlibabaCredentials{
			AccessKey: getString(data, p, "ALIBABA_CLOUD_ACCESS_KEY_ID"),
			SecretKey: getString(data, p, "ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
		}, nil
	case "ibm":
		return models.IBMCredentials{
			ApiKey:      getString(data, p, "IC_API_KEY"),
			S3AccessKey: getString(data, p, "IBM_S3_ACCESS_KEY"),
			S3SecretKey: getString(data, p, "IBM_S3_SECRET_KEY"),
		}, nil
	case "kt":
		return models.KTCredentials{
			Username:    getString(data, p, "KT_USERNAME"),
			Password:    getString(data, p, "KT_PASSWORD"),
			DomainName:  getString(data, p, "KT_DOMAIN_NAME"),
			ProjectID:   getString(data, p, "KT_PROJECT_ID"),
			S3AccessKey: getString(data, p, "KT_S3_ACCESS_KEY"),
			S3SecretKey: getString(data, p, "KT_S3_SECRET_KEY"),
		}, nil
	case "tencent":
		return models.TencentCredentials{
			SecretId:  getString(data, p, "TENCENTCLOUD_SECRET_ID"),
			SecretKey: getString(data, p, "TENCENTCLOUD_SECRET_KEY"),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
