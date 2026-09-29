// Package nrdbinstance orchestrates NRDB instance operations across CSPs:
// it resolves credentials by provider and dispatches to the matching
// provider implementation.
package nrdbinstance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance"
	alibabaprovider "github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance/alibaba"
	gcpprovider "github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance/gcp"
	ncpprovider "github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance/ncp"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/repository"

	"github.com/rs/zerolog/log"
)

// repo returns the namespace-scoped NRDB instance repository, backed by the
// shared config.DB connection (set up by config.InitDB() at server startup).
func repo() *repository.NRDBInstanceRepository {
	return repository.NewNRDBInstanceRepository(config.DB)
}

// randomSuffix returns n lowercase hex characters, used to make csp_instance_name
// unique across namespaces without eating into CSP name-length limits the way a
// full timestamp would.
func randomSuffix(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

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

func ListInstances(ctx context.Context, provider, region string) ([]models.NRDBInstance, error) {
	nsId := utils.GetNsId()

	// tbNrdbInstance 조회
	records, err := repo().FindByNamespace(provider, region, nsId)
	if err != nil {
		return nil, fmt.Errorf("failed to load NRDB instance records: %w", err)
	}
	if len(records) == 0 {
		return []models.NRDBInstance{}, nil
	}

	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return nil, err
	}

	cspInstances, err := p.ListInstances(ctx)
	if err != nil {
		return nil, err
	}
	cspByID := make(map[string]models.NRDBInstance, len(cspInstances))
	for _, inst := range cspInstances {
		cspByID[inst.InstanceID] = inst
	}

	// tbNrdbInstance와 싱크
	exists := []models.NRDBInstance{}
	var orphanIDs []string
	for _, record := range records {
		inst, ok := cspByID[record.InstanceID]
		if !ok {
			orphanIDs = append(orphanIDs, record.InstanceID)
			continue
		}

		// 사용자가 생성한 이름으로 교체
		inst.Name = record.InstanceName
		exists = append(exists, inst)
	}

	// csp에 없는 인스턴스 tbNrdbInstance에서 삭제
	if len(orphanIDs) > 0 {
		if err := repo().DeleteNRDBInstanceByID(provider, region, orphanIDs); err != nil {
			log.Error().Err(err).Str("provider", provider).Str("region", region).
				Msg("failed to remove orphaned NRDB instance records")
		}
	}

	return exists, nil
}

func CreateInstance(ctx context.Context, provider, region string, spec nrdbinstance.CreateSpec) (models.NRDBInstance, error) {
	nsId := utils.GetNsId()
	instanceName := spec.InstanceID

	// namespace별 instance_name(사용자 지정) 중복 가능
	if err := repo().CheckDuplicate(provider, region, nsId, instanceName); err != nil {
		return models.NRDBInstance{}, err
	}

	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return models.NRDBInstance{}, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return models.NRDBInstance{}, err
	}

	// 실제 csp에 생성되는 instance명
	cspInstanceName := fmt.Sprintf("%s-%s", instanceName, randomSuffix(6))
	spec.InstanceID = cspInstanceName

	instance, err := p.CreateInstance(ctx, spec)
	if err != nil {
		return models.NRDBInstance{}, err
	}

	record := &models.NRDBInstanceRecord{
		Provider:        provider,
		Region:          region,
		InstanceID:      instance.InstanceID,
		InstanceName:    instanceName,
		CspInstanceName: cspInstanceName,
		NamespaceID:     nsId,
	}
	if err := repo().CreateNRDBInstance(record); err != nil {
		return models.NRDBInstance{}, fmt.Errorf("CSP instance created but failed to save record: %w", err)
	}

	instance.Name = instanceName
	return instance, nil
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

	instance, err := p.DeleteInstance(ctx, instanceID)
	if err != nil {
		return models.NRDBInstance{}, err
	}

	if err := repo().DeleteNRDBInstanceByID(provider, region, []string{instanceID}); err != nil {
		log.Error().Err(err).Str("provider", provider).Str("region", region).Str("instanceId", instanceID).
			Msg("failed to remove NRDB instance record after CSP delete")
	}

	return instance, nil
}
