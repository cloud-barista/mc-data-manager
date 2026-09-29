package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/repository"

	"github.com/google/uuid"
)

func repo() *repository.BackupRepository {
	return repository.NewBackupRepository(config.DB)
}

func CreateBackup(serviceType string, source models.BackupSourceCommon, backupName, databaseName string) (*models.BackupRecord, error) {
	nsId := utils.GetNsId()
	if err := repo().CheckDuplicate(nsId, serviceType, backupName); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve home directory: %w", err)
	}
	path := filepath.Join(home, "backup", id)
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	var dbNamePtr *string
	if databaseName != "" {
		dbNamePtr = &databaseName
	}

	record := &models.BackupRecord{
		ID:           id,
		NamespaceID:  nsId,
		ServiceType:  serviceType,
		BackupName:   backupName,
		Provider:     source.Provider,
		Region:       source.Region,
		InstanceId:   source.InstanceId,
		InstanceName: source.InstanceName,
		DatabaseName: dbNamePtr,
		Path:         path,
		Status:       "in_progress",
		CreatedAt:    time.Now().UTC(),
	}
	if err := repo().CreateBackup(record); err != nil {
		return nil, err
	}
	return record, nil
}

func MarkStatus(id string, success bool) (string, error) {
	status := "failed"
	if success {
		status = "completed"
	}
	if err := repo().UpdateStatus(id, status); err != nil {
		return "", err
	}
	return status, nil
}

func ToListResponse(record *models.BackupRecord) models.BackupListResponse {
	return models.BackupListResponse{
		ID:           record.ID,
		BackupName:   record.BackupName,
		Provider:     record.Provider,
		Region:       record.Region,
		InstanceId:   record.InstanceId,
		InstanceName: record.InstanceName,
		DatabaseName: record.DatabaseName,
		Status:       record.Status,
		CreatedAt:    record.CreatedAt,
	}
}

func ListBackups(serviceType string) ([]models.BackupListResponse, error) {
	nsId := utils.GetNsId()
	records, err := repo().FindByNamespace(nsId, serviceType)
	if err != nil {
		return nil, err
	}

	responses := make([]models.BackupListResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, ToListResponse(&record))
	}
	return responses, nil
}

func GetRestorableBackup(id, serviceType string) (*models.BackupRecord, error) {
	record, err := repo().FindByID(id)
	if err != nil {
		return nil, err
	}
	if record.ServiceType != serviceType {
		return nil, fmt.Errorf("this is a %s backup, not for %s", record.ServiceType, serviceType)
	}
	if record.Status != "completed" {
		return nil, fmt.Errorf("backup is not completed (status: %s)", record.Status)
	}
	if !utils.DirExists(record.Path) {
		return nil, fmt.Errorf("backup files not found at %s. Please remove this backup", record.Path)
	}
	return record, nil
}

func DeleteBackup(id string) error {
	record, err := repo().FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to load backup record: %w", err)
	}

	if err := repo().DeleteByID(id); err != nil {
		return fmt.Errorf("failed to delete backup record: %w", err)
	}

	if err := os.RemoveAll(record.Path); err != nil {
		return fmt.Errorf("backup record deleted but failed to remove files at %s: %w", record.Path, err)
	}

	return nil
}
