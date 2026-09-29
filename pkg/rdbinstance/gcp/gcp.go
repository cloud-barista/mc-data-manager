// Package gcp implements rdbinstance.Provider for GCP Cloud SQL.
//
// GCP Cloud SQL combines engine and version into a single DatabaseVersion string
// (e.g. MYSQL_8_0). This provider splits them: engine = "MYSQL",
// engineVersion = "8_0" (GCP native underscore format). On create, they are
// joined as databaseVersion = engine + "_" + engineVersion.
//
// GCP Cloud SQL does not support MariaDB; only MySQL is available.
// CreateInstance is not yet implemented.
package gcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
	"google.golang.org/api/googleapi"
	sqladmin "google.golang.org/api/sqladmin/v1"
)

// GCPProvider implements rdbinstance.Provider for GCP Cloud SQL.
type GCPProvider struct {
	svc     *sqladmin.Service
	project string
	region  string
}

// New builds a GCP Cloud SQL provider from a service account credentials JSON,
// project ID, and region.
func New(credJSON []byte, projectID, region string) (rdbinstance.Provider, error) {
	svc, err := config.NewGCPSQLAdminClient(credJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud SQL client: %w", err)
	}
	return &GCPProvider{svc: svc, project: projectID, region: region}, nil
}

// supportedEngines lists the Cloud SQL engines this provider exposes.
// GCP native casing: all uppercase. MariaDB is not available in Cloud SQL.
var supportedEngines = []string{"MYSQL"}

// engineVersions maps each engine to its available versions in GCP native format
// (underscore-separated suffix matching the DatabaseVersion enum).
var engineVersions = map[string][]string{
	"MYSQL": {"5_7", "8_0", "8_4"},
}

// ListEngineVersions returns the supported engine versions.
// GCP Cloud SQL has no public API to list DatabaseVersions, so the list is fixed.
func (p *GCPProvider) ListEngineVersions(_ context.Context) ([]models.DBEngineVersion, error) {
	var out []models.DBEngineVersion
	for _, engine := range supportedEngines {
		for _, ver := range engineVersions[engine] {
			out = append(out, models.DBEngineVersion{
				Engine:        engine,
				EngineVersion: ver,
			})
		}
	}
	return out, nil
}

// ListInstanceClasses returns the available Cloud SQL tier names for the project.
// GCP tiers are not engine-specific; engine and engineVersion are accepted for
// interface compatibility but are unused.
func (p *GCPProvider) ListInstanceClasses(ctx context.Context, _, _ string) ([]string, error) {
	resp, err := p.svc.Tiers.List(p.project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list Cloud SQL tiers: %w", err)
	}
	classes := make([]string, 0, len(resp.Items))
	for _, t := range resp.Items {
		classes = append(classes, t.Tier)
	}
	sort.Strings(classes)
	return classes, nil
}

// splitDatabaseVersion splits a GCP DatabaseVersion string (e.g. "MYSQL_8_0")
// into engine prefix ("MYSQL") and version suffix ("8_0").
func splitDatabaseVersion(dbVersion string) (engine, version string) {
	engine, version, _ = strings.Cut(dbVersion, "_")
	return
}

// normalizeState maps GCP Cloud SQL instance states to lowercase equivalents,
// mapping "RUNNABLE" to "available" to match the AWS/Alibaba convention.
func normalizeState(state string) string {
	switch state {
	case "RUNNABLE":
		return "available"
	case "PENDING_CREATE":
		return "creating"
	case "PENDING_DELETE":
		return "deleting"
	default:
		return strings.ToLower(state)
	}
}

// toDBInstance converts a GCP DatabaseInstance into the CSP-agnostic model.
func toDBInstance(inst *sqladmin.DatabaseInstance, region string) models.DBInstance {
	engine, engineVersion := splitDatabaseVersion(inst.DatabaseVersion)
	tier := ""
	if inst.Settings != nil {
		tier = inst.Settings.Tier
	}

	endpoint := ""
	for _, a := range inst.IpAddresses {
		if a.Type == "PRIMARY" {
			endpoint = a.IpAddress
			break
		}
		if a.Type == "PRIVATE" {
			endpoint = "-"
		}
	}

	return models.DBInstance{
		Provider:      "gcp",
		InstanceID:    inst.Name,
		Name:          inst.Name,
		Engine:        engine,
		EngineVersion: engineVersion,
		Status:        normalizeState(inst.State),
		Endpoint:      endpoint,
		Port:          3306,
		Region:        region,
		InstanceClass: tier,
	}
}

// ListInstances returns all Cloud SQL instances in the provider's region.
func (p *GCPProvider) ListInstances(ctx context.Context) ([]models.DBInstance, error) {
	resp, err := p.svc.Instances.List(p.project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list Cloud SQL instances: %w", err)
	}
	out := []models.DBInstance{}
	for _, inst := range resp.Items {
		if inst.Region != p.region {
			continue
		}
		out = append(out, toDBInstance(inst, p.region))
	}
	return out, nil
}

// DeleteInstance deletes a Cloud SQL instance. The delete response carries no
// instance details, so the returned model is constructed from the instance ID.
func (p *GCPProvider) DeleteInstance(ctx context.Context, instanceID string) (models.DBInstance, error) {
	op, err := p.svc.Instances.Delete(p.project, instanceID).Context(ctx).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && (gerr.Code == http.StatusNotFound || gerr.Code == http.StatusForbidden) {
			return models.DBInstance{Provider: "gcp", InstanceID: instanceID, Status: "deleted", Region: p.region}, nil
		}
		return models.DBInstance{}, fmt.Errorf("failed to delete Cloud SQL instance: %w", err)
	}
	return models.DBInstance{
		Provider:   "gcp",
		InstanceID: op.TargetId,
		Name:       op.TargetId,
		Status:     "deleting",
		Region:     p.region,
	}, nil
}

func buildCreateRequest(spec rdbinstance.CreateSpec, region string) *sqladmin.DatabaseInstance {
	return &sqladmin.DatabaseInstance{
		Name:            spec.InstanceID,
		DatabaseVersion: spec.Engine + "_" + spec.EngineVersion,
		Region:          region,
		RootPassword:    spec.MasterPassword,
		Settings: &sqladmin.Settings{
			Edition:		"ENTERPRISE",
			Tier:           spec.InstanceClass,
			DataDiskSizeGb: int64(spec.AllocatedStorage),
			IpConfiguration: &sqladmin.IpConfiguration{
				Ipv4Enabled: true,
				AuthorizedNetworks: []*sqladmin.AclEntry{
					{Value: config.OutboundIP},
				},
			},
		},
	}
}

func (p *GCPProvider) CreateInstance(ctx context.Context, spec rdbinstance.CreateSpec) (models.DBInstance, error) {
	if spec.MasterUsername != "root" {
		return models.DBInstance{}, fmt.Errorf("GCP Cloud SQL only supports 'root' as masterUsername for MySQL")
	}

	op, err := p.svc.Instances.Insert(p.project, buildCreateRequest(spec, p.region)).Context(ctx).Do()
	if err != nil {
		return models.DBInstance{}, fmt.Errorf("failed to create Cloud SQL instance: %w", err)
	}

	return models.DBInstance{
		Provider:      "gcp",
		InstanceID:    op.TargetId,
		Name:          op.TargetId,
		Engine:        spec.Engine,
		EngineVersion: spec.EngineVersion,
		Status:        "creating",
		Port:          3306,
		Region:        p.region,
		InstanceClass: spec.InstanceClass,
	}, nil
}
