package main

import (
	"Agent_Auditor/key_manager"
	"flag"
	"fmt"
	"os"
	
)

func main() {
	// Define command-line flags
	keyFilePath := flag.String("key-file", "default.key", "Path to the encrypted key file")
	keyName := flag.String("key-name", "", "Name of the key to retrieve")
	listKeys := flag.Bool("list", false, "List all available keys")
	passEnv := flag.String("pass-env", "AEGONG_KEY_PASS", "Environment variable containing the passphrase")
	flag.Parse()

	// Get passphrase from environment variable
	passphrase := os.Getenv(*passEnv)
	if passphrase == "" {
		fmt.Fprintf(os.Stderr, "Error: Environment variable %s not set\n", *passEnv)
		os.Exit(1)
	}

	// Create key manager
	km := key_manager.NewKeyManager(*keyFilePath)
	if err := km.Initialize(passphrase); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing key manager: %v\n", err)
		os.Exit(1)
	}

	// Load keys
	if err := km.LoadKeys(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading keys: %v\n", err)
		os.Exit(1)
	}

	// List all keys if requested
	if *listKeys {
		keys := km.GetAllKeys()
		fmt.Printf("Available keys (%d):\n", len(keys))
		for _, k := range keys {
			fmt.Printf("- %s\n", k)
		}
		return
	}

	// Retrieve specific key if requested
	if *keyName != "" {
		value, err := km.GetKey(*keyName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving key '%s': %v\n", *keyName, err)
			os.Exit(1)
		}
		fmt.Printf("Key: %s\nValue: %s\n", *keyName, value)
		return
	}

	// If neither -list nor -key-name was provided, show usage
	fmt.Println("Please specify either -list to show all keys or -key-name to retrieve a specific key")
	fmt.Println("Usage:")
	fmt.Println("  ./test-keys -key-file default.key -list")
	fmt.Println("  ./test-keys -key-file default.key -key-name openai")
}