package openbao

import (
	"context"
	"fmt"
	"os"
	"strings"

	api "github.com/openbao/openbao/api/v2"
)

func newClient() (*api.Client, error) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		addr = "http://localhost:8200"
	}
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("VAULT_TOKEN is not set")
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenBao client: %w", err)
	}
	client.SetToken(token)
	return client, nil
}

// ReadSecret reads a KV v2 secret and returns the inner data map.
func ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenBao secret at %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid KV v2 format at %s: 'data' field missing or not a map", path)
	}
	return data, nil
}

// GetString safely extracts a string value from a data map.
func GetString(data map[string]interface{}, key string) string {
	v, _ := data[key].(string)
	return v
}

// ListProviders returns the CSP provider names stored under secret/metadata/csp in OpenBao.
func ListProviders(ctx context.Context) ([]string, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}

	secret, err := client.Logical().ListWithContext(ctx, "secret/metadata/csp")
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenBao secrets: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	providers := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			providers = append(providers, strings.TrimSuffix(s, "/"))
		}
	}
	return providers, nil
}
