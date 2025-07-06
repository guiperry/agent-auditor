package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"sync"
)

type AuditLogger struct {
	logFile *os.File
	mutex   sync.Mutex
}

func NewAuditLogger() *AuditLogger {
	logFile, err := os.OpenFile("aegong_audit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open audit log file:", err)
	}

	return &AuditLogger{
		logFile: logFile,
	}
}

func (a *AuditLogger) LogAudit(report *AuditReport) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Create immutable log entry
	logEntry := map[string]interface{}{
		"timestamp":       report.Timestamp,
		"agent_hash":      report.AgentHash,
		"threat_count":    len(report.Threats),
		"overall_risk":    report.OverallRisk,
		"threats":         report.Threats,
		"shield_results":  report.ShieldResults,
		"recommendations": report.Recommendations,
	}

	// Sign the log entry
	signature := a.signLogEntry(logEntry)
	logEntry["signature"] = signature

	// Write to log
	jsonData, _ := json.Marshal(logEntry)
	a.logFile.WriteString(string(jsonData) + "\n")
	a.logFile.Sync()
}

func (a *AuditLogger) signLogEntry(entry map[string]interface{}) string {
	// Create a simple signature for the log entry
	jsonData, _ := json.Marshal(entry)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

func (a *AuditLogger) Close() {
	a.logFile.Close()
}
