package key_manager

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// generateKeyFile is a utility function to create an encrypted key file
//
//nolint:unused // This function is used by the CLI tool
//lint:ignore U1000 This function is used by the CLI tool
func generateKeyFile() {
	// Define command-line flags
	keyFilePath := flag.String("output", "default.key", "Path to output key file")
	flag.Parse()

	// Get passphrase securely
	fmt.Print("Enter passphrase for encryption: ")
	passphraseBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading passphrase: %v\n", err)
		os.Exit(1)
	}
	passphrase := string(passphraseBytes)
	fmt.Println() // Print newline after password input

	// Confirm passphrase
	fmt.Print("Confirm passphrase: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading passphrase confirmation: %v\n", err)
		os.Exit(1)
	}
	confirm := string(confirmBytes)
	fmt.Println() // Print newline after password input

	if passphrase != confirm {
		fmt.Fprintf(os.Stderr, "Passphrases do not match\n")
		os.Exit(1)
	}

	// Collect API keys
	keys := make(map[string]string)
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\nEnter API keys (leave key name empty to finish):")
	for {
		fmt.Print("Key name (e.g., 'cerebras', 'google'): ")
		scanner.Scan()
		keyName := strings.TrimSpace(scanner.Text())
		if keyName == "" {
			break
		}

		fmt.Printf("Value for '%s': ", keyName)
		scanner.Scan()
		keyValue := strings.TrimSpace(scanner.Text())
		if keyValue == "" {
			fmt.Println("Warning: Empty value provided, skipping this key")
			continue
		}

		keys[keyName] = keyValue
		fmt.Printf("Added key: %s\n", keyName)
	}

	if len(keys) == 0 {
		fmt.Println("No keys provided, exiting")
		os.Exit(0)
	}

	// Create the encrypted key file
	if err := CreateKeyFile(*keyFilePath, passphrase, keys); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating key file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created encrypted key file: %s\n", *keyFilePath)
	fmt.Printf("Contains %d keys: %s\n", len(keys), strings.Join(getMapKeys(keys), ", "))
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

// This function can be called from main.go with a command-line flag
// Example: ./aegong -generate-keys
//
//nolint:unused // This function is used as an entry point for the key generation tool
//lint:ignore U1000 This function is used as an entry point for the key generation tool
func main() {
	generateKeyFile()
}
