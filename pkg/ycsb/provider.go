package ycsb

import (
	"context"
	"fmt"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
)

// ResolveAWSCredentials fetches AWS access/secret keys for the dynamodb binding.
func ResolveAWSCredentials(ctx context.Context) (accessKey, secretKey string, err error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, "aws")
	if err != nil {
		return "", "", fmt.Errorf("credential load failed: %w", err)
	}
	c, ok := creds.(models.AWSCredentials)
	if !ok {
		return "", "", fmt.Errorf("invalid credentials for aws: expected AWSCredentials")
	}
	return c.AccessKey, c.SecretKey, nil
}

// BuildWorkloadArgs builds the binding-independent workload flags shared by
// every provider. The fields are validated by the handler (counts > 0,
// proportions summing to 1), so every flag is always emitted — even a 0
// proportion must override the workload file's 0.5 default.
func BuildWorkloadArgs(req models.YcsbDiagnosticRequest) []string {
	return []string{
		"-P", "workloads/workloada",
		"-p", fmt.Sprintf("recordcount=%d", req.RecordCount),
		"-p", fmt.Sprintf("threadcount=%d", req.ThreadCount),
		"-p", fmt.Sprintf("readproportion=%g", req.ReadProportion),
		"-p", fmt.Sprintf("updateproportion=%g", req.UpdateProportion),
	}
}

func BuildDynamoDBArgs(accessKey, secretKey, region, tableName string) (env []string, args []string) {
	env = []string{
		"AWS_ACCESS_KEY_ID=" + accessKey,
		"AWS_SECRET_ACCESS_KEY=" + secretKey,
		"AWS_REGION=" + region,
	}
	args = []string{
		"-p", "dynamodb.tablename=" + tableName,
		"-p", "dynamodb.primarykey=ycsb_pk",
		"-p", "dynamodb.endpoint=https://dynamodb." + region + ".amazonaws.com",
		"-p", "dynamodb.delete.after.run.stage=true",
		"-p", "dynamodb.rc.units=1000",
		"-p", "dynamodb.wc.units=1000",
	}
	return env, args
}

func BuildMongoDBArgs(host, port, username, password, tableName string) []string {
	return []string{
		"-p", fmt.Sprintf("mongodb.url=mongodb://%s:%s/?directConnection=true", host, port),
		"-p", "mongodb.tls_skip_verify=true",
		"-p", "table=" + tableName,
		"-p", "mongodb.authdb=admin",
		"-p", "mongodb.username=" + username,
		"-p", "mongodb.password=" + password,
	}
}
