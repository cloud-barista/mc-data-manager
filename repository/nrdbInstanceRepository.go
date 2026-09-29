package repository

import (
	"errors"

	"github.com/cloud-barista/mc-data-manager/models"

	"gorm.io/gorm"
)

var ErrDuplicateNRDBInstance = errors.New("nrdb instance already exists")

type NRDBInstanceRepository struct {
	db *gorm.DB
}

func NewNRDBInstanceRepository(db *gorm.DB) *NRDBInstanceRepository {
	return &NRDBInstanceRepository{
		db: db,
	}
}

func (r *NRDBInstanceRepository) FindByNamespace(provider, region, namespaceID string) ([]models.NRDBInstanceRecord, error) {
	var records []models.NRDBInstanceRecord
	if err := r.db.Where("provider = ? AND region = ? AND namespace_id = ?", provider, region, namespaceID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *NRDBInstanceRepository) CheckDuplicate(provider, region, namespaceID, instanceName string) error {
	var count int64
	r.db.Model(&models.NRDBInstanceRecord{}).
		Where("provider = ? AND region = ? AND namespace_id = ? AND instance_name = ?", provider, region, namespaceID, instanceName).
		Count(&count)
	if count > 0 {
		return ErrDuplicateNRDBInstance
	}
	return nil
}

func (r *NRDBInstanceRepository) CreateNRDBInstance(record *models.NRDBInstanceRecord) error {
	return r.db.Create(record).Error
}

func (r *NRDBInstanceRepository) DeleteNRDBInstanceByID(provider, region string, instanceIDs []string) error {
	return r.db.Where("provider = ? AND region = ? AND instance_id IN ?", provider, region, instanceIDs).
		Delete(&models.NRDBInstanceRecord{}).Error
}
