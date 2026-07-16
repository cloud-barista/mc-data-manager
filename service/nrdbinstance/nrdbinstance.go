// Package nrdbinstance orchestrates NRDB instance operations across CSPs:
// it resolves credentials by provider and dispatches to the matching
// provider implementation.
package nrdbinstance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance"
	alibabaprovider "github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance/alibaba"
	gcpprovider "github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance/gcp"
	ncpprovider "github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance/ncp"
)

// providerFor selects and constructs the provider implementation for the given
// CSP, using the supplied credentials and region.
func providerFor(provider string, creds interface{}, region string) (nrdbinstance.Provider, error) {
	switch strings.ToLower(provider) {
	case "gcp":
		gcpc, ok := creds.(models.GCPCredentials)
		if !ok {
			return nil, fmt.Errorf("invalid credentials for gcp: expected GCPCredentials")
		}
		credJSON, err := json.Marshal(gcpc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal GCP credentials: %w", err)
		}
		return gcpprovider.New(credJSON, gcpc.ProjectID, region)
	case "alibaba":
		alic, ok := creds.(models.AlibabaCredentials)
		if !ok {
			return nil, fmt.Errorf("invalid credentials for alibaba: expected AlibabaCredentials")
		}
		return alibabaprovider.New(alic.AccessKey, alic.SecretKey, region)
	case "ncp":
		ncpc, ok := creds.(models.NCPCredentials)
		if !ok {
			return nil, fmt.Errorf("invalid credentials for ncp: expected NCPCredentials")
		}
		return ncpprovider.New(ncpc.AccessKey, ncpc.SecretKey, region)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ListInstances resolves credentials for the provider and returns its NRDB instances.
func ListInstances(ctx context.Context, provider, region string) ([]models.NRDBInstance, error) {
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

// CreateInstance resolves credentials for the provider and provisions an NRDB instance.
func CreateInstance(ctx context.Context, provider, region string, spec nrdbinstance.CreateSpec) (models.NRDBInstance, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return models.NRDBInstance{}, err
	}

	return p.CreateInstance(ctx, spec)
}

// ListEngineVersions returns available engine versions for the provider.
func ListEngineVersions(ctx context.Context, provider, region string) ([]models.NRDBEngineVersion, error) {
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

// ListInstanceClasses returns orderable instance classes for the given engine version.
func ListInstanceClasses(ctx context.Context, provider, region, engineVersion string) ([]string, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return nil, err
	}

	return p.ListInstanceClasses(ctx, engineVersion)
}

// DeleteInstance resolves credentials for the provider and deletes an NRDB instance.
func DeleteInstance(ctx context.Context, provider, region, instanceID string) (models.NRDBInstance, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return models.NRDBInstance{}, err
	}

	return p.DeleteInstance(ctx, instanceID)
}
