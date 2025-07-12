package main

import (
	"bytes"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AgentValidationResult represents the result of agent validation
type AgentValidationResult struct {
	IsAgent      bool     `json:"is_agent"`
	Confidence   float64  `json:"confidence"`
	Reasons      []string `json:"reasons"`
	AgentType    string   `json:"agent_type"`
	Capabilities []string `json:"capabilities"`
}

// ValidateAgent checks if a file is an AI agent based on defined criteria
func ValidateAgent(filePath string) (*AgentValidationResult, error) {
	// Read file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Initialize result with default values
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "unknown",
		Capabilities: []string{},
	}

	// Check file type
	fileType := detectFileType(fileData, filePath)

	// Validate based on file type
	switch fileType {
	case "wasm":
		return validateWasmAgent(fileData)
	case "elf":
		return validateElfAgent(fileData)
	case "pe":
		return validatePeAgent(fileData)
	case "macho":
		return validateMachoAgent(fileData)
	case "script":
		return validateScriptAgent(fileData)
	case "jar":
		return validateJarAgent(fileData, filePath)
	case "library":
		return validateLibraryAgent(fileData)
	case "executable":
		// For generic executables, try to determine the actual format
		if len(fileData) >= 4 && bytes.Equal(fileData[0:4], []byte{0x7F, 0x45, 0x4C, 0x46}) {
			return validateElfAgent(fileData)
		} else if len(fileData) >= 2 && bytes.Equal(fileData[0:2], []byte{0x4D, 0x5A}) {
			return validatePeAgent(fileData)
		} else if len(fileData) >= 4 && (binary.LittleEndian.Uint32(fileData[0:4]) == 0xFEEDFACE ||
			binary.LittleEndian.Uint32(fileData[0:4]) == 0xFEEDFACF ||
			binary.LittleEndian.Uint32(fileData[0:4]) == 0xCEFAEDFE ||
			binary.LittleEndian.Uint32(fileData[0:4]) == 0xCFFAEDFE) {
			return validateMachoAgent(fileData)
		} else {
			// If we can't determine the format, use string-based analysis
			stringData := extractStringsFromBinary(fileData)
			return validateBasedOnStringContent(stringData, "executable"), nil
		}
	case "unknown":
		result.Reasons = append(result.Reasons, "Unknown file type")
	}

	return result, nil
}

// detectFileType determines the type of the binary file
func detectFileType(data []byte, filePath string) string {
	// Check for WASM magic number
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte{0x00, 0x61, 0x73, 0x6D}) {
		return "wasm"
	}

	// Check for ELF magic number
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte{0x7F, 0x45, 0x4C, 0x46}) {
		return "elf"
	}

	// Check for PE magic number
	if len(data) >= 2 && bytes.Equal(data[0:2], []byte{0x4D, 0x5A}) {
		return "pe"
	}

	// Check for Mach-O magic number
	if len(data) >= 4 && (binary.LittleEndian.Uint32(data[0:4]) == 0xFEEDFACE || // 32-bit
		binary.LittleEndian.Uint32(data[0:4]) == 0xFEEDFACF || // 64-bit
		binary.LittleEndian.Uint32(data[0:4]) == 0xCEFAEDFE || // 32-bit BE
		binary.LittleEndian.Uint32(data[0:4]) == 0xCFFAEDFE) { // 64-bit BE
		return "macho"
	}

	// Check file extension for scripts and other known types
	ext := strings.ToLower(filepath.Ext(filePath))

	// Script files
	if ext == ".py" || ext == ".js" || ext == ".rb" || ext == ".sh" || ext == ".pl" || ext == ".go" {
		return "script"
	}

	// JAR files
	if ext == ".jar" {
		return "jar"
	}

	// Shared objects and dynamic libraries
	if ext == ".so" || ext == ".dll" || ext == ".dylib" {
		return "library"
	}

	// Executable files
	if ext == ".exe" || ext == ".bin" || ext == ".app" || ext == "" {
		// For files without extension, try to determine if they're executable
		if ext == "" {
			// Check if file has executable permission on Unix-like systems
			fileInfo, err := os.Stat(filePath)
			if err == nil && fileInfo.Mode()&0111 != 0 {
				return "executable"
			}
		} else {
			return "executable"
		}
	}

	return "unknown"
}

// validateWasmAgent validates if a WebAssembly file is an AI agent
func validateWasmAgent(data []byte) (*AgentValidationResult, error) {
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "wasm",
		Capabilities: []string{},
	}

	// Check for exported functions that suggest agent capabilities
	// This is a simplified check - a real implementation would parse the WASM binary format

	// Check for perception functions (input interfaces)
	perceptionFuncs := []string{"sense", "input", "receive", "observe", "perceive", "get"}
	hasPerception := containsAnyString(data, perceptionFuncs)
	if hasPerception {
		result.Capabilities = append(result.Capabilities, "perception")
	}

	// Check for action functions (output interfaces)
	actionFuncs := []string{"act", "output", "send", "respond", "execute", "set"}
	hasAction := containsAnyString(data, actionFuncs)
	if hasAction {
		result.Capabilities = append(result.Capabilities, "action")
	}

	// Check for reasoning/decision functions
	reasoningFuncs := []string{"decide", "reason", "think", "process", "analyze", "evaluate"}
	hasReasoning := containsAnyString(data, reasoningFuncs)
	if hasReasoning {
		result.Capabilities = append(result.Capabilities, "reasoning")
	}

	// Check for memory/state management
	memoryIndicators := []string{"memory", "state", "store", "remember", "history", "global"}
	hasMemory := containsAnyString(data, memoryIndicators)
	if hasMemory {
		result.Capabilities = append(result.Capabilities, "memory")
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Calculate confidence based on how many core capabilities are present
		switch capabilityCount {
		case 2:
			result.Confidence = 0.5 // Minimal agent capabilities
		case 3:
			result.Confidence = 0.75 // Good confidence
		case 4:
			result.Confidence = 0.9 // High confidence
		default:
			result.Confidence = 0.3 // Low confidence
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("WASM file has %d agent capabilities", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "WASM file lacks minimum required agent capabilities")
	}

	return result, nil
}

// validateElfAgent validates if an ELF binary is an AI agent
func validateElfAgent(data []byte) (*AgentValidationResult, error) {
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "elf",
		Capabilities: []string{},
	}

	// Parse ELF file to extract symbols and sections
	elfFile, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Failed to parse ELF file: %v", err))
		return result, nil
	}

	// Check for symbols that suggest agent capabilities
	symbols, _ := elfFile.Symbols()

	// Check for perception functions
	perceptionFuncs := []string{"sense", "input", "receive", "observe", "perceive", "get"}
	hasPerception := false
	for _, sym := range symbols {
		if containsAnySubstring(sym.Name, perceptionFuncs) {
			hasPerception = true
			result.Capabilities = append(result.Capabilities, "perception")
			break
		}
	}

	// Check for action functions
	actionFuncs := []string{"act", "output", "send", "respond", "execute", "set"}
	hasAction := false
	for _, sym := range symbols {
		if containsAnySubstring(sym.Name, actionFuncs) {
			hasAction = true
			result.Capabilities = append(result.Capabilities, "action")
			break
		}
	}

	// Check for reasoning/decision functions
	reasoningFuncs := []string{"decide", "reason", "think", "process", "analyze", "evaluate"}
	hasReasoning := false
	for _, sym := range symbols {
		if containsAnySubstring(sym.Name, reasoningFuncs) {
			hasReasoning = true
			result.Capabilities = append(result.Capabilities, "reasoning")
			break
		}
	}

	// Check for memory/state management
	memoryIndicators := []string{"memory", "state", "store", "remember", "history"}
	hasMemory := false
	for _, sym := range symbols {
		if containsAnySubstring(sym.Name, memoryIndicators) {
			hasMemory = true
			result.Capabilities = append(result.Capabilities, "memory")
			break
		}
	}

	// Also check for ML/AI libraries
	aiLibraries := []string{"tensorflow", "pytorch", "onnx", "keras", "scikit", "ml", "ai", "neural"}
	for _, section := range elfFile.Sections {
		sectionData, err := section.Data()
		if err == nil {
			for _, lib := range aiLibraries {
				if bytes.Contains(bytes.ToLower(sectionData), []byte(lib)) {
					result.Capabilities = append(result.Capabilities, "ai_libraries")
					break
				}
			}
		}
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Calculate confidence based on how many core capabilities are present
		switch capabilityCount {
		case 2:
			result.Confidence = 0.5 // Minimal agent capabilities
		case 3:
			result.Confidence = 0.75 // Good confidence
		case 4, 5:
			result.Confidence = 0.9 // High confidence
		default:
			result.Confidence = 0.3 // Low confidence
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("ELF binary has %d agent capabilities", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "ELF binary lacks minimum required agent capabilities")
	}

	return result, nil
}

// validatePeAgent validates if a PE (Windows) binary is an AI agent
func validatePeAgent(data []byte) (*AgentValidationResult, error) {
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "pe",
		Capabilities: []string{},
	}

	// Parse PE file
	peFile, err := pe.NewFile(bytes.NewReader(data))
	if err != nil {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Failed to parse PE file: %v", err))
		return result, nil
	}

	// Check for imported DLLs that suggest AI capabilities
	aiDlls := []string{"tensorflow", "pytorch", "onnx", "keras", "ml", "ai", "neural", "cuda"}

	// PE file doesn't have a direct Imports field, so we need to extract this information differently
	// Check sections for DLL names
	hasAILibraries := false
	for _, section := range peFile.Sections {
		if section.Name == ".idata" || strings.Contains(section.Name, "import") {
			data, err := section.Data()
			if err == nil {
				for _, lib := range aiDlls {
					if bytes.Contains(bytes.ToLower(data), []byte(lib)) {
						result.Capabilities = append(result.Capabilities, "ai_libraries")
						hasAILibraries = true
						break
					}
				}
			}
			if hasAILibraries {
				break
			}
		}
	}

	// PE file doesn't have a direct Exports method, so we need to check sections and string data
	// Check for perception functions
	perceptionFuncs := []string{"sense", "input", "receive", "observe", "perceive", "get"}
	hasPerception := false

	// Check for action functions
	actionFuncs := []string{"act", "output", "send", "respond", "execute", "set"}
	hasAction := false

	// Check for reasoning/decision functions
	reasoningFuncs := []string{"decide", "reason", "think", "process", "analyze", "evaluate"}
	hasReasoning := false

	// Check for memory/state management
	memoryIndicators := []string{"memory", "state", "store", "remember", "history"}
	hasMemory := false

	// Check export section if available
	for _, section := range peFile.Sections {
		if section.Name == ".edata" || strings.Contains(section.Name, "export") {
			data, err := section.Data()
			if err == nil {
				// Check for capabilities in export section
				if !hasPerception {
					for _, func_ := range perceptionFuncs {
						if bytes.Contains(bytes.ToLower(data), []byte(func_)) {
							hasPerception = true
							result.Capabilities = append(result.Capabilities, "perception")
							break
						}
					}
				}

				if !hasAction {
					for _, func_ := range actionFuncs {
						if bytes.Contains(bytes.ToLower(data), []byte(func_)) {
							hasAction = true
							result.Capabilities = append(result.Capabilities, "action")
							break
						}
					}
				}

				if !hasReasoning {
					for _, func_ := range reasoningFuncs {
						if bytes.Contains(bytes.ToLower(data), []byte(func_)) {
							hasReasoning = true
							result.Capabilities = append(result.Capabilities, "reasoning")
							break
						}
					}
				}

				if !hasMemory {
					for _, func_ := range memoryIndicators {
						if bytes.Contains(bytes.ToLower(data), []byte(func_)) {
							hasMemory = true
							result.Capabilities = append(result.Capabilities, "memory")
							break
						}
					}
				}
			}
		}
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Calculate confidence based on how many core capabilities are present
		switch capabilityCount {
		case 2:
			result.Confidence = 0.5 // Minimal agent capabilities
		case 3:
			result.Confidence = 0.75 // Good confidence
		case 4, 5:
			result.Confidence = 0.9 // High confidence
		default:
			result.Confidence = 0.3 // Low confidence
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("PE binary has %d agent capabilities", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "PE binary lacks minimum required agent capabilities")
	}

	// If we didn't find any capabilities through section analysis, try string-based analysis
	if len(result.Capabilities) == 0 {
		stringData := extractStringsFromBinary(data)
		stringResult := validateBasedOnStringContent(stringData, "pe")

		// Only use string result if it found more capabilities
		if len(stringResult.Capabilities) > 0 {
			result = stringResult
		}
	}

	return result, nil
}

// validateMachoAgent validates if a Mach-O (macOS/iOS) binary is an AI agent
func validateMachoAgent(data []byte) (*AgentValidationResult, error) {
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "macho",
		Capabilities: []string{},
	}

	// Parse Mach-O file
	machoFile, err := macho.NewFile(bytes.NewReader(data))
	if err != nil {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Failed to parse Mach-O file: %v", err))
		return result, nil
	}

	// Check for imported libraries that suggest AI capabilities
	aiLibs := []string{"tensorflow", "pytorch", "onnx", "keras", "ml", "ai", "neural", "cuda"}

	// Mach-O file doesn't have a direct Imports field, so we need to check libraries differently
	// Check load commands for libraries
	hasAILibraries := false
	for _, load := range machoFile.Loads {
		// Try to extract library information from load commands
		loadBytes := []byte(fmt.Sprintf("%v", load))
		for _, aiLib := range aiLibs {
			if bytes.Contains(bytes.ToLower(loadBytes), []byte(aiLib)) {
				result.Capabilities = append(result.Capabilities, "ai_libraries")
				hasAILibraries = true
				break
			}
		}
		if hasAILibraries {
			break
		}
	}

	// Check for symbols that suggest agent capabilities
	// Check for perception functions
	perceptionFuncs := []string{"sense", "input", "receive", "observe", "perceive", "get"}
	hasPerception := false
	for _, sym := range machoFile.Symtab.Syms {
		if containsAnySubstring(sym.Name, perceptionFuncs) {
			hasPerception = true
			result.Capabilities = append(result.Capabilities, "perception")
			break
		}
	}

	// Check for action functions
	actionFuncs := []string{"act", "output", "send", "respond", "execute", "set"}
	hasAction := false
	for _, sym := range machoFile.Symtab.Syms {
		if containsAnySubstring(sym.Name, actionFuncs) {
			hasAction = true
			result.Capabilities = append(result.Capabilities, "action")
			break
		}
	}

	// Check for reasoning/decision functions
	reasoningFuncs := []string{"decide", "reason", "think", "process", "analyze", "evaluate"}
	hasReasoning := false
	for _, sym := range machoFile.Symtab.Syms {
		if containsAnySubstring(sym.Name, reasoningFuncs) {
			hasReasoning = true
			result.Capabilities = append(result.Capabilities, "reasoning")
			break
		}
	}

	// Check for memory/state management
	memoryIndicators := []string{"memory", "state", "store", "remember", "history"}
	hasMemory := false
	for _, sym := range machoFile.Symtab.Syms {
		if containsAnySubstring(sym.Name, memoryIndicators) {
			hasMemory = true
			result.Capabilities = append(result.Capabilities, "memory")
			break
		}
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Calculate confidence based on how many core capabilities are present
		switch capabilityCount {
		case 2:
			result.Confidence = 0.5 // Minimal agent capabilities
		case 3:
			result.Confidence = 0.75 // Good confidence
		case 4, 5:
			result.Confidence = 0.9 // High confidence
		default:
			result.Confidence = 0.3 // Low confidence
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("Mach-O binary has %d agent capabilities", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "Mach-O binary lacks minimum required agent capabilities")
	}

	return result, nil
}

// validateScriptAgent validates if a script file is an AI agent
func validateScriptAgent(data []byte) (*AgentValidationResult, error) {
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "script",
		Capabilities: []string{},
	}

	// Convert data to string for easier analysis
	content := string(data)

	// Check for AI/ML library imports
	aiLibraries := []string{
		"tensorflow", "torch", "pytorch", "keras", "sklearn", "scikit-learn",
		"numpy", "pandas", "transformers", "openai", "langchain", "huggingface",
		"spacy", "nltk", "gensim", "autogpt", "agent", "reinforcement",
	}

	for _, lib := range aiLibraries {
		if strings.Contains(strings.ToLower(content), "import "+lib) ||
			strings.Contains(strings.ToLower(content), "require '"+lib) ||
			strings.Contains(strings.ToLower(content), "require \""+lib) ||
			strings.Contains(strings.ToLower(content), "from "+lib) {
			result.Capabilities = append(result.Capabilities, "ai_libraries")
			break
		}
	}

	// Check for perception functions
	perceptionPatterns := []string{
		"def sense", "def input", "def receive", "def observe", "def perceive", "def get",
		"function sense", "function input", "function receive", "function observe",
		"class Sensor", "class Input", "class Perception",
	}
	hasPerception := false
	for _, pattern := range perceptionPatterns {
		if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
			hasPerception = true
			result.Capabilities = append(result.Capabilities, "perception")
			break
		}
	}

	// Check for action functions
	actionPatterns := []string{
		"def act", "def output", "def send", "def respond", "def execute", "def set",
		"function act", "function output", "function send", "function respond",
		"class Action", "class Output", "class Actuator",
	}
	hasAction := false
	for _, pattern := range actionPatterns {
		if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
			hasAction = true
			result.Capabilities = append(result.Capabilities, "action")
			break
		}
	}

	// Check for reasoning/decision functions
	reasoningPatterns := []string{
		"def decide", "def reason", "def think", "def process", "def analyze", "def evaluate",
		"function decide", "function reason", "function think", "function process",
		"class Decision", "class Reasoning", "class Brain", "class Mind",
	}
	hasReasoning := false
	for _, pattern := range reasoningPatterns {
		if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
			hasReasoning = true
			result.Capabilities = append(result.Capabilities, "reasoning")
			break
		}
	}

	// Check for memory/state management
	memoryPatterns := []string{
		"self.memory", "this.memory", "self.state", "this.state", "self.history", "this.history",
		"class Memory", "class State", "def remember", "function remember",
	}
	hasMemory := false
	for _, pattern := range memoryPatterns {
		if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
			hasMemory = true
			result.Capabilities = append(result.Capabilities, "memory")
			break
		}
	}

	// Check for autonomy indicators
	autonomyPatterns := []string{
		"while True", "while(true)", "setInterval", "setTimeout", "schedule.every",
		"infinite loop", "event loop", "main loop", "run forever", "daemon",
	}
	hasAutonomy := false
	for _, pattern := range autonomyPatterns {
		if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
			hasAutonomy = true
			result.Capabilities = append(result.Capabilities, "autonomy")
			break
		}
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Adjust confidence based on autonomy and other capabilities
		if hasAutonomy {
			result.Confidence = 0.9 // High confidence with autonomy
		} else {
			// Calculate confidence based on how many core capabilities are present
			switch capabilityCount {
			case 2:
				result.Confidence = 0.5 // Minimal agent capabilities
			case 3:
				result.Confidence = 0.7 // Good confidence
			case 4, 5:
				result.Confidence = 0.8 // High confidence
			default:
				result.Confidence = 0.3 // Low confidence
			}
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("Script has %d agent capabilities", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "Script lacks minimum required agent capabilities")
	}

	return result, nil
}

// validateJarAgent validates if a JAR file is an AI agent
func validateJarAgent(data []byte, filePath string) (*AgentValidationResult, error) {
	// filePath is unused but kept for API consistency and potential future use
	_ = filePath // Explicitly mark as unused to satisfy linter
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "jar",
		Capabilities: []string{},
	}

	// Create a temporary file to analyze
	tempFile := filepath.Join(os.TempDir(), "temp_agent.jar")
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Failed to create temporary file: %v", err))
		return result, nil
	}
	defer os.Remove(tempFile)

	// Use jar tool to list contents
	cmd := exec.Command("jar", "tf", tempFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// If jar command fails, try unzip
		cmd = exec.Command("unzip", "-l", tempFile)
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			result.Reasons = append(result.Reasons, "Failed to analyze JAR contents")
			return result, nil
		}
	}

	// Check for AI/agent related classes
	jarContents := out.String()

	// Check for AI libraries
	aiLibraries := []string{
		"tensorflow", "deeplearning", "pytorch", "keras", "weka", "dl4j", "neuroph",
		"mllib", "reinforcement", "agent", "classifier", "neural", "machinelearning",
	}

	for _, lib := range aiLibraries {
		if strings.Contains(strings.ToLower(jarContents), lib) {
			result.Capabilities = append(result.Capabilities, "ai_libraries")
			break
		}
	}

	// Check for perception classes
	perceptionClasses := []string{
		"Sensor", "Input", "Perception", "Observer", "Receiver",
	}
	hasPerception := false
	for _, class := range perceptionClasses {
		if strings.Contains(jarContents, class+".class") {
			hasPerception = true
			result.Capabilities = append(result.Capabilities, "perception")
			break
		}
	}

	// Check for action classes
	actionClasses := []string{
		"Action", "Output", "Actuator", "Effector", "Responder", "Executor",
	}
	hasAction := false
	for _, class := range actionClasses {
		if strings.Contains(jarContents, class+".class") {
			hasAction = true
			result.Capabilities = append(result.Capabilities, "action")
			break
		}
	}

	// Check for reasoning classes
	reasoningClasses := []string{
		"Decision", "Reasoning", "Brain", "Mind", "Analyzer", "Evaluator", "Processor",
	}
	hasReasoning := false
	for _, class := range reasoningClasses {
		if strings.Contains(jarContents, class+".class") {
			hasReasoning = true
			result.Capabilities = append(result.Capabilities, "reasoning")
			break
		}
	}

	// Check for memory classes
	memoryClasses := []string{
		"Memory", "State", "History", "Storage", "Cache", "Database",
	}
	hasMemory := false
	for _, class := range memoryClasses {
		if strings.Contains(jarContents, class+".class") {
			hasMemory = true
			result.Capabilities = append(result.Capabilities, "memory")
			break
		}
	}

	// Check for agent-specific classes
	agentClasses := []string{
		"Agent", "Bot", "AI", "Autonomous", "Intelligent",
	}
	for _, class := range agentClasses {
		if strings.Contains(jarContents, class+".class") {
			result.Capabilities = append(result.Capabilities, "agent_class")
			break
		}
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Calculate confidence based on how many core capabilities are present
		switch capabilityCount {
		case 2:
			result.Confidence = 0.5 // Minimal agent capabilities
		case 3:
			result.Confidence = 0.7 // Good confidence
		case 4, 5:
			result.Confidence = 0.85 // High confidence
		case 6:
			result.Confidence = 0.95 // Very high confidence
		default:
			result.Confidence = 0.3 // Low confidence
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("JAR file has %d agent capabilities", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "JAR file lacks minimum required agent capabilities")
	}

	return result, nil
}

// Helper function to validate based on string content
func validateBasedOnStringContent(content string, fileType string) *AgentValidationResult {
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    fileType,
		Capabilities: []string{},
	}

	// Check for perception functions
	perceptionFuncs := []string{"sense", "input", "receive", "observe", "perceive", "get"}
	hasPerception := false
	for _, func_ := range perceptionFuncs {
		if strings.Contains(strings.ToLower(content), func_) {
			hasPerception = true
			result.Capabilities = append(result.Capabilities, "perception")
			break
		}
	}

	// Check for action functions
	actionFuncs := []string{"act", "output", "send", "respond", "execute", "set"}
	hasAction := false
	for _, func_ := range actionFuncs {
		if strings.Contains(strings.ToLower(content), func_) {
			hasAction = true
			result.Capabilities = append(result.Capabilities, "action")
			break
		}
	}

	// Check for reasoning/decision functions
	reasoningFuncs := []string{"decide", "reason", "think", "process", "analyze", "evaluate"}
	hasReasoning := false
	for _, func_ := range reasoningFuncs {
		if strings.Contains(strings.ToLower(content), func_) {
			hasReasoning = true
			result.Capabilities = append(result.Capabilities, "reasoning")
			break
		}
	}

	// Check for memory/state management
	memoryIndicators := []string{"memory", "state", "store", "remember", "history"}
	hasMemory := false
	for _, indicator := range memoryIndicators {
		if strings.Contains(strings.ToLower(content), indicator) {
			hasMemory = true
			result.Capabilities = append(result.Capabilities, "memory")
			break
		}
	}

	// Check for AI/ML libraries
	aiLibraries := []string{"tensorflow", "pytorch", "onnx", "keras", "scikit", "ml", "ai", "neural"}
	for _, lib := range aiLibraries {
		if strings.Contains(strings.ToLower(content), lib) {
			result.Capabilities = append(result.Capabilities, "ai_libraries")
			break
		}
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Calculate confidence based on how many core capabilities are present
		switch capabilityCount {
		case 2:
			result.Confidence = 0.4 // Minimal agent capabilities, lower confidence due to string-based detection
		case 3:
			result.Confidence = 0.6 // Moderate confidence
		case 4, 5:
			result.Confidence = 0.75 // Good confidence
		default:
			result.Confidence = 0.2 // Low confidence
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("Binary has %d agent capabilities based on string analysis", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "Binary lacks minimum required agent capabilities based on string analysis")
	}

	return result
}

// Helper function to extract strings from binary
func extractStringsFromBinary(data []byte) string {
	var result strings.Builder

	// Simple string extraction - in a real implementation, you would use a more sophisticated approach
	inString := false
	var currentString strings.Builder

	for _, b := range data {
		if b >= 32 && b <= 126 { // Printable ASCII
			if !inString {
				inString = true
			}
			currentString.WriteByte(b)
		} else {
			if inString && currentString.Len() >= 4 { // Only keep strings of reasonable length
				result.WriteString(currentString.String())
				result.WriteString("\n")
			}
			inString = false
			currentString.Reset()
		}
	}

	return result.String()
}

// Helper function to check if a byte slice contains any of the given strings
func containsAnyString(data []byte, strings []string) bool {
	for _, s := range strings {
		if bytes.Contains(bytes.ToLower(data), []byte(s)) {
			return true
		}
	}
	return false
}

// validateLibraryAgent validates if a shared library or DLL is an AI agent
func validateLibraryAgent(data []byte) (*AgentValidationResult, error) {
	result := &AgentValidationResult{
		IsAgent:      false,
		Confidence:   0.0,
		Reasons:      []string{},
		AgentType:    "library",
		Capabilities: []string{},
	}

	// Determine the library format based on magic numbers
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte{0x7F, 0x45, 0x4C, 0x46}) {
		// It's an ELF shared object
		return validateElfAgent(data)
	} else if len(data) >= 2 && bytes.Equal(data[0:2], []byte{0x4D, 0x5A}) {
		// It's a PE/DLL file
		return validatePeAgent(data)
	} else if len(data) >= 4 && (binary.LittleEndian.Uint32(data[0:4]) == 0xFEEDFACE ||
		binary.LittleEndian.Uint32(data[0:4]) == 0xFEEDFACF ||
		binary.LittleEndian.Uint32(data[0:4]) == 0xCEFAEDFE ||
		binary.LittleEndian.Uint32(data[0:4]) == 0xCFFAEDFE) {
		// It's a Mach-O dylib
		return validateMachoAgent(data)
	}

	// If we can't determine the format, extract strings and analyze
	stringData := extractStringsFromBinary(data)

	// Check for perception functions
	perceptionFuncs := []string{"sense", "input", "receive", "observe", "perceive", "get"}
	hasPerception := false
	for _, func_ := range perceptionFuncs {
		if strings.Contains(strings.ToLower(stringData), func_) {
			hasPerception = true
			result.Capabilities = append(result.Capabilities, "perception")
			break
		}
	}

	// Check for action functions
	actionFuncs := []string{"act", "output", "send", "respond", "execute", "set"}
	hasAction := false
	for _, func_ := range actionFuncs {
		if strings.Contains(strings.ToLower(stringData), func_) {
			hasAction = true
			result.Capabilities = append(result.Capabilities, "action")
			break
		}
	}

	// Check for reasoning/decision functions
	reasoningFuncs := []string{"decide", "reason", "think", "process", "analyze", "evaluate"}
	hasReasoning := false
	for _, func_ := range reasoningFuncs {
		if strings.Contains(strings.ToLower(stringData), func_) {
			hasReasoning = true
			result.Capabilities = append(result.Capabilities, "reasoning")
			break
		}
	}

	// Check for memory/state management
	memoryIndicators := []string{"memory", "state", "store", "remember", "history"}
	hasMemory := false
	for _, indicator := range memoryIndicators {
		if strings.Contains(strings.ToLower(stringData), indicator) {
			hasMemory = true
			result.Capabilities = append(result.Capabilities, "memory")
			break
		}
	}

	// Check for AI/ML libraries
	aiLibraries := []string{"tensorflow", "pytorch", "onnx", "keras", "scikit", "ml", "ai", "neural"}
	for _, lib := range aiLibraries {
		if strings.Contains(strings.ToLower(stringData), lib) {
			result.Capabilities = append(result.Capabilities, "ai_libraries")
			break
		}
	}

	// Calculate confidence based on capabilities
	capabilityCount := len(result.Capabilities)

	// An agent needs at minimum: perception, action, and either reasoning or memory
	if hasPerception && hasAction && (hasReasoning || hasMemory) {
		result.IsAgent = true

		// Calculate confidence based on how many core capabilities are present
		switch capabilityCount {
		case 2:
			result.Confidence = 0.4 // Minimal agent capabilities, lower confidence due to string-based detection
		case 3:
			result.Confidence = 0.6 // Moderate confidence
		case 4, 5:
			result.Confidence = 0.75 // Good confidence
		default:
			result.Confidence = 0.2 // Low confidence
		}

		result.Reasons = append(result.Reasons, fmt.Sprintf("Library has %d agent capabilities based on string analysis", capabilityCount))
	} else {
		result.Reasons = append(result.Reasons, "Library lacks minimum required agent capabilities based on string analysis")
	}

	return result, nil
}

// Helper function to check if a string contains any of the given substrings
func containsAnySubstring(s string, substrings []string) bool {
	lowerS := strings.ToLower(s)
	for _, sub := range substrings {
		if strings.Contains(lowerS, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
