package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloud-barista/mc-data-manager/models"
	service "github.com/cloud-barista/mc-data-manager/service/credential"
)

type ProfileManager interface {
	LoadAllProfiles() (map[string]models.ProfileCredentials, error)
	SaveAllProfiles(profiles map[string]models.ProfileCredentials) error
	CreateProfile(profileName string, credentials models.ProfileCredentials) error
	UpdateProfile(profileName string, credentials models.ProfileCredentials) error
	DeleteProfile(profileName string) error
	LoadCredentialsByProfile(profileName string, provider string) (interface{}, error)
}

type FileProfileManager struct {
	credentialService *service.CredentialService
	profileFilePath   string
	mu                sync.Mutex
}

// todo : profileopath -> credentialService 으로 변경 필요
func NewProfileManager(profileFilePath ...string) *FileProfileManager {
	var path string
	if len(profileFilePath) > 0 && profileFilePath[0] != "" {
		path = profileFilePath[0]
	} else {
		path = filepath.Join(".", "data", "var", "run", "data-manager", "profile", "profile.json")
	}
	return &FileProfileManager{profileFilePath: path}
}

func NewProfileManagerDefault() *FileProfileManager {
	defaultPath := filepath.Join(".", "data", "var", "run", "data-manager", "profile", "profile.json")
	return &FileProfileManager{profileFilePath: defaultPath}
}

// R  profiles
func (fpm *FileProfileManager) LoadAllProfiles() (map[string]models.ProfileCredentials, error) {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()

	file, err := os.Open(fpm.profileFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var profiles []struct {
		ProfileName string                    `json:"profileName"`
		Credentials models.ProfileCredentials `json:"credentials"`
	}

	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}

	profileMap := make(map[string]models.ProfileCredentials)
	for _, profile := range profiles {
		profileMap[profile.ProfileName] = profile.Credentials
	}

	return profileMap, nil
}

// Save File with profiles
func (fpm *FileProfileManager) SaveAllProfiles(profiles map[string]models.ProfileCredentials) error {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()

	file, err := os.Create(fpm.profileFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var profilesList []struct {
		ProfileName string                    `json:"profileName"`
		Credentials models.ProfileCredentials `json:"credentials"`
	}

	for name, creds := range profiles {
		profilesList = append(profilesList, struct {
			ProfileName string                    `json:"profileName"`
			Credentials models.ProfileCredentials `json:"credentials"`
		}{
			ProfileName: name,
			Credentials: creds,
		})
	}

	data, err := json.MarshalIndent(profilesList, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// C profile by name
func (fpm *FileProfileManager) CreateProfile(profileName string, credentials models.ProfileCredentials) error {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()
	profiles, err := fpm.LoadAllProfiles()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; exists {
		return errors.New("profile already exists")
	}

	profiles[profileName] = credentials
	return fpm.SaveAllProfiles(profiles)
}

// U profile by name
func (fpm *FileProfileManager) UpdateProfile(profileName string, credentials models.ProfileCredentials) error {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()
	profiles, err := fpm.LoadAllProfiles()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; !exists {
		return errors.New("profile not found")
	}

	profiles[profileName] = credentials
	return fpm.SaveAllProfiles(profiles)
}

// D profile by name
func (fpm *FileProfileManager) DeleteProfile(profileName string) error {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()
	profiles, err := fpm.LoadAllProfiles()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; !exists {
		return errors.New("profile not found")
	}

	delete(profiles, profileName)
	return fpm.SaveAllProfiles(profiles)
}

// R profile by name
func (fpm *FileProfileManager) LoadCredentialsByProfile(profileName string, provider string) (interface{}, error) {
	profiles, err := fpm.LoadAllProfiles()
	if err != nil {
		return nil, err
	}
	credentials, exists := profiles[profileName]
	if !exists {
		return nil, errors.New("profile not found")
	}
	switch provider {
	case "aws":
		return credentials.AWS, nil
	case "ncp":
		return credentials.NCP, nil
	case "gcp":
		return credentials.GCP, nil
	default:
		return nil, errors.New("unsupported provider")
	}
}

func (fpm *FileProfileManager) LoadCredentialsById(credentialId uint64, provider string) (interface{}, error) {

	// credentialService := controllers.NewCredentialHandler()

	credential, err := fpm.credentialService.GetCredentialById(credentialId)

	if err != nil {
		return nil, fmt.Errorf("credential info not found")
	}

	decryptedJson, err := c.aesConverter.DecryptAESGCM(credential.CredentialJson)
	if err != nil {
		return nil, err
	}

	var creds interface{}
	if err := json.Unmarshal([]byte(decryptedJson), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credential json: %w", err)
	}

	switch provider {
	case "aws":
		return creds.(map[string]interface{})["AWS"], nil
	case "ncp":
		return creds.(map[string]interface{})["NCP"], nil
	case "gcp":
		return creds.(map[string]interface{})["GCP"], nil
	default:
		return nil, errors.New("unsupported provider")
	}
}

// ValidateProfiles checks that at least one profile exists, each profile has at least one credential,
// and that all provided credentials have non-empty required fields.
// func (fpm *FileProfileManager) ValidateProfiles() error {
// 	fpm.mu.Lock()
// 	defer fpm.mu.Unlock()

// 	// Open the profile file
// 	file, err := os.Open(fpm.profileFilePath)
// 	if err != nil {
// 		return fmt.Errorf("unable to open profile file: %v", err)
// 	}
// 	defer file.Close()

// 	// Read the file content
// 	data, err := io.ReadAll(file)
// 	if err != nil {
// 		return fmt.Errorf("unable to read profile file: %v", err)
// 	}

// 	// Unmarshal JSON data into profiles
// 	var profiles []struct {
// 		ProfileName string                    `json:"profileName"`
// 		Credentials models.ProfileCredentials `json:"credentials"`
// 	}

// 	if err := json.Unmarshal(data, &profiles); err != nil {
// 		return fmt.Errorf("unable to parse profile JSON: %v", err)
// 	}

// 	// Check if there are any profiles
// 	if len(profiles) == 0 {
// 		return errors.New("no profiles found")
// 	}

// 	// Validate each profile's credentials
// 	for _, profile := range profiles {
// 		if profile.ProfileName == "" {
// 			return errors.New("a profile has an empty name")
// 		}

// 		creds := profile.Credentials

// 		// Flag to check if at least one credential is present
// 		hasAtLeastOneCredential := false

// 		// Validate AWS credentials if present
// 		if creds.AWS.AccessKey != "" || creds.AWS.SecretKey != "" {
// 			hasAtLeastOneCredential = true
// 			if creds.AWS.AccessKey == "" {
// 				return fmt.Errorf("AWS AccessKey for profile '%s' is missing", profile.ProfileName)
// 			}
// 			if creds.AWS.SecretKey == "" {
// 				return fmt.Errorf("AWS SecretKey for profile '%s' is missing", profile.ProfileName)
// 			}
// 		}

// 		// Validate NCP credentials if present
// 		if creds.NCP.AccessKey != "" || creds.NCP.SecretKey != "" {
// 			hasAtLeastOneCredential = true
// 			if creds.NCP.AccessKey == "" {
// 				return fmt.Errorf("NCP AccessKey for profile '%s' is missing", profile.ProfileName)
// 			}
// 			if creds.NCP.SecretKey == "" {
// 				return fmt.Errorf("NCP SecretKey for profile '%s' is missing", profile.ProfileName)
// 			}
// 		}

// 		// Validate GCP credentials if present
// 		if creds.GCP.PrivateKeyID != "" {
// 			hasAtLeastOneCredential = true
// 			if creds.GCP.PrivateKeyID == "" {
// 				return fmt.Errorf("GCP PrivateKeyID for profile '%s' is missing", profile.ProfileName)
// 			}
// 		}

// 		// Ensure that at least one credential is present
// 		if !hasAtLeastOneCredential {
// 			return fmt.Errorf("profile '%s' must have at least one set of credentials (AWS, NCP, or GCP)", profile.ProfileName)
// 		}
// 	}

// 	return nil
// }
