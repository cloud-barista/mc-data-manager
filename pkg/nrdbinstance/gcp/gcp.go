// Package gcp implements nrdbinstance.Provider for GCP Firestore.
//
// GCP의 NoSQL 인스턴스 단위는 "Firestore Database"이다.
// 하나의 GCP 프로젝트 안에 여러 Firestore Database를 생성할 수 있으며,
// Firestore Admin API(google.golang.org/api/firestore/v1)를 통해 관리한다.
//
// CreateSpec.Type이 비어 있으면 "FIRESTORE_NATIVE"를 기본값으로 사용한다.
package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance"
	firestorev1 "google.golang.org/api/firestore/v1"
)

// GCPProvider implements nrdbinstance.Provider for GCP Firestore.
type GCPProvider struct {
	svc     *firestorev1.Service
	project string
	region  string
}

// New builds a GCP Firestore provider from a service account credentials JSON,
// project ID, and region.
func New(credJSON []byte, projectID, region string) (nrdbinstance.Provider, error) {
	svc, err := config.NewGCPFirestoreAdminClient(credJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore Admin client: %w", err)
	}
	return &GCPProvider{svc: svc, project: projectID, region: region}, nil
}

// databaseIDFromName extracts the database ID from a full resource name.
// Input: "projects/{project}/databases/{database}" → "{database}"
func databaseIDFromName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) == 0 {
		return name
	}
	return parts[len(parts)-1]
}

// toNRDBInstance converts a Firestore GoogleFirestoreAdminV1Database to NRDBInstance.
// The Admin API v0.194.0 does not expose a state field on the database resource;
// a database returned by List is always reachable, so status is "available".
func toNRDBInstance(db *firestorev1.GoogleFirestoreAdminV1Database, region string) models.NRDBInstance {
	id := databaseIDFromName(db.Name)
	return models.NRDBInstance{
		Provider:   "gcp",
		InstanceID: id,
		Name:       id,
		Status:     "available",
		Region:     region,
	}
}

// ListInstances returns all Firestore databases in the provider's project.
func (p *GCPProvider) ListInstances(ctx context.Context) ([]models.NRDBInstance, error) {
	parent := fmt.Sprintf("projects/%s", p.project)
	resp, err := p.svc.Projects.Databases.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list Firestore databases: %w", err)
	}
	out := []models.NRDBInstance{}
	for _, db := range resp.Databases {
		out = append(out, toNRDBInstance(db, p.region))
	}
	return out, nil
}

// CreateInstance provisions a new Firestore database (FIRESTORE_NATIVE type).
func (p *GCPProvider) CreateInstance(ctx context.Context, spec nrdbinstance.CreateSpec) (models.NRDBInstance, error) {
	parent := fmt.Sprintf("projects/%s", p.project)
	db := &firestorev1.GoogleFirestoreAdminV1Database{
		Type:       "FIRESTORE_NATIVE",
		LocationId: p.region,
	}

	op, err := p.svc.Projects.Databases.Create(parent, db).DatabaseId(spec.InstanceID).Context(ctx).Do()
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("failed to create Firestore database: %w", err)
	}
	_ = op

	return models.NRDBInstance{
		Provider:   "gcp",
		InstanceID: spec.InstanceID,
		Name:       spec.InstanceID,
		Status:     "creating",
		Region:     p.region,
	}, nil
}

// ListEngineVersions returns an empty list; Firestore has no engine versioning concept.
func (p *GCPProvider) ListEngineVersions(_ context.Context) ([]models.NRDBEngineVersion, error) {
	return []models.NRDBEngineVersion{}, nil
}

// ListInstanceClasses returns an empty list; Firestore has no instance class concept.
func (p *GCPProvider) ListInstanceClasses(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

// DeleteInstance deletes a Firestore database by its ID.
func (p *GCPProvider) DeleteInstance(ctx context.Context, instanceID string) (models.NRDBInstance, error) {
	name := fmt.Sprintf("projects/%s/databases/%s", p.project, instanceID)
	op, err := p.svc.Projects.Databases.Delete(name).Context(ctx).Do()
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("failed to delete Firestore database: %w", err)
	}
	_ = op

	return models.NRDBInstance{
		Provider:   "gcp",
		InstanceID: instanceID,
		Name:       instanceID,
		Status:     "deleting",
		Region:     p.region,
	}, nil
}
