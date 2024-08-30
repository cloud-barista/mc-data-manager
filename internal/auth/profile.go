package auth

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloud-barista/mc-data-manager/models"
)

type Credentials = models.ProfileCredentials

var mu sync.Mutex

// LoadAllCredentials loads all credentials from the auth.json file
func LoadAllCredentials() (map[string]Credentials, error) {
	mu.Lock()
	defer mu.Unlock()
	authFilePath := filepath.Join("profile", "auth", "auth.json")

	file, err := os.Open(authFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var profiles []struct {
		ProfileName string      `json:"profileName"`
		Credentials Credentials `json:"credentials"`
	}

	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}

	credentialsMap := make(map[string]Credentials)
	for _, profile := range profiles {
		credentialsMap[profile.ProfileName] = profile.Credentials
	}

	return credentialsMap, nil
}

// SaveAllCredentials saves all credentials to the auth.json file
func SaveAllCredentials(profiles map[string]Credentials) error {
	mu.Lock()
	defer mu.Unlock()
	authFilePath := filepath.Join("profile", "auth", "auth.json")

	file, err := os.Create(authFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var profilesList []struct {
		ProfileName string      `json:"profileName"`
		Credentials Credentials `json:"credentials"`
	}

	for name, creds := range profiles {
		profilesList = append(profilesList, struct {
			ProfileName string      `json:"profileName"`
			Credentials Credentials `json:"credentials"`
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

// CreateProfile adds a new profile
func CreateProfile(profileName string, credentials Credentials) error {
	mu.Lock()
	defer mu.Unlock()
	profiles, err := LoadAllCredentials()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; exists {
		return errors.New("profile already exists")
	}

	profiles[profileName] = credentials
	return SaveAllCredentials(profiles)
}

// UpdateProfile updates an existing profile
func UpdateProfile(profileName string, credentials Credentials) error {
	mu.Lock()
	defer mu.Unlock()
	profiles, err := LoadAllCredentials()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; !exists {
		return errors.New("profile not found")
	}

	profiles[profileName] = credentials
	return SaveAllCredentials(profiles)
}

// DeleteProfile removes a profile
func DeleteProfile(profileName string) error {
	mu.Lock()
	defer mu.Unlock()
	profiles, err := LoadAllCredentials()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; !exists {
		return errors.New("profile not found")
	}

	delete(profiles, profileName)
	return SaveAllCredentials(profiles)
}

// LoadCredentialsByProfile loads credentials by profile name and provider
func LoadCredentialsByProfile(profileName string, provider string) (interface{}, error) {
	mu.Lock()
	defer mu.Unlock()
	profiles, err := LoadAllCredentials()
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
