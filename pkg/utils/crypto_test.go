package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestEncryptDecryptAESGCM(t *testing.T) {

	// 기존 환경변수 값 백업
	originalKey := os.Getenv("ENCODING_SECRET_KEY")

	// 테스트 종료 후 환경변수 복원
	defer func() {
		if originalKey == "" {
			os.Unsetenv("ENCODING_SECRET_KEY")
		} else {
			os.Setenv("ENCODING_SECRET_KEY", originalKey)
		}
	}()

	key := "12345678901234567890123456789012" // 32 bytes = AES-256
	os.Setenv("ENCODING_SECRET_KEY", key)
	original := `{
  		"access_key": "AKIAxxxxxxxxx",
  		"secret_key": "xxxxxxxxxxxxxxxxxxx"
		}`

	aesConverter := NewAESConverter()

	// 암호화
	encrypted, err := aesConverter.EncryptAESGCM(original)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	fmt.Println("Encrypted:", encrypted)

	// 복호화
	decrypted, err := aesConverter.DecryptAESGCM(encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	fmt.Println("Decrypted:", decrypted)

	// 검증
	if decrypted != original {
		t.Errorf("expected %s, got %s", original, decrypted)
	}
}
