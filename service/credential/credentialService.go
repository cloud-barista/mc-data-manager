package service

import (
	"encoding/json"
	"errors"

	// "fmt"
	// "strings"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/repository"
	"gorm.io/gorm"
)

type CredentialService struct {
	db                   *gorm.DB
	credentialRepository *repository.CredentialRepository
	AesConverter         *utils.AESconverter
}

func NewCredentialService(db *gorm.DB) *CredentialService {
	credentialRepository := repository.NewCredentialRepository(db)
	aesConverter := utils.NewAESConverter()

	return &CredentialService{
		db:                   db,
		credentialRepository: credentialRepository,
		AesConverter:         aesConverter,
	}
}

// TODO - 이름 중복 체크 추가
func (c *CredentialService) CreateCredential(req models.CredentialCreateRequest) (*models.Credential, error) {
	// if existing, err := c.credentialRepository.FindByName(req.Name); err == nil && existing != nil {
	//     return nil, fmt.Errorf("credential with name '%s' already exists", req.Name)
	// } else if !errors.Is(err, gorm.ErrRecordNotFound) {
	//     return nil, err
	// }
	jsonBytes, _ := json.Marshal(req.GetCredential())
	encoded, err := c.AesConverter.EncryptAESGCM(string(jsonBytes))
	if err != nil {
		return nil, err
	}
	cred := models.Credential{
		Name:           req.Name,
		CspType:        req.CspType,
		CredentialJson: encoded,
	}

	if err := c.credentialRepository.CreateCredential(&cred); err != nil {
		// if errors.Is(err, gorm.ErrDuplicatedKey) {
		//     return nil, fmt.Errorf("credential name '%s' already exists", req.Name)
		// }

		// if strings.Contains(err.Error(), "duplicate key") {
		//     return nil, fmt.Errorf("credential name '%s' already exists", req.Name)
		// }
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

func (c *CredentialService) GetCredentialById(id uint64) (*models.Credential, error) {
	return c.credentialRepository.GetCredentialByID(id)
}

func (c *CredentialService) UpdateCredential(id uint64, req models.Credential) (*models.Credential, error) {
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

func (c *CredentialService) DeleteCredential(id uint64) error {
	return c.credentialRepository.DeleteCredential(id)
}
