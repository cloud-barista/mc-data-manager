package rdbinstance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"encoding/json"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
	alibabaprovider "github.com/cloud-barista/mc-data-manager/pkg/rdbinstance/alibaba"
	awsprovider "github.com/cloud-barista/mc-data-manager/pkg/rdbinstance/aws"
	gcpprovider "github.com/cloud-barista/mc-data-manager/pkg/rdbinstance/gcp"
	ncpprovider "github.com/cloud-barista/mc-data-manager/pkg/rdbinstance/ncp"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/repository"

	"github.com/rs/zerolog/log"
)

// repo returns the namespace-scoped RDB instance repository, backed by the
// shared config.DB connection (set up by config.InitDB() at server startup).
func repo() *repository.RDBInstanceRepository {
	return repository.NewRDBInstanceRepository(config.DB)
}

func randomSuffix(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

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
	case "alibaba":
		alic, ok := creds.(models.AlibabaCredentials)
		if !ok {
			return nil, fmt.Errorf("invalid credentials for alibaba: expected AlibabaCredentials")
		}
		return alibabaprovider.New(alic.AccessKey, alic.SecretKey, region)
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

func ListInstances(ctx context.Context, provider, region string) ([]models.DBInstance, error) {
	nsId := utils.GetNsId()

	// tbRdbInstance 조회
	records, err := repo().FindByNamespace(provider, region, nsId)
	if err != nil {
		return nil, fmt.Errorf("failed to load RDB instance records: %w", err)
	}
	if len(records) == 0 {
		return []models.DBInstance{}, nil
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
	cspByID := make(map[string]models.DBInstance, len(cspInstances))
	for _, inst := range cspInstances {
		cspByID[inst.InstanceID] = inst
	}

	// tbRdbInstance와 싱크
	exists := []models.DBInstance{}
	var orphanIDs []string
	for _, record := range records {
		inst, ok := cspByID[record.InstanceID]
		if !ok {
			orphanIDs = append(orphanIDs, record.InstanceID)
			continue
		}

		// alibaba status
		if record.PublicNetPending {
			inst.Status = "creating"
		} else if record.AccountCreateFailed {
			inst.Status = "masteruser_failed"
		}

		// 사용자가 생상한 이름으로 교체
		inst.Name = record.InstanceName
		exists = append(exists, inst)
	}

	// csp에 없는 인스턴스 tbRdbInstance에서 삭제
	if len(orphanIDs) > 0 {
		if err := repo().DeleteRDBInstanceByID(provider, region, orphanIDs); err != nil {
			log.Error().Err(err).Str("provider", provider).Str("region", region).
				Msg("failed to remove orphaned RDB instance records")
		}
	}

	return exists, nil
}

func CreateInstance(ctx context.Context, provider, region string, spec rdbinstance.CreateSpec) (models.DBInstance, error) {
	nsId := utils.GetNsId()
	instanceName := spec.InstanceID

	// namespace별 instance_name(사용자 지정) 중복 가능
	if err := repo().CheckDuplicate(provider, region, nsId, instanceName); err != nil {
		return models.DBInstance{}, err
	}

	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return models.DBInstance{}, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return models.DBInstance{}, err
	}

	// 실제 csp에 생성되는 instance명
	cspInstanceName := fmt.Sprintf("%s-%s", instanceName, randomSuffix(6))
	spec.InstanceID = cspInstanceName

	instance, err := p.CreateInstance(ctx, spec)
	if err != nil {
		return models.DBInstance{}, err
	}

	record := &models.RDBInstanceRecord{
		Provider:         provider,
		Region:           region,
		InstanceID:       instance.InstanceID,
		InstanceName:     instanceName,
		CspInstanceName:  cspInstanceName,
		NamespaceID:      nsId,
		PublicNetPending: strings.ToLower(provider) == "alibaba",
	}
	if err := repo().CreateRDBInstance(record); err != nil {
		return models.DBInstance{}, fmt.Errorf("CSP instance created but failed to save record: %w", err)
	}

	instance.Name = instanceName
	return instance, nil
}

// DeleteInstance resolves credentials for the provider and deletes an instance.
func DeleteInstance(ctx context.Context, provider, region, instanceID string) (models.DBInstance, error) {
	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return models.DBInstance{}, fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return models.DBInstance{}, err
	}

	instance, err := p.DeleteInstance(ctx, instanceID)
	if err != nil {
		return models.DBInstance{}, err
	}

	if err := repo().DeleteRDBInstanceByID(provider, region, []string{instanceID}); err != nil {
		log.Error().Err(err).Str("provider", provider).Str("region", region).Str("instanceId", instanceID).
			Msg("failed to remove RDB instance record after CSP delete")
	}

	return instance, nil
}

func CreateAccount(ctx context.Context, provider, region, instanceID, username, password string) error {
	if strings.ToLower(provider) != "alibaba" {
		return fmt.Errorf("account creation API is only supported for alibaba")
	}

	// tbRdbInstance 상태 체크
	record, err := repo().FindByInstanceID(provider, region, instanceID)
	if err != nil {
		return fmt.Errorf("failed to load RDB instance record: %w", err)
	}
	if !record.AccountCreateFailed {
		return fmt.Errorf("account creation was not flagged as failed for this instance")
	}

	creds, err := config.NewAuthManager().LoadCredentialsByProvider(ctx, provider)
	if err != nil {
		return fmt.Errorf("credential load failed: %w", err)
	}

	p, err := providerFor(provider, creds, region)
	if err != nil {
		return err
	}

	ap, ok := p.(*alibabaprovider.AlibabaProvider)
	if !ok {
		return fmt.Errorf("unexpected provider implementation for alibaba")
	}

	if err := ap.CreateAccount(instanceID, username, password); err != nil {
		return err
	}

	return repo().UpdateAccountCreateFailed(provider, region, instanceID, false)
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
