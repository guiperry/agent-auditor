package main

import (
	"encoding/json"
	"fmt"
	
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	
)

// VoiceInferenceConfig holds configuration for the voice inference system
type VoiceInferenceConfig struct {
	Enabled   bool   `json:"enabled"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	WSURL     string `json:"ws_url"`
	OutputDir string `json:"output_dir"`
}

// VoiceInferenceManager manages voice report generation
type VoiceInferenceManager struct {
	config     VoiceInferenceConfig
	reportLock sync.Mutex
	audioCache map[string]string // Maps report hash to audio file path
}

// NewVoiceInferenceManager creates a new voice inference manager
func NewVoiceInferenceManager(configPath string) (*VoiceInferenceManager, error) {
	// Default configuration
	config := VoiceInferenceConfig{
		Enabled:   false,
		OutputDir: "voice_reports",
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

	return &VoiceInferenceManager{
		config:     config,
		audioCache: make(map[string]string),
	}, nil
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
	// Prepare the command
	cmd := exec.Command(
		"python3",
		"voice_inference.py",
		"--report", reportPath,
		"--output", v.config.OutputDir,
		"--api-key", v.config.APIKey,
		"--api-secret", v.config.APISecret,
		"--ws-url", v.config.WSURL,
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("voice inference script failed: %v, output: %s", err, output)
	}

	// Parse the output to get the audio file path
	// This is a simplified approach - in a real implementation, you might want to parse JSON output
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