package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/cloud-barista/mc-data-manager/models"
)

type Credentials = models.ProfileCredentials

var mu sync.Mutex

const (
	// Define the constant path for the credentials file
	CredentialsFilePath = ".dm/var/run/mc-data-manager/profile/auth/auth.json"
)

// CredentialsManager interface definition
type CredentialsManager interface {
	LoadAllCredentials() (map[string]Credentials, error)
	SaveAllCredentials(profiles map[string]Credentials) error
	CreateProfile(profileName string, credentials Credentials) error
	UpdateProfile(profileName string, credentials Credentials) error
	DeleteProfile(profileName string) error
	LoadCredentialsByProfile(profileName string, provider models.Provider) (interface{}, error)
}

// FileCredentialsManager struct definition
type FileCredentialsManager struct {
	authFilePath string
}

// NewFileCredentialsManager constructor function
func NewFileCredentialsManager() *FileCredentialsManager {
	return &FileCredentialsManager{authFilePath: CredentialsFilePath}
}

// LoadAllCredentials loads all credentials from the auth.json file
func (fcm *FileCredentialsManager) LoadAllCredentials() (map[string]Credentials, error) {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(fcm.authFilePath)
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
func (fcm *FileCredentialsManager) SaveAllCredentials(profiles map[string]Credentials) error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Create(fcm.authFilePath)
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
func (fcm *FileCredentialsManager) CreateProfile(profileName string, credentials Credentials) error {
	mu.Lock()
	defer mu.Unlock()
	profiles, err := fcm.LoadAllCredentials()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; exists {
		return errors.New("profile already exists")
	}

	profiles[profileName] = credentials
	return fcm.SaveAllCredentials(profiles)
}

// UpdateProfile updates an existing profile
func (fcm *FileCredentialsManager) UpdateProfile(profileName string, credentials Credentials) error {
	mu.Lock()
	defer mu.Unlock()
	profiles, err := fcm.LoadAllCredentials()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; !exists {
		return errors.New("profile not found")
	}

	profiles[profileName] = credentials
	return fcm.SaveAllCredentials(profiles)
}

// DeleteProfile removes a profile
func (fcm *FileCredentialsManager) DeleteProfile(profileName string) error {
	mu.Lock()
	defer mu.Unlock()
	profiles, err := fcm.LoadAllCredentials()
	if err != nil {
		return err
	}

	if _, exists := profiles[profileName]; !exists {
		return errors.New("profile not found")
	}

	delete(profiles, profileName)
	return fcm.SaveAllCredentials(profiles)
}

// LoadCredentialsByProfile loads credentials by profile name and provider
func (fcm *FileCredentialsManager) LoadCredentialsByProfile(profileName string, provider string) (interface{}, error) {
	profiles, err := fcm.LoadAllCredentials()
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

// AWSClient wraps aws.Config and provides additional methods
type AWSClient struct {
	Config aws.Config
}

// NewAWSClient creates a new AWSClient
func NewAWSClient(profileName string) (*AWSClient, error) {
	fmt.Println("Creating CredentialsManager")
	credentialsManager := NewFileCredentialsManager()

	fmt.Println("Loading credentials")
	creds, err := credentialsManager.LoadCredentialsByProfile(profileName, string(models.AWS))
	if err != nil {
		return nil, fmt.Errorf("Error loading credentials: %v", err)
	}

	fmt.Println("Casting credentials")
	awsCreds, ok := creds.(models.AWSCredentials)
	if !ok {
		return nil, fmt.Errorf("Invalid credentials type: %v", creds)
	}

	fmt.Println("Creating AWS config")
	return newAWSConfigure(awsCreds)
}

// newAWSConfig creates a new AWSClient with the given credentials
func newAWSConfigure(params models.AWSCredentials) (*AWSClient, error) {
	cfg, err := loadConfig(params)
	if err != nil {
		return nil, err
	}
	return &AWSClient{Config: cfg}, nil
}

// loadConfig loads the AWS configuration with the given credentials
func loadConfig(params models.AWSCredentials) (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(params.AccessKey, params.SecretKey, "")),
		config.WithRetryMaxAttempts(5),
	)
}

// SetRegion sets the region in the AWS configuration
func (client *AWSClient) SetRegion(region string) {
	client.Config.Region = region
}

// GetConfig returns the current AWS configuration
func (client *AWSClient) GetConfig() aws.Config {
	return client.Config
}
