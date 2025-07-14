package configs

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gorilla/securecookie"
)

type SessionKeys struct {
	AuthKey []byte
	EncKey  []byte
}

func LoadSessionKeysFromEnv() (*SessionKeys, error) {
	env := LoadEnv()

	authKeyBase64 := env.AppAuthKey
	encKeyBase64 := env.AppEncKey

	log.Printf("DEBUG: Raw APP_AUTH_KEY from .env: '%s' (length: %d)", authKeyBase64, len(authKeyBase64))
	log.Printf("DEBUG: Raw APP_ENC_KEY from .env: '%s' (length: %d)", encKeyBase64, len(encKeyBase64))

	if authKeyBase64 == "" {
		return nil, fmt.Errorf("APP_AUTH_KEY environment variable not set")
	}
	if encKeyBase64 == "" {
		return nil, fmt.Errorf("APP_ENC_KEY environment variable not set")
	}

	authKey, err := base64.URLEncoding.DecodeString(authKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode APP_AUTH_KEY from Base64: %w", err)
	}
	encKey, err := base64.URLEncoding.DecodeString(encKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode APP_ENC_KEY from Base64: %w", err)
	}

	if len(encKey) != 16 && len(encKey) != 24 && len(encKey) != 32 {
		return nil, fmt.Errorf("APP_ENC_KEY has invalid length %d after decoding. Must be 16, 24, or 32 bytes for AES encryption", len(encKey))
	}

	log.Println("✅ Session keys loaded and decoded successfully.")
	return &SessionKeys{
		AuthKey: authKey,
		EncKey:  encKey,
	}, nil
}

func GenerateAndPrintSessionKeys() error {
	fmt.Println("Generating new session keys...")

	authKey := securecookie.GenerateRandomKey(64)
	if authKey == nil {
		return fmt.Errorf("error: could not generate authentication key")
	}

	encKey := securecookie.GenerateRandomKey(32)
	if encKey == nil {
		return fmt.Errorf("error: could not generate encryption key")
	}

	authKeyBase64 := base64.URLEncoding.EncodeToString(authKey)
	encKeyBase64 := base64.URLEncoding.EncodeToString(encKey)

	fmt.Println("\n================================================")
	fmt.Println("Generated keys:")
	fmt.Printf("APP_AUTH_KEY=%s\n", authKeyBase64)
	fmt.Printf("APP_ENC_KEY=%s\n", encKeyBase64)
	fmt.Println("================================================")

	envFilePath := ".env.new_keys"
	fullPath, err := filepath.Abs(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", envFilePath, err)
	}

	file, err := os.Create(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", envFilePath, err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "APP_AUTH_KEY=%s\nAPP_ENC_KEY=%s\n", authKeyBase64, encKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to write keys to file %s: %w", envFilePath, err)
	}

	fmt.Printf("\n✅ Keys have been written to '%s'.\n", envFilePath)
	fmt.Println("Please copy these lines from that file into your actual .env file.")
	fmt.Println("REMINDER: Store these keys securely and only generate them ONCE for your production environment.")
	fmt.Println("If you regenerate, existing user sessions will be invalidated.")

	fmt.Printf("fullpath: %s", fullPath)

	return nil
}
