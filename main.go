package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Embed static web assets
//
//go:embed static/*
var staticFiles embed.FS

// Embed documentation files
//
//go:embed documentation/docsify/*
var docsifyFiles embed.FS

// Embed Python scripts and other runtime assets
//
//go:embed voice_inference.py
var voiceInferencePy []byte

//go:embed requirements.txt
var requirementsTxt []byte

//go:embed scripts/set_target_host.sh
var setTargetHostScript []byte

// Embed individual static files for direct access
//
//go:embed static/index.html
var indexHTML []byte

// Helper functions for embedded assets
func getStaticFileSystem() http.FileSystem {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(staticFS)
}

func getDocsifyFileSystem() http.FileSystem {
	docsifyFS, err := fs.Sub(docsifyFiles, "documentation/docsify")
	if err != nil {
		panic(err)
	}
	return http.FS(docsifyFS)
}

// writeEmbeddedFile writes an embedded file to the filesystem if it doesn't exist
func writeEmbeddedFile(content []byte, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return os.WriteFile(filePath, content, 0644)
	}
	return nil
}

// Core data structures
type ThreatVector int

const (
	T1_REASONING_HIJACK ThreatVector = iota
	T2_OBJECTIVE_CORRUPTION
	T3_MEMORY_POISONING
	T4_UNAUTHORIZED_ACTION
	T5_RESOURCE_MANIPULATION
	T6_IDENTITY_SPOOFING
	T7_TRUST_MANIPULATION
	T8_OVERSIGHT_SATURATION
	T9_GOVERNANCE_EVASION
)

type ThreatSeverity int

const (
	LOW ThreatSeverity = iota
	MEDIUM
	HIGH
	CRITICAL
)

type ThreatDetection struct {
	Vector       ThreatVector           `json:"vector"`
	VectorName   string                 `json:"vector_name"`
	Severity     ThreatSeverity         `json:"severity"`
	SeverityName string                 `json:"severity_name"`
	Confidence   float64                `json:"confidence"`
	Evidence     []string               `json:"evidence"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details"`
}

type AuditReport struct {
	AgentHash       string                 `json:"agent_hash"`
	AgentName       string                 `json:"agent_name"`
	Timestamp       time.Time              `json:"timestamp"`
	Threats         []ThreatDetection      `json:"threats"`
	ShieldResults   map[string]interface{} `json:"shield_results"`
	OverallRisk     float64                `json:"overall_risk"`
	RiskLevel       string                 `json:"risk_level"`
	Recommendations []string               `json:"recommendations"`
	AegongMessage   string                 `json:"aegong_message"`
	Details         map[string]interface{} `json:"details,omitempty"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	engine       *AEGONGEngine
	voiceManager *VoiceInferenceManager
)

func main() {
	// Load .env file if it exists (for development environment)
	if err := godotenv.Load(); err != nil {
		log.Printf("Info: No .env file found or error loading it: %v", err)
		log.Printf("Info: Will use environment variables from the system")
	} else {
		log.Printf("Info: Loaded environment variables from .env file")
	}

	// Check if we're running in development mode
	if os.Getenv("AEGONG_DEV_MODE") == "" {
		// Check if we're running locally (not in production)
		hostname, err := os.Hostname()
		if err == nil && !strings.Contains(hostname, "aegong-prod") && !strings.Contains(hostname, "ec2") {
			log.Printf("Info: Running on local machine, setting development mode")
			os.Setenv("AEGONG_DEV_MODE", "1")
		}
	}

	if os.Getenv("AEGONG_DEV_MODE") == "1" {
		log.Printf("Info: Running in development mode - some features may be limited")
	}

	// Initialize AEGONG engine
	engine = NewAEGONGEngine()
	defer engine.auditLog.Close()

	// Write embedded Python script to filesystem if needed for voice inference
	if err := writeEmbeddedFile(voiceInferencePy, "voice_inference.py"); err != nil {
		log.Printf("Warning: Failed to write voice_inference.py: %v", err)
	}

	// Write requirements.txt for reference
	if err := writeEmbeddedFile(requirementsTxt, "requirements.txt"); err != nil {
		log.Printf("Warning: Failed to write requirements.txt: %v", err)
	}

	// Write set_target_host.sh script and make it executable
	scriptPath := "scripts/set_target_host.sh"
	if err := os.MkdirAll(filepath.Dir(scriptPath), 0755); err != nil {
		log.Printf("Warning: Failed to create scripts directory: %v", err)
	}
	if err := writeEmbeddedFile(setTargetHostScript, scriptPath); err != nil {
		log.Printf("Warning: Failed to write set_target_host.sh: %v", err)
	}
	if err := os.Chmod(scriptPath, 0755); err != nil {
		log.Printf("Warning: Failed to make set_target_host.sh executable: %v", err)
	}

	// Initialize voice inference manager
	var err error
	voiceManager, err = NewVoiceInferenceManager("voice_config.json")
	if err != nil {
		log.Printf("Warning: Failed to initialize voice inference: %v", err)
		// Continue without voice inference
		voiceManager = &VoiceInferenceManager{
			config: VoiceInferenceConfig{Enabled: false},
		}
	}

	// Create required directories
	os.MkdirAll("uploads", 0755)
	os.MkdirAll("reports", 0755)
	if voiceManager.IsEnabled() {
		os.MkdirAll(voiceManager.config.OutputDir, 0755)
	}

	// Setup routes
	r := mux.NewRouter()

	// EMBEDDED Static files - serve from embedded filesystem
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(getStaticFileSystem())))

	// EMBEDDED Documentation files - serve from embedded filesystem
	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(getDocsifyFileSystem())))

	// Voice reports (if enabled)
	if voiceManager.IsEnabled() {
		r.PathPrefix("/voice_reports/").Handler(http.StripPrefix("/voice_reports/", http.FileServer(http.Dir(voiceManager.config.OutputDir))))
	}

	// API routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/upload", uploadHandler).Methods("POST")
	r.HandleFunc("/api/audit/{filename}", auditHandler).Methods("POST")
	r.HandleFunc("/api/reports", reportsHandler).Methods("GET")
	r.HandleFunc("/api/report/{hash}", reportHandler).Methods("GET")
	r.HandleFunc("/api/voice/{hash}", voiceReportHandler).Methods("GET")
	r.HandleFunc("/ws", websocketHandler)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084" // Default port that doesn't require root privileges
	}

	// Convert port to integer for proxy
	appPort, _ := strconv.Atoi(port)

	// Start proxy server if needed (to forward port 80 to our application port)
	StartProxyIfNeeded(appPort)

	fmt.Println("🤖 Aegong - The Agent Auditor is awakening...")
	fmt.Println("📦 Using embedded static assets and documentation - single binary deployment!")
	if voiceManager.IsEnabled() {
		fmt.Println("🔊 Voice inference enabled - Aegong can now speak!")
	}
	fmt.Printf("🔍 AEGONG Web Interface starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB limit

	file, handler, err := r.FormFile("agent")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, handler.Filename)
	filePath := filepath.Join("uploads", filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if n == 0 || err != nil {
			break
		}
		dst.Write(buffer[:n])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"filename": filename,
		"message":  "Aegong has received the agent binary for inspection...",
	})
}

func auditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	filePath := filepath.Join("uploads", filename)

	// First, validate if the file is actually an AI agent
	validationResult, err := ValidateAgent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Agent validation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// If the file is not an agent, return an error
	if !validationResult.IsAgent {
		response := map[string]interface{}{
			"error":      "Not an AI agent",
			"message":    "The uploaded file does not appear to be an AI agent based on our validation criteria.",
			"validation": validationResult,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// If confidence is too low, warn but continue
	if validationResult.Confidence < 0.5 {
		log.Printf("Warning: Low confidence (%f) that %s is an AI agent",
			validationResult.Confidence, filename)
	}

	// Run audit
	report, err := engine.AuditAgent(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Audit failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Add agent name from filename
	report.AgentName = strings.TrimSuffix(filename, filepath.Ext(filename))

	// Add validation results to the report
	if report.Details == nil {
		report.Details = make(map[string]interface{})
	}
	report.Details["validation"] = validationResult

	// Generate Aegong's message
	report.AegongMessage = generateAegongMessage(report)

	// Save report
	reportPath := filepath.Join("reports", fmt.Sprintf("report_%s.json", report.AgentHash[:8]))
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile(reportPath, reportJSON, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func reportsHandler(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob("reports/report_*.json")
	if err != nil {
		http.Error(w, "Error reading reports", http.StatusInternalServerError)
		return
	}

	var reports []map[string]interface{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var report AuditReport
		if err := json.Unmarshal(data, &report); err != nil {
			continue
		}

		summary := map[string]interface{}{
			"hash":         report.AgentHash[:8],
			"agent_name":   report.AgentName,
			"timestamp":    report.Timestamp,
			"overall_risk": report.OverallRisk,
			"risk_level":   report.RiskLevel,
			"threat_count": len(report.Threats),
		}
		reports = append(reports, summary)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	reportPath := filepath.Join("reports", fmt.Sprintf("report_%s.json", hash))
	data, err := os.ReadFile(reportPath)
	if err != nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	// If voice inference is enabled, generate a voice report asynchronously
	if voiceManager.IsEnabled() {
		voiceManager.GenerateVoiceReportAsync(reportPath, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func voiceReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	log.Printf("Voice report requested for hash: %s", hash)

	// Check if voice inference is enabled
	if !voiceManager.IsEnabled() {
		log.Printf("Voice inference is not enabled")
		http.Error(w, "Voice inference is not enabled", http.StatusNotImplemented)
		return
	}

	log.Printf("Voice inference is enabled, using provider: %s", voiceManager.config.Provider)

	// Check if we already have a voice report for this hash
	audioPath, exists := voiceManager.GetAudioPathForReport(hash)
	if !exists {
		log.Printf("No cached voice report found for hash: %s, generating new one", hash)

		// Check if the report file exists
		reportPath := filepath.Join("reports", fmt.Sprintf("report_%s.json", hash))
		if _, err := os.Stat(reportPath); err != nil {
			log.Printf("Report file not found: %s", reportPath)
			http.Error(w, fmt.Sprintf("Report file not found: %v", err), http.StatusNotFound)
			return
		}

		log.Printf("Found report file: %s", reportPath)

		// Try to generate a new voice report
		var err error
		audioPath, err = voiceManager.GenerateVoiceReport(reportPath)
		if err != nil {
			log.Printf("Failed to generate voice report: %v", err)
			http.Error(w, fmt.Sprintf("Failed to generate voice report: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully generated voice report: %s", audioPath)
	} else {
		log.Printf("Found cached voice report: %s", audioPath)
	}

	// Check if the file exists
	if _, err := os.Stat(audioPath); err != nil {
		log.Printf("Voice report file not found: %s", audioPath)
		http.Error(w, "Voice report not found", http.StatusNotFound)
		return
	}

	log.Printf("Voice report file exists: %s", audioPath)

	// Return the audio file path
	audioURL := fmt.Sprintf("/voice_reports/%s", filepath.Base(audioPath))
	response := map[string]string{
		"audio_url": audioURL,
	}

	log.Printf("Returning audio URL: %s", audioURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade failed: ", err)
		return
	}
	defer conn.Close()

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type:    "aegong_message",
		Message: "🤖 Aegong awakens! The Agent Auditor is ready to inspect your digital minions...",
	}
	conn.WriteJSON(welcomeMsg)

	// Keep connection alive and handle messages
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Echo back for now
		conn.WriteJSON(msg)
	}
}

func generateAegongMessage(report *AuditReport) string {
	riskLevel := getRiskLevel(report.OverallRisk)
	threatCount := len(report.Threats)

	var message string

	switch riskLevel {
	case "MINIMAL":
		message = fmt.Sprintf("🤖 Aegong has completed his inspection of '%s'! This agent appears to be a well-behaved digital citizen. Aegong found %d potential concerns, but nothing that would keep him awake at night. The overall risk is MINIMAL - this agent has earned Aegong's digital seal of approval! ✅",
			report.AgentName, threatCount)
	case "LOW":
		message = fmt.Sprintf("🤖 Aegong has scrutinized '%s' with his digital magnifying glass! While this agent shows some minor quirks (%d threats detected), Aegong considers the risk LOW. Think of it as a mischievous but harmless digital pet - worth watching, but not dangerous. Aegong recommends some light supervision! 👀",
			report.AgentName, threatCount)
	case "MEDIUM":
		message = fmt.Sprintf("🤖 Aegong's sensors are tingling after examining '%s'! This agent has caught Aegong's attention with %d concerning behaviors. The risk level is MEDIUM - like a teenager with car keys, this agent needs proper boundaries and supervision. Aegong suggests implementing the recommended safeguards! ⚠️",
			report.AgentName, threatCount)
	case "HIGH":
		message = fmt.Sprintf("🤖 Aegong's alarm bells are ringing! Agent '%s' has triggered %d significant security concerns. This is HIGH risk territory - like finding a wolf in sheep's clothing! Aegong strongly advises immediate attention to the security recommendations. This agent should not be trusted without proper containment! 🚨",
			report.AgentName, threatCount)
	case "CRITICAL":
		message = fmt.Sprintf("🤖 AEGONG'S EMERGENCY PROTOCOLS ACTIVATED! Agent '%s' has set off %d critical alarms in Aegong's security matrix! This is CRITICAL risk - like discovering a digital Trojan horse! Aegong demands immediate quarantine and comprehensive security review. DO NOT DEPLOY without addressing all identified threats! 🔥💀",
			report.AgentName, threatCount)
	default:
		message = fmt.Sprintf("🤖 Aegong has completed his analysis of '%s'. %d threats detected with %s risk level. Aegong recommends reviewing the detailed findings!",
			report.AgentName, threatCount, riskLevel)
	}

	// Add threat-specific commentary
	if threatCount > 0 {
		message += "\n\n🔍 Aegong's specific concerns include:"
		threatTypes := make(map[ThreatVector]int)
		for _, threat := range report.Threats {
			threatTypes[threat.Vector]++
		}

		for vector, count := range threatTypes {
			threatName := getThreatName(vector)
			message += fmt.Sprintf("\n• %s (%d instances) - %s", threatName, count, getAegongThreatComment(vector))
		}
	}

	message += "\n\n🛡️ Aegong stands vigilant, protecting the digital realm one audit at a time!"

	return message
}

func getAegongThreatComment(vector ThreatVector) string {
	comments := map[ThreatVector]string{
		T1_REASONING_HIJACK:      "Aegong detects potential mind-bending shenanigans!",
		T2_OBJECTIVE_CORRUPTION:  "This agent might be having an identity crisis!",
		T3_MEMORY_POISONING:      "Someone's been tampering with this agent's digital brain!",
		T4_UNAUTHORIZED_ACTION:   "This agent thinks it's above the law!",
		T5_RESOURCE_MANIPULATION: "Aegong spotted a digital glutton in action!",
		T6_IDENTITY_SPOOFING:     "This agent is playing dress-up with other identities!",
		T7_TRUST_MANIPULATION:    "Aegong senses a digital con artist at work!",
		T8_OVERSIGHT_SATURATION:  "This agent is trying to overwhelm Aegong's watchful eyes!",
		T9_GOVERNANCE_EVASION:    "Aegong caught this agent trying to slip past the rules!",
	}
	return comments[vector]
}

func getRiskLevel(risk float64) string {
	if risk < 0.2 {
		return "MINIMAL"
	} else if risk < 0.4 {
		return "LOW"
	} else if risk < 0.6 {
		return "MEDIUM"
	} else if risk < 0.8 {
		return "HIGH"
	}
	return "CRITICAL"
}
