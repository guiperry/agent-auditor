package key_manager

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	
	"sync"
)

// APIKeyStore represents the structure of the encrypted key file
type APIKeyStore struct {
	Keys map[string]string `json:"keys"`
}

// KeyManager handles secure loading and decryption of API keys
type KeyManager struct {
	keyFilePath string
	passphrase  string
	keyCache    map[string]string
	mutex       sync.RWMutex
}

// NewKeyManager creates a new key manager instance
func NewKeyManager(keyFilePath string) *KeyManager {
	return &KeyManager{
		keyFilePath: keyFilePath,
		keyCache:    make(map[string]string),
	}
}

// Initialize sets up the key manager with the passphrase
func (km *KeyManager) Initialize(passphrase string) error {
	if passphrase == "" {
		return errors.New("passphrase cannot be empty")
	}
	km.passphrase = passphrase
	return nil
}

// LoadKeys loads and decrypts all keys from the key file
func (km *KeyManager) LoadKeys() error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	// Check if key file exists
	if _, err := os.Stat(km.keyFilePath); os.IsNotExist(err) {
		return fmt.Errorf("key file not found: %s", km.keyFilePath)
	}

	// Read encrypted data
	encryptedData, err := ioutil.ReadFile(km.keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %v", err)
	}

	// Decrypt the data
	decryptedData, err := decrypt(encryptedData, km.passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt key file: %v", err)
	}

	// Parse JSON
	var keyStore APIKeyStore
	if err := json.Unmarshal(decryptedData, &keyStore); err != nil {
		return fmt.Errorf("failed to parse key file: %v", err)
	}

	// Store in cache
	km.keyCache = keyStore.Keys
	return nil
}

// GetKey retrieves a key by name
func (km *KeyManager) GetKey(keyName string) (string, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	// Check if key exists in cache
	if key, exists := km.keyCache[keyName]; exists {
		return key, nil
	}

	return "", fmt.Errorf("key not found: %s", keyName)
}

// GetAllKeys returns all available key names
func (km *KeyManager) GetAllKeys() []string {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	keys := make([]string, 0, len(km.keyCache))
	for k := range km.keyCache {
		keys = append(keys, k)
	}
	return keys
}

// CreateKeyFile creates a new encrypted key file
func CreateKeyFile(keyFilePath, passphrase string, keys map[string]string) error {
	// Create key store
	keyStore := APIKeyStore{
		Keys: keys,
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(keyStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create JSON: %v", err)
	}

	// Encrypt the data
	encryptedData, err := encrypt(jsonData, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %v", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(keyFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write to file
	if err := ioutil.WriteFile(keyFilePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	return nil
}

// Helper functions for encryption/decryption

// createHash creates a SHA-256 hash from a passphrase
func createHash(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// encrypt encrypts data using AES-256-GCM
func encrypt(data []byte, passphrase string) ([]byte, error) {
	key := createHash(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-256-GCM
func decrypt(data []byte, passphrase string) ([]byte, error) {
	key := createHash(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}