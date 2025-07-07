package key_manager

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateKeyFile is a utility function to create an encrypted key file
//
// This function is exported for use by the CLI tool
func GenerateKeyFile() {
	// Define command-line flags
	primaryKeyPath := flag.String("output", "ansible/roles/agent_auditor/files/default.key", "Path for the primary (Ansible) key file.")
	envFilePath := flag.String("env", ".env", "Path to the .env file containing API keys.")
	flag.Parse()

	// Get absolute path to the .env file
	absEnvPath, err := filepath.Abs(*envFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving .env file path: %v\n", err)
		os.Exit(1)
	}

	// Parse the .env file to extract API keys
	fmt.Printf("Reading API keys from %s...\n", absEnvPath)
	keys, err := parseEnvFile(absEnvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing .env file: %v\n", err)
		os.Exit(1)
	}

	// Check if we have the passphrase in the .env file
	passphrase, ok := keys["AEGONG_KEY_PASS"]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: AEGONG_KEY_PASS not found in .env file\n")
		os.Exit(1)
	}

	// Remove the passphrase from the keys map so it's not included in the encrypted file
	delete(keys, "AEGONG_KEY_PASS")

	if len(keys) == 0 {
		fmt.Println("No API keys found in .env file (excluding AEGONG_KEY_PASS), exiting")
		os.Exit(0)
	}

	fmt.Printf("Found %d API keys in .env file: %s\n", len(keys), strings.Join(getMapKeys(keys), ", "))

	// Create the encrypted key file
	if err := CreateKeyFile(*primaryKeyPath, passphrase, keys); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating key file at %s: %v\n", *primaryKeyPath, err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Successfully created encrypted key file:\n")
	fmt.Printf("   - For Ansible deployments: %s\n", *primaryKeyPath)
	fmt.Printf("\n   Contains %d keys from .env file: %s\n", len(keys), strings.Join(getMapKeys(keys), ", "))
	fmt.Println("\nThe key file is now ready for deployment.")
}

// getMapKeys extracts keys from a map and returns them as a slice
//
//nolint:unused // This function is used by the generateKeyFile function
//lint:ignore U1000 This function is used by the generateKeyFile function
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// parseEnvFile reads the .env file and extracts API keys
func parseEnvFile(envFilePath string) (map[string]string, error) {
	// Open the .env file
	file, err := os.Open(envFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening .env file: %v", err)
	}
	defer file.Close()

	keys := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse key=value format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip lines that don't have the key=value format
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}

		keys[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %v", err)
	}

	return keys, nil
}

// This function can be called from cmd/generate_keys/main.go
// Example: go run cmd/generate_keys/main.go -output ansible/roles/agent_auditor/files/default.key -env .env
// Note: The .env file must contain AEGONG_KEY_PASS for encryption
