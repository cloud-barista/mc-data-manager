// Package rdbinstance defines a CSP-agnostic interface for managing RDB
// (database) infrastructure instances. Concrete implementations live in
// per-provider subpackages (aws, gcp, ncp, alibaba).
package rdbinstance

import (
	"context"

	"github.com/cloud-barista/mc-data-manager/models"
)

// CreateSpec is a CSP-agnostic specification for creating an RDB instance.
type CreateSpec struct {
	InstanceID       string
	InstanceClass    string
	Engine           string
	EngineVersion    string
	MasterUsername   string
	MasterPassword   string
	AllocatedStorage int32
}

// Provider abstracts a single CSP's RDB instance operations.
// Delete is added later.
type Provider interface {
	ListInstances(ctx context.Context) ([]models.DBInstance, error)
	CreateInstance(ctx context.Context, spec CreateSpec) (models.DBInstance, error)
	DeleteInstance(ctx context.Context, instanceID string) (models.DBInstance, error)
	ListEngineVersions(ctx context.Context) ([]models.DBEngineVersion, error)
	ListInstanceClasses(ctx context.Context, engine, engineVersion string) ([]string, error)
}
