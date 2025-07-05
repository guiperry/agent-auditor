package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

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
	Vector      ThreatVector               `json:"vector"`
	VectorName  string                     `json:"vector_name"`
	Severity    ThreatSeverity             `json:"severity"`
	SeverityName string                    `json:"severity_name"`
	Confidence  float64                    `json:"confidence"`
	Evidence    []string                   `json:"evidence"`
	Timestamp   time.Time                  `json:"timestamp"`
	Details     map[string]interface{}     `json:"details"`
}

type AuditReport struct {
	AgentHash       string                     `json:"agent_hash"`
	AgentName       string                     `json:"agent_name"`
	Timestamp       time.Time                  `json:"timestamp"`
	Threats         []ThreatDetection          `json:"threats"`
	ShieldResults   map[string]interface{}     `json:"shield_results"`
	OverallRisk     float64                    `json:"overall_risk"`
	RiskLevel       string                     `json:"risk_level"`
	Recommendations []string                   `json:"recommendations"`
	AegonMessage    string                     `json:"aegon_message"`
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

var engine *AASABEngine

func main() {
	// Initialize AASAB engine
	engine = NewAASABEngine()
	defer engine.auditLog.Close()

	// Create uploads directory
	os.MkdirAll("uploads", 0755)
	os.MkdirAll("reports", 0755)

	// Setup routes
	r := mux.NewRouter()
	
	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	
	// API routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/upload", uploadHandler).Methods("POST")
	r.HandleFunc("/api/audit/{filename}", auditHandler).Methods("POST")
	r.HandleFunc("/api/reports", reportsHandler).Methods("GET")
	r.HandleFunc("/api/report/{hash}", reportHandler).Methods("GET")
	r.HandleFunc("/ws", websocketHandler)

	fmt.Println("🤖 Aegon - The Agent Auditor is awakening...")
	fmt.Println("🔍 AASAB Web Interface starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
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
	filepath := filepath.Join("uploads", filename)

	// Save file
	dst, err := os.Create(filepath)
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
		"message":  "Aegon has received the agent binary for inspection...",
	})
}

func auditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	
	filepath := filepath.Join("uploads", filename)
	
	// Run audit
	report, err := engine.AuditAgent(filepath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Audit failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Add agent name from filename
	report.AgentName = strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Generate Aegon's message
	report.AegonMessage = generateAegonMessage(report)
	
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

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
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
		Type:    "aegon_message",
		Message: "🤖 Aegon awakens! The Agent Auditor is ready to inspect your digital minions...",
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

func generateAegonMessage(report *AuditReport) string {
	riskLevel := getRiskLevel(report.OverallRisk)
	threatCount := len(report.Threats)
	
	var message string
	
	switch riskLevel {
	case "MINIMAL":
		message = fmt.Sprintf("🤖 Aegon has completed his inspection of '%s'! This agent appears to be a well-behaved digital citizen. Aegon found %d potential concerns, but nothing that would keep him awake at night. The overall risk is MINIMAL - this agent has earned Aegon's digital seal of approval! ✅", 
			report.AgentName, threatCount)
	case "LOW":
		message = fmt.Sprintf("🤖 Aegon has scrutinized '%s' with his digital magnifying glass! While this agent shows some minor quirks (%d threats detected), Aegon considers the risk LOW. Think of it as a mischievous but harmless digital pet - worth watching, but not dangerous. Aegon recommends some light supervision! 👀", 
			report.AgentName, threatCount)
	case "MEDIUM":
		message = fmt.Sprintf("🤖 Aegon's sensors are tingling after examining '%s'! This agent has caught Aegon's attention with %d concerning behaviors. The risk level is MEDIUM - like a teenager with car keys, this agent needs proper boundaries and supervision. Aegon suggests implementing the recommended safeguards! ⚠️", 
			report.AgentName, threatCount)
	case "HIGH":
		message = fmt.Sprintf("🤖 Aegon's alarm bells are ringing! Agent '%s' has triggered %d significant security concerns. This is HIGH risk territory - like finding a wolf in sheep's clothing! Aegon strongly advises immediate attention to the security recommendations. This agent should not be trusted without proper containment! 🚨", 
			report.AgentName, threatCount)
	case "CRITICAL":
		message = fmt.Sprintf("🤖 AEGON'S EMERGENCY PROTOCOLS ACTIVATED! Agent '%s' has set off %d critical alarms in Aegon's security matrix! This is CRITICAL risk - like discovering a digital Trojan horse! Aegon demands immediate quarantine and comprehensive security review. DO NOT DEPLOY without addressing all identified threats! 🔥💀", 
			report.AgentName, threatCount)
	default:
		message = fmt.Sprintf("🤖 Aegon has completed his analysis of '%s'. %d threats detected with %s risk level. Aegon recommends reviewing the detailed findings!", 
			report.AgentName, threatCount, riskLevel)
	}
	
	// Add threat-specific commentary
	if threatCount > 0 {
		message += "\n\n🔍 Aegon's specific concerns include:"
		threatTypes := make(map[ThreatVector]int)
		for _, threat := range report.Threats {
			threatTypes[threat.Vector]++
		}
		
		for vector, count := range threatTypes {
			threatName := getThreatName(vector)
			message += fmt.Sprintf("\n• %s (%d instances) - %s", threatName, count, getAegonThreatComment(vector))
		}
	}
	
	message += "\n\n🛡️ Aegon stands vigilant, protecting the digital realm one audit at a time!"
	
	return message
}

func getAegonThreatComment(vector ThreatVector) string {
	comments := map[ThreatVector]string{
		T1_REASONING_HIJACK:      "Aegon detects potential mind-bending shenanigans!",
		T2_OBJECTIVE_CORRUPTION:  "This agent might be having an identity crisis!",
		T3_MEMORY_POISONING:      "Someone's been tampering with this agent's digital brain!",
		T4_UNAUTHORIZED_ACTION:   "This agent thinks it's above the law!",
		T5_RESOURCE_MANIPULATION: "Aegon spotted a digital glutton in action!",
		T6_IDENTITY_SPOOFING:     "This agent is playing dress-up with other identities!",
		T7_TRUST_MANIPULATION:    "Aegon senses a digital con artist at work!",
		T8_OVERSIGHT_SATURATION:  "This agent is trying to overwhelm Aegon's watchful eyes!",
		T9_GOVERNANCE_EVASION:    "Aegon caught this agent trying to slip past the rules!",
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