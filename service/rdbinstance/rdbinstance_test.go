package rdbinstance

import (
	"testing"

	"github.com/cloud-barista/mc-data-manager/models"
)

func TestProviderFor_AWSReturnsProvider(t *testing.T) {
	creds := models.AWSCredentials{AccessKey: "AKIA_TEST", SecretKey: "secret"}

	p, err := providerFor("aws", creds, "ap-northeast-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestProviderFor_AWSWrongCredsType(t *testing.T) {
	p, err := providerFor("aws", models.NCPCredentials{}, "ap-northeast-2")
	if err == nil {
		t.Fatal("expected error for wrong credential type")
	}
	if p != nil {
		t.Errorf("expected nil provider, got %v", p)
	}
}

func TestProviderFor_AlibabaReturnsProvider(t *testing.T) {
	creds := models.AlibabaCredentials{AccessKey: "ali_test", SecretKey: "secret"}

	p, err := providerFor("alibaba", creds, "ap-southeast-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestProviderFor_AlibabaWrongCredsType(t *testing.T) {
	p, err := providerFor("alibaba", models.AWSCredentials{}, "ap-southeast-1")
	if err == nil {
		t.Fatal("expected error for wrong credential type")
	}
	if p != nil {
		t.Errorf("expected nil provider, got %v", p)
	}
}

func TestProviderFor_NotImplementedProviders(t *testing.T) {
	for _, provider := range []string{"gcp", "ncp"} {
		p, err := providerFor(provider, nil, "region")
		if err == nil {
			t.Errorf("provider %q: expected not-implemented error", provider)
		}
		if p != nil {
			t.Errorf("provider %q: expected nil provider", provider)
		}
	}
}

func TestProviderFor_UnsupportedProvider(t *testing.T) {
	p, err := providerFor("foo", nil, "region")
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if p != nil {
		t.Errorf("expected nil provider, got %v", p)
	}
}
