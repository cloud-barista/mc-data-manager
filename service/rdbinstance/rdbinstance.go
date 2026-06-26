// Package rdbinstance orchestrates RDB (database) instance operations across
// CSPs: it resolves credentials by provider and dispatches to the matching
// provider implementation.
package rdbinstance

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
	awsprovider "github.com/cloud-barista/mc-data-manager/pkg/rdbinstance/aws"
)

// providerFor selects and constructs the provider implementation for the given
// CSP, using the supplied credentials and region.
func providerFor(provider string, creds interface{}, region string) (rdbinstance.Provider, error) {
	switch strings.ToLower(provider) {
	case "aws":
		awsc, ok := creds.(models.AWSCredentials)
		if !ok {
			return nil, fmt.Errorf("invalid credentials for aws: expected AWSCredentials")
		}
		return awsprovider.New(awsc.AccessKey, awsc.SecretKey, region)
	case "gcp", "ncp", "alibaba":
		return nil, fmt.Errorf("provider %q is not implemented yet", strings.ToLower(provider))
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ListInstances resolves credentials for the provider and returns its instances.
func ListInstances(ctx context.Context, provider, region string) ([]models.DBInstance, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return nil, err
	}

	return p.ListInstances(ctx)
}

// CreateInstance resolves credentials for the provider and provisions an instance.
func CreateInstance(ctx context.Context, provider, region string, spec rdbinstance.CreateSpec) (models.DBInstance, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return models.DBInstance{}, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return models.DBInstance{}, err
	}

	return p.CreateInstance(ctx, spec)
}

// ListEngineVersions returns available DB engine versions for the provider.
func ListEngineVersions(ctx context.Context, provider, region string) ([]models.DBEngineVersion, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return nil, err
	}

	return p.ListEngineVersions(ctx)
}

// ListInstanceClasses returns orderable instance classes for engine+version.
func ListInstanceClasses(ctx context.Context, provider, region, engine, engineVersion string) ([]string, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return nil, err
	}

	return p.ListInstanceClasses(ctx, engine, engineVersion)
}
