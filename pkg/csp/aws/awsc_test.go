package aws

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	cfg "github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	fmt.Println("Starting Test function")

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		logrus.Fatalf("Failed to get current file path")
	}
	rootDir := filepath.Join(filepath.Dir(filename), "../../..")

	profileName := "default"
	provider := "aws"
	defaultRegion := "ap-northeast-2"

	fmt.Println("Creating CredentialsManager")

	profilePath := filepath.Join(rootDir, "data", "var", "run", "data-manager", "profile", "profile.json")
	credentialsManager := cfg.NewProfileManager(profilePath)

	fmt.Println("Loading credentials")
	// Load credentials for the specified profile and provider
	creds, err := credentialsManager.LoadCredentialsByProfile(profileName, provider)
	if err != nil {
		fmt.Println("Error loading credentials:", err)
		return
	}

	fmt.Println("Casting credentials")
	awsCreds, ok := creds.(models.AWSCredentials)
	if !ok {
		fmt.Println(creds)
		fmt.Println("Invalid credentials type")
		return
	}

	fmt.Println("Creating AWS config")
	client, err := newAWSConfig(awsCreds)
	if err != nil {
		fmt.Println("Error creating AWS config:", err)
		return
	}

	regions, err := client.ListRegions()
	if err != nil {
		fmt.Println("Error listing regions:", err)
		return
	}
	fmt.Println("Regions:", regions)

	client.Config.Region = defaultRegion
	fmt.Println("Listing AWS resources")

	// List AWS resources
	listResources(client) // listResources

	fmt.Println("Finished main function")
}
