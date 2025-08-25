package service

import (
	"encoding/json"
	"errors"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/repository"
	"gorm.io/gorm"
)

type CredentialService struct {
	db                   *gorm.DB
	credentialRepository *repository.CredentialRepository
	aesConverter         *utils.AESconverter
}

func NewCredentialService(db *gorm.DB) *CredentialService {
	credentialRepository := repository.NewCredentialRepository(db)
	aesConverter := utils.NewAESConverter()

	return &CredentialService{
		db:                   db,
		credentialRepository: credentialRepository,
		aesConverter:         aesConverter,
	}
}

func (c *CredentialService) CreateCredential(req models.CredentialCreateRequest) (*models.Credential, error) {
	jsonBytes, _ := json.Marshal(req.GetCredential())
	encoded, err := c.aesConverter.EncryptAESGCM(string(jsonBytes))
	if err != nil {
		return nil, err
	}

	cred := models.Credential{
		Name:           req.Name,
		CspType:        req.CspType,
		CredentialJson: encoded, // TODO - 암호화된 JSON 문자열
	}

	if err := c.credentialRepository.CreateCredential(&cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

func (c *CredentialService) ListCredentials() ([]models.CredentialListResponse, error) {
	credentials, err := c.credentialRepository.ListCredentials()
	if err != nil {
		return nil, err
	}

	responses := make([]models.CredentialListResponse, len(credentials))
	for i, u := range credentials {
		responses[i] = models.CredentialListResponse{
			CredentialId: u.CredentialId,
			CspType:      u.CspType,
			Name:         u.Name,
		}
	}

	return responses, nil
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
