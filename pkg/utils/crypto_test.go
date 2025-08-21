package utils

import (
	"fmt"
	"testing"
)

func TestEncryptDecryptAESGCM(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes = AES-256
	original := `{
  		"access_key": "AKIAxxxxxxxxx",
  		"secret_key": "xxxxxxxxxxxxxxxxxxx"
		}`

	// 암호화
	encrypted, err := EncryptAESGCM(key, original)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	fmt.Println("Encrypted:", encrypted)

	// 복호화
	decrypted, err := DecryptAESGCM(key, encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	fmt.Println("Decrypted:", decrypted)

	// 검증
	if decrypted != original {
		t.Errorf("expected %s, got %s", original, decrypted)
	}
}
