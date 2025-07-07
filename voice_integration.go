package main

import (
	keys "Agent_Auditor/key_manager"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// VoiceInferenceConfig holds configuration for the voice inference system
type VoiceInferenceConfig struct {
	Enabled      bool   `json:"enabled"`
	Provider     string `json:"provider"`
	KeyFile      string `json:"key_file"`
	KeyPassEnv   string `json:"key_pass_env"`
	OutputDir    string `json:"output_dir"`
	DefaultVoice string `json:"default_voice"`
	DefaultModel string `json:"default_model"`
	WSURL        string `json:"ws_url"` // WebSocket URL for LiveKit
}

// VoiceInferenceManager manages voice report generation
type VoiceInferenceManager struct {
	config     VoiceInferenceConfig
	reportLock sync.Mutex
	audioCache map[string]string // Maps report hash to audio file path
	keyManager *keys.KeyManager  // Secure key manager
}

// NewVoiceInferenceManager creates a new voice inference manager
func NewVoiceInferenceManager(configPath string) (*VoiceInferenceManager, error) {
	// Default configuration
	config := VoiceInferenceConfig{
		Enabled:      false,
		Provider:     "openai",
		KeyFile:      "default.key",
		KeyPassEnv:   "AEGONG_KEY_PASS",
		OutputDir:    "voice_reports",
		DefaultVoice: "alloy",
		DefaultModel: "gpt-4o-mini-tts",
		// WSURL is not used directly as a command-line parameter
		// It's set in the configuration for reference only
		WSURL: "",
	}

	// Try to load configuration from file
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			if err := json.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse voice config: %v", err)
			}
		}
	}

	// Create output directory if it doesn't exist
	if config.Enabled {
		if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create voice reports directory: %v", err)
		}
	}

	// Create voice inference manager
	vim := &VoiceInferenceManager{
		config:     config,
		audioCache: make(map[string]string),
	}

	// Initialize key manager if enabled
	if config.Enabled {
		// Check if we're in development mode (using .env file)
		// In development mode, we'll use environment variables directly
		cerebrasKey := os.Getenv("CEREBRAS_API_KEY")
		cartesiaKey := os.Getenv("CARTESIA_API_KEY")
		livekitKey := os.Getenv("LIVEKIT_API_KEY")
		livekitSecret := os.Getenv("LIVEKIT_API_SECRET")

		// If we have keys in environment variables, create an in-memory key manager
		if cerebrasKey != "" || cartesiaKey != "" || livekitKey != "" || livekitSecret != "" {
			log.Printf("Using API keys from environment variables (development mode)")

			// Create a map of keys
			keyMap := make(map[string]string)
			if cerebrasKey != "" {
				keyMap["cerebras"] = cerebrasKey
			}
			if cartesiaKey != "" {
				keyMap["cartesia"] = cartesiaKey
			}
			if livekitKey != "" {
				keyMap["LIVEKIT_API_KEY"] = livekitKey
			}
			if livekitSecret != "" {
				keyMap["LIVEKIT_API_SECRET"] = livekitSecret
			}

			// Create a temporary key file
			tempKeyFile := "temp_keys.json"
			if err := keys.CreateKeyFile(tempKeyFile, "dummy", keyMap); err != nil {
				log.Printf("Warning: Failed to create temporary key file: %v", err)
			} else {
				// Use the temporary key file
				vim.keyManager = keys.NewKeyManager(tempKeyFile)
				vim.keyManager.Initialize("dummy")
				if err := vim.keyManager.LoadKeys(); err != nil {
					log.Printf("Warning: Failed to load API keys from temporary file: %v", err)
				} else {
					log.Printf("Successfully loaded API keys from environment variables")
				}
				// Clean up the temporary file
				os.Remove(tempKeyFile)
			}
		} else if config.KeyFile != "" {
			// Try to use the encrypted key file (production mode)
			vim.keyManager = keys.NewKeyManager(config.KeyFile)

			// Try to initialize with passphrase from environment variable
			if passphrase := os.Getenv(config.KeyPassEnv); passphrase != "" {
				if err := vim.keyManager.Initialize(passphrase); err != nil {
					log.Printf("Warning: Failed to initialize key manager: %v", err)
				} else {
					// Load keys
					if err := vim.keyManager.LoadKeys(); err != nil {
						log.Printf("Warning: Failed to load API keys: %v", err)
					} else {
						log.Printf("Successfully loaded API keys from %s", config.KeyFile)
					}
				}
			} else {
				log.Printf("Warning: Environment variable %s not set, API keys will not be available", config.KeyPassEnv)
			}
		} else {
			log.Printf("Warning: No API keys available, voice inference will not work")
		}
	}

	return vim, nil
}

// GenerateVoiceReport generates a voice report for the given audit report
func (v *VoiceInferenceManager) GenerateVoiceReport(reportPath string) (string, error) {
	if !v.config.Enabled {
		return "", fmt.Errorf("voice inference is disabled")
	}

	v.reportLock.Lock()
	defer v.reportLock.Unlock()

	// Extract report hash from filename
	reportHash := filepath.Base(reportPath)
	reportHash = reportHash[7:15] // Extract hash from "report_XXXXXXXX.json"

	// Check if we already have an audio file for this report
	if audioPath, exists := v.audioCache[reportHash]; exists {
		// Check if the file exists
		if _, err := os.Stat(audioPath); err == nil {
			return audioPath, nil
		}
	}

	// Generate a new voice report
	audioPath, err := v.runVoiceInference(reportPath)
	if err != nil {
		return "", fmt.Errorf("voice inference failed: %v", err)
	}

	// Cache the result
	v.audioCache[reportHash] = audioPath
	return audioPath, nil
}

// runVoiceInference runs the Python voice inference script
func (v *VoiceInferenceManager) runVoiceInference(reportPath string) (string, error) {
	// Check if key manager is initialized
	if v.keyManager == nil {
		return "", fmt.Errorf("key manager not initialized, cannot access API keys")
	}

	// Base command with common arguments
	args := []string{
		"voice_inference.py",
		"--report", reportPath,
		"--output", v.config.OutputDir,
		"--provider", v.config.Provider,
	}

	// Add voice if specified
	if v.config.DefaultVoice != "" {
		args = append(args, "--voice", v.config.DefaultVoice)
	}

	// Add model if specified
	if v.config.DefaultModel != "" {
		args = append(args, "--model", v.config.DefaultModel)
	}

	// Add timeout parameter to prevent hanging
	args = append(args, "--timeout", "60")

	// Note: WebSocket URL is handled by the LiveKit environment variables
	// and doesn't need to be passed as a command-line argument

	// Add provider-specific API keys
	switch v.config.Provider {
	case "openai":
		// Get OpenAI API key
		apiKey, err := v.keyManager.GetKey("openai")
		if err != nil {
			return "", fmt.Errorf("failed to get OpenAI API key: %v", err)
		}
		args = append(args, "--openai-api-key", apiKey)

	case "cerebras":
		// Get Cerebras API key
		cerebrasKey, err := v.keyManager.GetKey("cerebras")
		if err != nil {
			return "", fmt.Errorf("failed to get Cerebras API key: %v", err)
		}
		args = append(args, "--cerebras-api-key", cerebrasKey)

		// Get Google credentials path (for Cerebras hybrid approach)
		googleCreds, err := v.keyManager.GetKey("google_credentials_path")
		if err != nil {
			return "", fmt.Errorf("failed to get Google credentials path: %v", err)
		}
		args = append(args, "--google-credentials", googleCreds)

	case "google":
		// Get Google credentials path
		googleCreds, err := v.keyManager.GetKey("google_credentials_path")
		if err != nil {
			return "", fmt.Errorf("failed to get Google credentials path: %v", err)
		}
		args = append(args, "--google-credentials", googleCreds)

	case "azure":
		// Get Azure API key
		azureKey, err := v.keyManager.GetKey("azure")
		if err != nil {
			return "", fmt.Errorf("failed to get Azure API key: %v", err)
		}
		args = append(args, "--azure-api-key", azureKey)

		// Get Azure region if available
		if azureRegion, err := v.keyManager.GetKey("azure_region"); err == nil {
			args = append(args, "--azure-region", azureRegion)
		}

	case "cartesia":
		// Get Cartesia API key
		cartesiaKey, err := v.keyManager.GetKey("cartesia")
		if err != nil {
			return "", fmt.Errorf("failed to get Cartesia API key: %v", err)
		}
		args = append(args, "--cartesia-api-key", cartesiaKey)

	case "livekit":
		// Get LiveKit API key
		livekitKey, err := v.keyManager.GetKey("LIVEKIT_API_KEY")
		if err != nil {
			return "", fmt.Errorf("failed to get LiveKit API key: %v", err)
		}
		args = append(args, "--livekit-api-key", livekitKey)

		// Get LiveKit API secret
		livekitSecret, err := v.keyManager.GetKey("LIVEKIT_API_SECRET")
		if err != nil {
			return "", fmt.Errorf("failed to get LiveKit API secret: %v", err)
		}
		args = append(args, "--livekit-api-secret", livekitSecret)

	default:
		return "", fmt.Errorf("unsupported TTS provider: %s", v.config.Provider)
	}

	// Prepare the command
	cmd := exec.Command("python3", args...)

	// Log the command being executed
	log.Printf("Running voice inference command: python3 %s", strings.Join(args, " "))

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Voice inference script failed with error: %v", err)
		log.Printf("Script output: %s", string(output))
		return "", fmt.Errorf("voice inference script failed: %v, output: %s", err, output)
	}

	// Log the output
	log.Printf("Voice inference script output: %s", string(output))

	// Parse the output to get the audio file path
	outputStr := string(output)
	var audioPath string
	_, err = fmt.Sscanf(outputStr, "Voice report generated: %s", &audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse voice inference output: %v", err)
	}

	return audioPath, nil
}

// IsEnabled returns whether voice inference is enabled
func (v *VoiceInferenceManager) IsEnabled() bool {
	return v.config.Enabled
}

// GetAudioPathForReport returns the cached audio path for a report hash, if available
func (v *VoiceInferenceManager) GetAudioPathForReport(reportHash string) (string, bool) {
	v.reportLock.Lock()
	defer v.reportLock.Unlock()

	path, exists := v.audioCache[reportHash]
	return path, exists
}

// GenerateVoiceReportAsync generates a voice report asynchronously
func (v *VoiceInferenceManager) GenerateVoiceReportAsync(reportPath string, callback func(string, error)) {
	if !v.config.Enabled {
		if callback != nil {
			callback("", fmt.Errorf("voice inference is disabled"))
		}
		return
	}

	go func() {
		audioPath, err := v.GenerateVoiceReport(reportPath)
		if callback != nil {
			callback(audioPath, err)
		}
	}()
}
