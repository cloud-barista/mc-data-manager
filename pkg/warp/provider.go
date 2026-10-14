package warp

import (
	"fmt"

	"github.com/cloud-barista/mc-data-manager/models"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
)

// providerConfig holds the warp-specific S3 addressing details for a provider.
type providerConfig struct {
	Endpoint  func(region string) string
	ExtraArgs []string
}

// providers maps a CSP name to its hardcoded warp connection details.
// Endpoint/ExtraArgs values are filled in from real warp runs against each CSP.
var providers = map[string]providerConfig{
	"aws": {
		Endpoint: func(region string) string { return fmt.Sprintf("s3.%s.amazonaws.com", region) },
	},
	"ncp": {
		Endpoint: func(region string) string { return fmt.Sprintf("%s.object.ncloudstorage.com", region) },
	},
	"alibaba": {
		Endpoint: func(region string) string { return fmt.Sprintf("s3.oss-%s.aliyuncs.com", region) },
	},
	"nhn": {
		Endpoint: func(region string) string { return fmt.Sprintf("%s-api-object-storage.nhncloudservice.com", region) },
	},
	"kt": {
		Endpoint: func(region string) string { return "obj-e-1.ktcloud.com" },
		ExtraArgs: []string{"--tls"},
	},
	"ibm": {
		Endpoint: func(region string) string { return fmt.Sprintf("s3.%s.cloud-object-storage.appdomain.cloud", region) },
	},
	"tencent": {
		Endpoint:  func(region string) string { return fmt.Sprintf("cos.%s.myqcloud.com", region) },
		ExtraArgs: []string{"--lookup=host"},
	},
	"gcp": {
		Endpoint: func(region string) string { return "storage.googleapis.com" },
	},
}

// Endpoint returns the hardcoded S3-compatible endpoint host for the given provider+region.
func Endpoint(provider, region string) (string, error) {
	cfg, ok := providers[provider]
	if !ok {
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
	return cfg.Endpoint(region), nil
}

// ExtraArgs returns any additional warp CLI flags required for this provider.
func ExtraArgs(provider string) []string {
	return providers[provider].ExtraArgs
}

// ResolveS3Keys extracts the S3 access/secret key pair from a provider's credential struct.
func ResolveS3Keys(provider string, creds interface{}) (accessKey, secretKey string, err error) {
	switch provider {
	case "aws":
		c, ok := creds.(models.AWSCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for aws: expected AWSCredentials")
		}
		return c.AccessKey, c.SecretKey, nil
	case "ncp":
		c, ok := creds.(models.NCPCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for ncp: expected NCPCredentials")
		}
		return c.AccessKey, c.SecretKey, nil
	case "alibaba":
		c, ok := creds.(models.AlibabaCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for alibaba: expected AlibabaCredentials")
		}
		return c.AccessKey, c.SecretKey, nil
	case "gcp":
		c, ok := creds.(models.GCPCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for gcp: expected GCPCredentials")
		}
		return c.S3AccessKey, c.S3SecretKey, nil
	case "nhn":
		c, ok := creds.(models.NHNCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for nhn: expected NHNCredentials")
		}
		return c.S3AccessKey, c.S3SecretKey, nil
	case "kt":
		c, ok := creds.(models.KTCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for kt: expected KTCredentials")
		}
		return c.S3AccessKey, c.S3SecretKey, nil
	case "ibm":
		c, ok := creds.(models.IBMCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for ibm: expected IBMCredentials")
		}
		return c.S3AccessKey, c.S3SecretKey, nil
	case "tencent":
		c, ok := creds.(models.TencentCredentials)
		if !ok {
			return "", "", fmt.Errorf("invalid credentials for tencent: expected TencentCredentials")
		}
		return c.SecretId, c.SecretKey, nil
	default:
		return "", "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ResolveBucketName returns the actual bucket name to use for warp/S3 calls.
// Tencent COS bucket names must include the account's numeric AppId as a
// "-<appid>" suffix; every other provider uses the given bucketId as-is.
func ResolveBucketName(provider, secretId, secretKey, bucketId string) (string, error) {
	if provider != "tencent" {
		return bucketId, nil
	}

	client, err := cam.NewClientWithSecretId(secretId, secretKey, "")
	if err != nil {
		return "", fmt.Errorf("failed to create Tencent CAM client: %w", err)
	}

	resp, err := client.GetUserAppId(cam.NewGetUserAppIdRequest())
	if err != nil {
		return "", fmt.Errorf("failed to get Tencent AppId: %w", err)
	}
	if resp.Response == nil || resp.Response.AppId == nil {
		return "", fmt.Errorf("tencent GetUserAppId returned no AppId")
	}

	return fmt.Sprintf("%s-%d", bucketId, *resp.Response.AppId), nil
}
