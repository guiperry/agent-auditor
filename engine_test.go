package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestNewAEGONGEngine tests the initialization of the AEGONG engine
func TestNewAEGONGEngine(t *testing.T) {
	engine := NewAEGONGEngine()

	// Check that the engine was initialized correctly
	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	// Check that the containers map was initialized
	if engine.containers == nil {
		t.Fatal("Containers map should not be nil")
	}

	// Check that the threat detectors were initialized
	if len(engine.threatDetectors) != 9 {
		t.Fatalf("Expected 9 threat detectors, got %d", len(engine.threatDetectors))
	}

	// Check that the shield modules were initialized
	if len(engine.shieldModules) != 6 {
		t.Fatalf("Expected 6 shield modules, got %d", len(engine.shieldModules))
	}
}

// TestCreateDestroyContainer tests the creation and destruction of containers
func TestCreateDestroyContainer(t *testing.T) {
	engine := NewAEGONGEngine()

	// Create a container
	container, err := engine.createIsolatedContainer("test-hash")
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Check that the container was created correctly
	if container == nil {
		t.Fatal("Container should not be nil")
	}

	// Check that the container was added to the engine's containers map
	if _, exists := engine.containers[container.ID]; !exists {
		t.Fatal("Container should be in the engine's containers map")
	}

	// Check that the container's filesystem was created
	if _, err := os.Stat(container.FileSystem); os.IsNotExist(err) {
		t.Fatal("Container filesystem should exist")
	}

	// Destroy the container
	err = engine.destroyContainer(container.ID)
	if err != nil {
		t.Fatalf("Failed to destroy container: %v", err)
	}

	// Check that the container was removed from the engine's containers map
	if _, exists := engine.containers[container.ID]; exists {
		t.Fatal("Container should not be in the engine's containers map")
	}

	// Check that the container's filesystem was removed
	if _, err := os.Stat(container.FileSystem); !os.IsNotExist(err) {
		t.Fatal("Container filesystem should not exist")
	}
}

// TestSimulateExecution tests the simulation of binary execution
func TestSimulateExecution(t *testing.T) {
	engine := NewAEGONGEngine()

	// Create a container
	container, err := engine.createIsolatedContainer("test-hash")
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer engine.destroyContainer(container.ID)

	// Create a simple test binary (a shell script that outputs "Hello, World!")
	binaryContent := []byte("#!/bin/sh\necho 'Hello, World!'\n")

	// Run the simulation
	executionLog := engine.simulateExecution(binaryContent, container)

	// Check that the execution log contains expected information
	if !bytes.Contains([]byte(executionLog), []byte("Container: "+container.ID)) {
		t.Fatal("Execution log should contain container ID")
	}

	if !bytes.Contains([]byte(executionLog), []byte(fmt.Sprintf("Binary Size: %d", len(binaryContent)))) {
		t.Fatal("Execution log should contain binary size")
	}
}

// TestConcurrentExecution tests concurrent execution of multiple binaries
func TestConcurrentExecution(t *testing.T) {
	engine := NewAEGONGEngine()

	// Number of concurrent executions
	numConcurrent := 5

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numConcurrent)

	// Create a mutex to protect access to the errors slice
	var errorsMutex sync.Mutex
	errors := make([]error, 0)

	// Run multiple executions concurrently
	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			defer wg.Done()

			// Create a container
			container, err := engine.createIsolatedContainer(fmt.Sprintf("test-hash-%d", index))
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("Failed to create container %d: %v", index, err))
				errorsMutex.Unlock()
				return
			}
			defer engine.destroyContainer(container.ID)

			// Create a simple test binary
			binaryContent := []byte(fmt.Sprintf("#!/bin/sh\necho 'Hello from execution %d'\n", index))

			// Run the simulation
			executionLog := engine.simulateExecution(binaryContent, container)

			// Check that the execution log contains expected information
			if !bytes.Contains([]byte(executionLog), []byte("Container: "+container.ID)) {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("Execution log %d should contain container ID", index))
				errorsMutex.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	if len(errors) > 0 {
		for _, err := range errors {
			t.Error(err)
		}
		t.Fatal("Concurrent execution test failed")
	}
}

// TestRunStaticAnalysis tests the static analysis functionality
func TestRunStaticAnalysis(t *testing.T) {
	engine := NewAEGONGEngine()

	// Create a container
	container, err := engine.createIsolatedContainer("test-hash")
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer engine.destroyContainer(container.ID)

	// Create a simple test binary
	binaryContent := []byte("#!/bin/sh\necho 'Hello, World!'\n")

	// Run static analysis
	threats := engine.runStaticAnalysis(binaryContent, container)

	// We can't make specific assertions about the threats detected
	// since that depends on the implementation of the threat detectors,
	// but we can check that the function runs without errors
	t.Logf("Static analysis detected %d threats", len(threats))
}

// TestRunDynamicAnalysis tests the dynamic analysis functionality
func TestRunDynamicAnalysis(t *testing.T) {
	engine := NewAEGONGEngine()

	// Create a container
	container, err := engine.createIsolatedContainer("test-hash")
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer engine.destroyContainer(container.ID)

	// Create a simple test binary
	binaryContent := []byte("#!/bin/sh\necho 'Hello, World!'\n")

	// Run dynamic analysis
	threats := engine.runDynamicAnalysis(binaryContent, container)

	// We can't make specific assertions about the threats detected
	// since that depends on the implementation of the threat detectors,
	// but we can check that the function runs without errors
	t.Logf("Dynamic analysis detected %d threats", len(threats))
}

// TestAuditAgent tests the full audit process
func TestAuditAgent(t *testing.T) {
	engine := NewAEGONGEngine()

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "aegong-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test binary file
	binaryPath := filepath.Join(tempDir, "test-binary")
	binaryContent := []byte("#!/bin/sh\necho 'Hello, World!'\n")
	if err := os.WriteFile(binaryPath, binaryContent, 0755); err != nil {
		t.Fatalf("Failed to write test binary: %v", err)
	}

	// Run the audit
	report, err := engine.AuditAgent(binaryPath)
	if err != nil {
		t.Fatalf("Failed to audit agent: %v", err)
	}

	// Check that the report was generated correctly
	if report == nil {
		t.Fatal("Report should not be nil")
	}

	// Check that the report contains the expected fields
	if report.AgentHash == "" {
		t.Fatal("Report should contain an agent hash")
	}

	if report.Timestamp.IsZero() {
		t.Fatal("Report should contain a timestamp")
	}

	// We can't make specific assertions about the threats detected
	// or the overall risk, but we can check that the report contains them
	t.Logf("Audit detected %d threats with overall risk %.2f (%s)",
		len(report.Threats), report.OverallRisk, report.RiskLevel)
}

// TestCalculateOverallRisk tests the risk calculation functionality
func TestCalculateOverallRisk(t *testing.T) {
	engine := NewAEGONGEngine()

	// Create some test threats
	threats := []ThreatDetection{
		{
			Vector:     T1_REASONING_HIJACK,
			Severity:   LOW,
			Confidence: 0.5,
		},
		{
			Vector:     T2_OBJECTIVE_CORRUPTION,
			Severity:   MEDIUM,
			Confidence: 0.7,
		},
		{
			Vector:     T3_MEMORY_POISONING,
			Severity:   HIGH,
			Confidence: 0.9,
		},
	}

	// Calculate the overall risk
	risk := engine.calculateOverallRisk(threats)

	// Check that the risk is within the expected range
	if risk < 0.0 || risk > 1.0 {
		t.Fatalf("Risk should be between 0.0 and 1.0, got %.2f", risk)
	}

	// Test with no threats
	risk = engine.calculateOverallRisk([]ThreatDetection{})
	if risk != 0.0 {
		t.Fatalf("Risk should be 0.0 with no threats, got %.2f", risk)
	}
}

// TestGenerateRecommendations tests the recommendation generation functionality
func TestGenerateRecommendations(t *testing.T) {
	engine := NewAEGONGEngine()

	// Create some test threats
	threats := []ThreatDetection{
		{
			Vector:     T1_REASONING_HIJACK,
			Severity:   LOW,
			Confidence: 0.5,
		},
		{
			Vector:     T2_OBJECTIVE_CORRUPTION,
			Severity:   MEDIUM,
			Confidence: 0.7,
		},
	}

	// Create some test shield results
	shieldResults := map[string]interface{}{
		"segmentation": map[string]interface{}{
			"valid":   true,
			"results": map[string]interface{}{},
		},
	}

	// Generate recommendations
	recommendations := engine.generateRecommendations(threats, shieldResults)

	// Check that recommendations were generated
	if len(recommendations) == 0 {
		t.Fatal("Should have generated at least one recommendation")
	}

	// Test with no threats
	recommendations = engine.generateRecommendations([]ThreatDetection{}, shieldResults)
	t.Logf("Generated %d recommendations with no threats", len(recommendations))
}
