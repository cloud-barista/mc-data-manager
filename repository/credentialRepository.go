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

func (c *CredentialRepository) GetCredentialByID(id string) (*models.Credential, error) {
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

func (c *CredentialRepository) DeleteCredential(id string) error {
	return c.db.Delete(&models.Credential{}, "credentialId = ?", id).Error
}
