package auth

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/cloud-barista/mc-data-manager/models"
)

// EnsureDirectory ensures that the directory exists
func EnsureDirectory(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// CreateAuthFile creates a sample auth.json file for testing if it doesn't exist
func CreateAuthFile() error {
	authFilePath := "profile/auth/auth.json"
	if _, err := os.Stat(authFilePath); os.IsNotExist(err) {
		// Create sample data
		profiles := []struct {
			ProfileName string                    `json:"profileName"`
			Credentials models.ProfileCredentials `json:"credentials"`
		}{
			{
				ProfileName: "default",
				Credentials: models.ProfileCredentials{
					AWS: models.AWSCredentials{
						AccessKey: "AKIAYDYDQQ5UTODWKD7R",
						SecretKey: "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",
					},
					NCP: models.NCPCredentials{
						AccessKey: "Td5ckFdtaa3qjR82viCW",
						SecretKey: "CAVd0qf7toLhgfHIHbPWBh7FjfswwNUcV3UMsQEC",
					},
					GCP: models.GCPCredentials{
						Type:                "service_account",
						ProjectID:           "spatial-conduit-399006",
						PrivateKeyID:        "d441ca03bc06f32e5420d40a6d1c647b29ee6bde",
						PrivateKey:          "-----BEGIN PRIVATE KEY-----\nprivate-key-content\n-----END PRIVATE KEY-----\n",
						ClientEmail:         "gcp-client-email@example.com",
						ClientID:            "gcp-client-id-1",
						AuthURI:             "https://accounts.google.com/o/oauth2/auth",
						TokenURI:            "https://oauth2.googleapis.com/token",
						AuthProviderCertURL: "https://www.googleapis.com/oauth2/v1/certs",
						ClientCertURL:       "https://www.googleapis.com/robot/v1/metadata/x509/gcp-client-email",
						UniverseDomain:      "example.com",
					},
				},
			},
		}

		// Create the file
		file, err := os.Create(authFilePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Write sample data to the file
		if err := json.NewEncoder(file).Encode(profiles); err != nil {
			return err
		}
	}
	return nil
}

// TestLoadAllCredentials tests the LoadAllCredentials function
func TestLoadAllCredentials(t *testing.T) {
	// Ensure the directory exists
	if err := EnsureDirectory("profile/auth"); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create the auth.json file if it doesn't exist
	if err := CreateAuthFile(); err != nil {
		t.Fatalf("failed to create auth.json: %v", err)
	}

	// Copy the original auth.json to test.json
	originalFilePath := "profile/auth/auth.json"
	testFilePath := "profile/auth/test.json"

	inputFile, err := os.Open(originalFilePath)
	if err != nil {
		t.Fatalf("failed to open original file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(testFilePath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(testFilePath) // Clean up after test
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}

	credentials, err := LoadAllCredentials()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assertions
	if len(credentials) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(credentials))
	}
	if credentials["default"].AWS.AccessKey != "AKIAYDYDQQ5UTODWKD7R" {
		t.Fatalf("expected access key to be 'AKIAYDYDQQ5UTODWKD7R', got %v", credentials["default"].AWS.AccessKey)
	}
}

// TestSaveAllCredentials tests the SaveAllCredentials function
func TestSaveAllCredentials(t *testing.T) {
	// Ensure the directory exists
	if err := EnsureDirectory("profile/auth"); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create the auth.json file if it doesn't exist
	if err := CreateAuthFile(); err != nil {
		t.Fatalf("failed to create auth.json: %v", err)
	}

	profiles := map[string]models.ProfileCredentials{
		"testProfile": {
			AWS: models.AWSCredentials{
				AccessKey: "aws-access-key-1",
				SecretKey: "aws-secret-key-1",
			},
		},
	}

	err := SaveAllCredentials(profiles)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the saved data
	authFilePath := "profile/auth/auth.json"
	file, err := os.Open(authFilePath)
	if err != nil {
		t.Fatalf("failed to open auth.json: %v", err)
	}
	defer file.Close()

	var savedProfiles []struct {
		ProfileName string                    `json:"profileName"`
		Credentials models.ProfileCredentials `json:"credentials"`
	}

	if err := json.NewDecoder(file).Decode(&savedProfiles); err != nil {
		t.Fatalf("failed to decode saved data: %v", err)
	}

	if len(savedProfiles) != 1 || savedProfiles[0].ProfileName != "testProfile" {
		t.Fatalf("expected 1 profile named 'testProfile', got %v", savedProfiles)
	}
}

// TestCreateProfile tests the CreateProfile function
func TestCreateProfile(t *testing.T) {
	// Ensure the directory exists
	if err := EnsureDirectory("profile/auth"); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create the auth.json file if it doesn't exist
	if err := CreateAuthFile(); err != nil {
		t.Fatalf("failed to create auth.json: %v", err)
	}

	creds := models.ProfileCredentials{
		AWS: models.AWSCredentials{
			AccessKey: "aws-access-key-1",
			SecretKey: "aws-secret-key-1",
		},
	}
	err := CreateProfile("newProfile", creds)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the profile was created
	credentials, err := LoadAllCredentials()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, exists := credentials["newProfile"]; !exists {
		t.Fatalf("expected profile 'newProfile' to exist")
	}
}

// TestUpdateProfile tests the UpdateProfile function
func TestUpdateProfile(t *testing.T) {
	// Ensure the directory exists
	if err := EnsureDirectory("profile/auth"); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create the auth.json file if it doesn't exist
	if err := CreateAuthFile(); err != nil {
		t.Fatalf("failed to create auth.json: %v", err)
	}

	creds := models.ProfileCredentials{
		AWS: models.AWSCredentials{
			AccessKey: "updated-aws-access-key",
			SecretKey: "updated-aws-secret-key",
		},
	}
	err := UpdateProfile("newProfile", creds)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the profile was updated
	credentials, err := LoadAllCredentials()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if credentials["newProfile"].AWS.AccessKey != "updated-aws-access-key" {
		t.Fatalf("expected updated-aws-access-key, got %v", credentials["newProfile"].AWS.AccessKey)
	}
}

// TestDeleteProfile tests the DeleteProfile function
func TestDeleteProfile(t *testing.T) {
	// Ensure the directory exists
	if err := EnsureDirectory("profile/auth"); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create the auth.json file if it doesn't exist
	if err := CreateAuthFile(); err != nil {
		t.Fatalf("failed to create auth.json: %v", err)
	}

	err := DeleteProfile("newProfile")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the profile was deleted
	credentials, err := LoadAllCredentials()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, exists := credentials["newProfile"]; exists {
		t.Fatalf("expected profile 'newProfile' to be deleted")
	}
}

// TestLoadCredentialsByProfile tests the LoadCredentialsByProfile function
func TestLoadCredentialsByProfile(t *testing.T) {
	// Ensure the directory exists
	if err := EnsureDirectory("profile/auth"); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create the auth.json file if it doesn't exist
	if err := CreateAuthFile(); err != nil {
		t.Fatalf("failed to create auth.json: %v", err)
	}

	creds := models.ProfileCredentials{
		AWS: models.AWSCredentials{
			AccessKey: "aws-access-key-1",
			SecretKey: "aws-secret-key-1",
		},
	}
	err := CreateProfile("testProfile", creds)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Load credentials by profile name
	loadedCreds, err := LoadCredentialsByProfile("testProfile", "aws")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if loadedCreds != "aws-access-key-1" {
		t.Fatalf("expected aws-access-key-1, got %v", loadedCreds)
	}
}
