package repository

import (
	"github.com/cloud-barista/mc-data-manager/models"

	"gorm.io/gorm"
)

type CredentialRepository struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{
		db: db,
	}
}

func (c *CredentialRepository) CreateCredential(cred *models.Credential) error {
	return c.db.Create(cred).Error
}

func (c *CredentialRepository) GetCredentialByID(id uint64) (*models.Credential, error) {
	var cred models.Credential

	if err := c.db.First(&cred, "credentialId = ?", id).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}

func (c *CredentialRepository) ListCredentials() ([]models.Credential, error) {
	var creds []models.Credential
	if err := c.db.Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

func (c *CredentialRepository) UpdateCredential(cred *models.Credential) error {
	return c.db.Save(cred).Error
}

func (c *CredentialRepository) DeleteCredential(id uint64) error {
	return c.db.Delete(&models.Credential{}, "credentialId = ?", id).Error
}

func(c *CredentialRepository) FindByName(name string) (*models.Credential, error) {
	var cred models.Credential
	if err := c.db.Where("name = ?", name).First(&cred).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}

func(c *CredentialRepository) CheckNameDuplicate(name string, cspType string) (*models.Credential, error) {
	var cred models.Credential
	if err := c.db.Where("name = ? AND cspType = ?", name, cspType).First(&cred).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}

