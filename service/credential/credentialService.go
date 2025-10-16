package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

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
	// 	return nil, fmt.Errorf("credential with name '%s' already exists", req.Name)
	// } else if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 	return nil, err
	// }

	// if existing, err := c.credentialRepository.CheckNameDuplicate(req.Name, req.CspType); err == nil && existing != nil {
	// 	return nil, fmt.Errorf("credential with name '%s' already exists", req.Name)
	// } else if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 	return nil, err
	// }

	if existing, err := c.credentialRepository.CheckProviderDuplicate(req.CspType); err == nil && existing != nil {
		return nil, fmt.Errorf("credential of '%s' already exists", req.CspType)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	credStr, err := req.GetCredential()
	if err != nil {
		return nil, err
	}

	if slices.Contains([]string{"aws", "ncp", "gcp"}, req.CspType) {
		terr := createTumblebugCredential(req)
		if terr != nil {
			return nil, terr
		}
	}

	encoded, err := c.AesConverter.EncryptAESGCM(credStr)
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

func createTumblebugCredential(req models.CredentialCreateRequest) error {
	publicKey, publicKeyTokenId, err := getPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	keyValues, err := getCredentialKeyValues(req)
	if err != nil {
		return fmt.Errorf("invalid credential json: %w", err)
	}

	encryptedCredentials, encryptedAesKey, err := encryptCredentialsWithPublicKey(publicKey, keyValues)
	if err != nil {
		return fmt.Errorf("invalid credential json: %w", err)
	}

	payload := map[string]interface{}{
		"credentialHolder":                 "admin",
		"providerName":                     req.CspType,
		"publicKeyTokenId":                 publicKeyTokenId,
		"encryptedClientAesKeyByPublicKey": encryptedAesKey,
		"credentialKeyValueList":           encryptedCredentials,
	}

	cerr := sendCredentials(payload)
	if cerr != nil {
		return fmt.Errorf("create credential failed: %w", cerr)
	}
	return nil
}

func getCredentialKeyValues(req models.CredentialCreateRequest) (map[string]string, error) {
	switch req.CspType {
	case "aws":
		var aws models.AWSCredentials
		if err := json.Unmarshal(req.CredentialJson, &aws); err != nil {
			return nil, fmt.Errorf("invalid aws credential json: %w", err)
		}

		return map[string]string{
			"ClientId":     aws.AccessKey,
			"ClientSecret": aws.SecretKey,
		}, nil
	case "ncp":
		var ncp models.NCPCredentials
		if err := json.Unmarshal(req.CredentialJson, &ncp); err != nil {
			return nil, fmt.Errorf("invalid ncp credential json: %w", err)
		}

		return map[string]string{
			"ClientId":     ncp.AccessKey,
			"ClientSecret": ncp.SecretKey,
		}, nil
	case "gcp":
		var gcp models.GCPCredentials
		if err := json.Unmarshal(req.CredentialJson, &gcp); err != nil {
			return nil, fmt.Errorf("invalid gcp credential json: %w", err)
		}

		return map[string]string{
			// "client_id":      gcp.ClientID,
			"ClientEmail": gcp.ClientEmail,
			// "private_key_id": gcp.PrivateKeyID,
			"PrivateKey":  gcp.PrivateKey,
			"ProjectID":   gcp.ProjectID,
			"S3AccessKey": req.S3AccessKey,
			"S3SecretKey": req.S3SecretKey,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported cspType: %q", req.CspType)
	}
}

func getPublicKey() (string, string, error) {
	url := "http://localhost:1323/tumblebug/credential/publicKey"
	// url := "http://mc-infra-manager:1323/tumblebug/credential/publicKey"
	method := http.MethodGet

	body, err := utils.RequestTumblebug(url, method, "", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Parse the response to extract public key and token ID
	var res models.PublicKeyResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", "", fmt.Errorf("get public key failed: %w", err)
	}

	return res.PublicKey, res.PublicKeyTokenId, nil
}

func encryptCredentialsWithPublicKey(publicKeyPem string, credentials map[string]string) ([]map[string]string, string, error) {
	// PEM → rsa.PublicKey 변환
	block, _ := pem.Decode([]byte(strings.ReplaceAll(publicKeyPem, `\n`, "\n")))
	fmt.Println("block: ", block)
	if block == nil {
		return nil, "", fmt.Errorf("invalid public key PEM")
	}

	rsaPublicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse public key: %w", err)
	}

	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, "", fmt.Errorf("encryptCredentialsWithPublicKey failed: %w", err)
	}

	encryptedList := []map[string]string{}

	// 각 credential 값 암호화
	for k, v := range credentials {
		encryptedCredentials := make(map[string]string)
		aesCipher, err := aes.NewCipher(aesKey)
		if err != nil {
			return nil, "", fmt.Errorf("encryptCredentialsWithPublicKey failed: %w", err)
		}

		// IV (CBC에서는 반드시 16바이트)
		iv := make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, "", fmt.Errorf("failed to generate IV: %w", err)
		}

		// CBC 모드 암호화기 생성
		mode := cipher.NewCBCEncrypter(aesCipher, iv)

		// PKCS7 padding 적용
		padded := pkcs7Pad([]byte(v), aes.BlockSize)
		ciphertext := make([]byte, len(padded))

		mode.CryptBlocks(ciphertext, padded)

		// IV + Ciphertext 결합 후 Base64 인코딩
		finalCipher := append(iv, ciphertext...)

		encryptedCredentials["key"] = k
		encryptedCredentials["value"] = base64.StdEncoding.EncodeToString(finalCipher)

		encryptedList = append(encryptedList, encryptedCredentials)
	}

	encryptedAesKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPublicKey, aesKey, nil)
	if err != nil {
		return nil, "", fmt.Errorf("encryptCredentialsWithPublicKey failed: %w", err)
	}
	return encryptedList, base64.StdEncoding.EncodeToString(encryptedAesKey), nil
}

// PKCS7 패딩 추가
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func sendCredentials(payload map[string]interface{}) error {
	url := "http://localhost:1323/tumblebug/credential"
	// url := "http://mc-infra-manager:1323/tumblebug/credential"
	method := http.MethodPost
	reqBody, _ := json.Marshal(payload)

	_, err := utils.RequestTumblebug(url, method, "", reqBody)
	if err != nil {
		return fmt.Errorf("create credential failed: %w", err)
	}

	return nil
}
