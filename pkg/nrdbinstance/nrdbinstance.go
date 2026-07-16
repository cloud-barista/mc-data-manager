// Package nrdbinstance defines a CSP-agnostic interface for managing NoSQL
// database instances. Concrete implementations live in per-provider subpackages
// (gcp). NCP and Alibaba will be added in subsequent tasks.
package nrdbinstance

import (
	"context"

	"github.com/cloud-barista/mc-data-manager/models"
)

// CreateSpec is a CSP-agnostic specification for creating an NRDB instance.
type CreateSpec struct {
	InstanceID       string
	EngineVersion    string
	InstanceClass    string
	AllocatedStorage int32
	MasterUsername   string
	MasterPassword   string
}

// Provider abstracts a single CSP's NRDB instance operations.
type Provider interface {
	ListInstances(ctx context.Context) ([]models.NRDBInstance, error)
	CreateInstance(ctx context.Context, spec CreateSpec) (models.NRDBInstance, error)
	DeleteInstance(ctx context.Context, instanceID string) (models.NRDBInstance, error)
	ListEngineVersions(ctx context.Context) ([]models.NRDBEngineVersion, error)
	ListInstanceClasses(ctx context.Context, engineVersion string) ([]string, error)
}
