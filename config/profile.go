package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloud-barista/mc-data-manager/models"
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
	profileFilePath string
	mu              sync.Mutex
}

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
