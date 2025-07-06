package key_manager

import (
	"flag"
	"fmt"
	"os"
)

// testKeyManager is a utility function to test the key manager
//
//nolint:unused // This function is used for manual testing
//lint:ignore U1000 This function is used for manual testing
func testKeyManager() {
	// Define command-line flags
	keyFilePath := flag.String("key-file", "default.key", "Path to key file")
	keyName := flag.String("key-name", "", "Name of the key to retrieve")
	listKeys := flag.Bool("list", false, "List all available keys")
	flag.Parse()

	// Get passphrase from environment variable
	passphrase := os.Getenv("AEGONG_KEY_PASS")
	if passphrase == "" {
		fmt.Fprintf(os.Stderr, "Error: AEGONG_KEY_PASS environment variable not set\n")
		os.Exit(1)
	}

	// Create key manager
	keyManager := NewKeyManager(*keyFilePath)
	if err := keyManager.Initialize(passphrase); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing key manager: %v\n", err)
		os.Exit(1)
	}

	// Load keys
	if err := keyManager.LoadKeys(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading keys: %v\n", err)
		os.Exit(1)
	}

	// List all keys if requested
	if *listKeys {
		keys := keyManager.GetAllKeys()
		fmt.Println("Available keys:")
		for _, key := range keys {
			fmt.Printf("- %s\n", key)
		}
		return
	}

	// Get specific key if requested
	if *keyName != "" {
		value, err := keyManager.GetKey(*keyName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving key '%s': %v\n", *keyName, err)
			os.Exit(1)
		}
		fmt.Printf("Key '%s': %s\n", *keyName, value)
		return
	}

	// If no specific action was requested, show usage
	fmt.Println("Please specify either --list or --key-name")
}
