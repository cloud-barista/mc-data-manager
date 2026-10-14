package repository

import (
	"errors"

	"github.com/cloud-barista/mc-data-manager/models"

	"gorm.io/gorm"
)

var ErrDuplicateBackup = errors.New("backup already exists")

type BackupRepository struct {
	db *gorm.DB
}

func NewBackupRepository(db *gorm.DB) *BackupRepository {
	return &BackupRepository{
		db: db,
	}
}

func (r *BackupRepository) CheckDuplicate(namespaceID, serviceType, backupName string) error {
	var count int64
	r.db.Model(&models.BackupRecord{}).
		Where("namespace_id = ? AND service_type = ? AND backup_name = ?", namespaceID, serviceType, backupName).
		Count(&count)
	if count > 0 {
		return ErrDuplicateBackup
	}
	return nil
}

func (r *BackupRepository) CreateBackup(record *models.BackupRecord) error {
	return r.db.Create(record).Error
}

func (r *BackupRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&models.BackupRecord{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *BackupRepository) FindByNamespace(namespaceID, serviceType string) ([]models.BackupRecord, error) {
	var records []models.BackupRecord
	query := r.db.Where("namespace_id = ?", namespaceID)
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *BackupRepository) FindByID(id string) (*models.BackupRecord, error) {
	var record models.BackupRecord
	if err := r.db.Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *BackupRepository) DeleteByID(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.BackupRecord{}).Error
}
