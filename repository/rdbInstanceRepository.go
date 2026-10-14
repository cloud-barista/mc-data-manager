package repository

import (
	"errors"

	"github.com/cloud-barista/mc-data-manager/models"

	"gorm.io/gorm"
)

var ErrDuplicateRDBInstance = errors.New("rdb instance already exists")

type RDBInstanceRepository struct {
	db *gorm.DB
}

func NewRDBInstanceRepository(db *gorm.DB) *RDBInstanceRepository {
	return &RDBInstanceRepository{
		db: db,
	}
}

func (r *RDBInstanceRepository) FindByNamespace(provider, region, namespaceID string) ([]models.RDBInstanceRecord, error) {
	var records []models.RDBInstanceRecord
	if err := r.db.Where("provider = ? AND region = ? AND namespace_id = ?", provider, region, namespaceID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *RDBInstanceRepository) FindByInstanceID(provider, region, instanceID string) (*models.RDBInstanceRecord, error) {
	var record models.RDBInstanceRecord
	if err := r.db.Where("provider = ? AND region = ? AND instance_id = ?", provider, region, instanceID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *RDBInstanceRepository) CheckDuplicate(provider, region, namespaceID, instanceName string) error {
	var count int64
	r.db.Model(&models.RDBInstanceRecord{}).
		Where("provider = ? AND region = ? AND namespace_id = ? AND instance_name = ?", provider, region, namespaceID, instanceName).
		Count(&count)
	if count > 0 {
		return ErrDuplicateRDBInstance
	}
	return nil
}

func (r *RDBInstanceRepository) CreateRDBInstance(record *models.RDBInstanceRecord) error {
	return r.db.Create(record).Error
}

func (r *RDBInstanceRepository) DeleteRDBInstanceByID(provider, region string, instanceIDs []string) error {
	return r.db.Where("provider = ? AND region = ? AND instance_id IN ?", provider, region, instanceIDs).
		Delete(&models.RDBInstanceRecord{}).Error
}

func (r *RDBInstanceRepository) UpdatePublicNetPending(provider, region, instanceID string, pending bool) error {
	return r.db.Model(&models.RDBInstanceRecord{}).
		Where("provider = ? AND region = ? AND instance_id = ?", provider, region, instanceID).
		Update("public_net_pending", pending).Error
}

func (r *RDBInstanceRepository) UpdateAccountCreateFailed(provider, region, instanceID string, failed bool) error {
	return r.db.Model(&models.RDBInstanceRecord{}).
		Where("provider = ? AND region = ? AND instance_id = ?", provider, region, instanceID).
		Update("account_create_failed", failed).Error
}
