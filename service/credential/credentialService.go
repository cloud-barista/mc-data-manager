package service

import (
	"errors"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/repository"
	"gorm.io/gorm"
)

type CredentialService struct {
	db                   *gorm.DB
	credentialRepository *repository.CredentialRepository
}

func NewCredentialService(db *gorm.DB) *CredentialService {
	credentialRepository := repository.NewCredentialRepository(db)

	return &CredentialService{
		db:                   db,
		credentialRepository: credentialRepository,
	}
}

func (c *CredentialService) CreateCredential(req models.Credential) (*models.Credential, error) {
	cred := models.Credential{
		Name:           req.Name,
		CspType:        req.CspType,
		CredentialJson: req.CredentialJson, // 암호화된 JSON 문자열
	}

	if err := c.credentialRepository.CreateCredential(&cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

func (c *CredentialService) ListCredentials() ([]models.Credential, error) {
	return c.credentialRepository.ListCredentials()
}

func (c *CredentialService) GetCredentialById(id string) (*models.Credential, error) {
	return c.credentialRepository.GetCredentialByID(id)
}

func (c *CredentialService) UpdateCredential(id string, req models.Credential) (*models.Credential, error) {
	cred, err := c.credentialRepository.GetCredentialByID(id)
	if err != nil {
		return nil, errors.New("not found")
	}

	cred.Name = req.Name
	cred.CspType = req.CspType
	cred.CredentialJson = req.CredentialJson
	if err := c.credentialRepository.UpdateCredential(cred); err != nil {
		return nil, err
	}

	return cred, nil
}

func (c *CredentialService) DeleteCredential(id string) error {
	return c.credentialRepository.DeleteCredential(id)
}
