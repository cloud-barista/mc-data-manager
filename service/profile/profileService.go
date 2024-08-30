package auth

import (
	"log"
	"sync"
)

type ProfileService struct {
	mu sync.Mutex
}

// NewProfileService creates a new ProfileService
func NewProfileService() *ProfileService {
	return &ProfileService{}
}

// CreateProfile adds a new profile
func (ps *ProfileService) CreateProfile(profileName string, credentials Credentials) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return CreateProfile(profileName, credentials)
}

// UpdateProfile updates an existing profile
func (ps *ProfileService) UpdateProfile(profileName string, credentials Credentials) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return UpdateProfile(profileName, credentials)
}

// DeleteProfile removes a profile
func (ps *ProfileService) DeleteProfile(profileName string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return DeleteProfile(profileName)
}

// LoadCredentialsByProfile loads credentials by profile name and provider
func (ps *ProfileService) LoadCredentialsByProfile(profileName string, provider string) (interface{}, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return LoadCredentialsByProfile(profileName, provider)
}

// LoadAllProfiles loads all profiles
func (ps *ProfileService) LoadAllProfiles() (map[string]Credentials, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return LoadAllCredentials()
}

// Example usage of ProfileService with WaitGroup
func ExampleUsage() {
	service := NewProfileService()
	var wg sync.WaitGroup

	// Create a new profile
	wg.Add(1)
	go func() {
		defer wg.Done()
		creds := Credentials{ /* ... initialize credentials ... */ }
		if err := service.CreateProfile("exampleProfile", creds); err != nil {
			log.Println("Error creating profile:", err)
		}
	}()

	// Update the profile
	wg.Add(1)
	go func() {
		defer wg.Done()
		updatedCreds := Credentials{ /* ... initialize updated credentials ... */ }
		if err := service.UpdateProfile("exampleProfile", updatedCreds); err != nil {
			log.Println("Error updating profile:", err)
		}
	}()

	// Delete the profile
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := service.DeleteProfile("exampleProfile"); err != nil {
			log.Println("Error deleting profile:", err)
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("All operations completed.")
}
