package key_manager

import (
	"os"
	"path/filepath"
	"testing"
)

// TestKeyManagerInitialization tests the initialization of the key manager
func TestKeyManagerInitialization(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "key-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test key file path
	keyFilePath := filepath.Join(tempDir, "test.key")

	// Create a key manager
	keyManager := NewKeyManager(keyFilePath)

	// Check that the key manager was initialized correctly
	if keyManager == nil {
		t.Fatal("Key manager should not be nil")
	}

	if keyManager.keyFilePath != keyFilePath {
		t.Fatalf("Key file path mismatch: expected %s, got %s", keyFilePath, keyManager.keyFilePath)
	}
}

// TestKeyManagerCreateAndLoad tests creating and loading a key file
func TestKeyManagerCreateAndLoad(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "key-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test key file path
	keyFilePath := filepath.Join(tempDir, "test.key")

	// Create test keys
	testKeys := map[string]string{
		"test-key-1": "test-value-1",
		"test-key-2": "test-value-2",
	}

	// Create a test passphrase
	testPassphrase := "test-passphrase"

	// Create the key file
	err = CreateKeyFile(keyFilePath, testPassphrase, testKeys)
	if err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}

	// Check that the key file was created
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		t.Fatal("Key file should exist")
	}

	// Create a key manager
	keyManager := NewKeyManager(keyFilePath)

	// Initialize the key manager
	err = keyManager.Initialize(testPassphrase)
	if err != nil {
		t.Fatalf("Failed to initialize key manager: %v", err)
	}

	// Load the keys
	err = keyManager.LoadKeys()
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	// Check that the keys were loaded correctly
	for key, expectedValue := range testKeys {
		value, err := keyManager.GetKey(key)
		if err != nil {
			t.Fatalf("Failed to get key %s: %v", key, err)
		}

		if value != expectedValue {
			t.Fatalf("Key value mismatch for %s: expected %s, got %s", key, expectedValue, value)
		}
	}

	// Check that getting a non-existent key returns an error
	_, err = keyManager.GetKey("non-existent-key")
	if err == nil {
		t.Fatal("Getting a non-existent key should return an error")
	}

	// Check that GetAllKeys returns all keys
	allKeys := keyManager.GetAllKeys()
	if len(allKeys) != len(testKeys) {
		t.Fatalf("Expected %d keys, got %d", len(testKeys), len(allKeys))
	}

	// Check that all keys are in the result
	for _, key := range allKeys {
		if _, exists := testKeys[key]; !exists {
			t.Fatalf("Unexpected key in GetAllKeys result: %s", key)
		}
	}
}

// TestKeyManagerInvalidPassphrase tests loading keys with an invalid passphrase
func TestKeyManagerInvalidPassphrase(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "key-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test key file path
	keyFilePath := filepath.Join(tempDir, "test.key")

	// Create test keys
	testKeys := map[string]string{
		"test-key": "test-value",
	}

	// Create a test passphrase
	testPassphrase := "test-passphrase"

	// Create the key file
	err = CreateKeyFile(keyFilePath, testPassphrase, testKeys)
	if err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}

	// Create a key manager
	keyManager := NewKeyManager(keyFilePath)

	// Initialize the key manager with an invalid passphrase
	err = keyManager.Initialize("invalid-passphrase")
	if err != nil {
		t.Fatalf("Initializing with an invalid passphrase should not return an error: %v", err)
	}

	// Try to load keys with the invalid passphrase - this should fail
	err = keyManager.LoadKeys()
	if err == nil {
		t.Fatal("Loading keys with an invalid passphrase should return an error")
	}
}

// TestKeyManagerNonExistentFile tests loading keys from a non-existent file
func TestKeyManagerNonExistentFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "key-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test key file path for a non-existent file
	keyFilePath := filepath.Join(tempDir, "non-existent.key")

	// Create a key manager
	keyManager := NewKeyManager(keyFilePath)

	// Initialize the key manager
	err = keyManager.Initialize("test-passphrase")
	if err != nil {
		t.Fatalf("Initializing with a non-existent file should not return an error: %v", err)
	}

	// Try to load keys from the non-existent file - this should fail
	err = keyManager.LoadKeys()
	if err == nil {
		t.Fatal("Loading keys from a non-existent file should return an error")
	}
}

// TestKeyManagerEmptyFile tests loading keys from an empty file
func TestKeyManagerEmptyFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "key-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test key file path
	keyFilePath := filepath.Join(tempDir, "empty.key")

	// Create an empty file
	file, err := os.Create(keyFilePath)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	file.Close()

	// Create a key manager
	keyManager := NewKeyManager(keyFilePath)

	// Initialize the key manager
	err = keyManager.Initialize("test-passphrase")
	if err != nil {
		t.Fatalf("Initializing with an empty file should not return an error: %v", err)
	}

	// Try to load keys from the empty file - this should fail
	err = keyManager.LoadKeys()
	if err == nil {
		t.Fatal("Loading keys from an empty file should return an error")
	}
}
