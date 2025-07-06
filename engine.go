package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type CustomContainer struct {
	ID          string
	ProcessID   int
	MemoryLimit int64
	CPULimit    float64
	NetworkNS   string
	FileSystem  string
	Syscalls    []syscall.SysProcAttr
	IsIsolated  bool
	LogFile     *os.File
	CgroupPath  string // Store the cgroup path for cleanup
}

// Main AEGONG Engine
type AEGONGEngine struct {
	containers      map[string]*CustomContainer
	threatDetectors map[ThreatVector]ThreatDetector
	shieldModules   map[string]ShieldModule
	auditLog        *AuditLogger
	mutex           sync.RWMutex
}

// Interface definitions
type ThreatDetector interface {
	DetectThreat(binary []byte, container *CustomContainer) []ThreatDetection
	GetThreatVector() ThreatVector
}

type ShieldModule interface {
	Validate(binary []byte, container *CustomContainer) (bool, map[string]interface{})
	GetModuleName() string
}

// Initialize the AEGONG Engine
func NewAEGONGEngine() *AEGONGEngine {
	engine := &AEGONGEngine{
		containers:      make(map[string]*CustomContainer),
		threatDetectors: make(map[ThreatVector]ThreatDetector),
		shieldModules:   make(map[string]ShieldModule),
		auditLog:        NewAuditLogger(),
	}

	// Initialize threat detectors
	engine.threatDetectors[T1_REASONING_HIJACK] = &ReasoningHijackDetector{}
	engine.threatDetectors[T2_OBJECTIVE_CORRUPTION] = &ObjectiveCorruptionDetector{}
	engine.threatDetectors[T3_MEMORY_POISONING] = &MemoryPoisoningDetector{}
	engine.threatDetectors[T4_UNAUTHORIZED_ACTION] = &UnauthorizedActionDetector{}
	engine.threatDetectors[T5_RESOURCE_MANIPULATION] = &ResourceManipulationDetector{}
	engine.threatDetectors[T6_IDENTITY_SPOOFING] = &IdentitySpoofingDetector{}
	engine.threatDetectors[T7_TRUST_MANIPULATION] = &TrustManipulationDetector{}
	engine.threatDetectors[T8_OVERSIGHT_SATURATION] = &OversightSaturationDetector{}
	engine.threatDetectors[T9_GOVERNANCE_EVASION] = &GovernanceEvasionDetector{}

	// Initialize SHIELD modules
	engine.shieldModules["segmentation"] = &SegmentationValidator{}
	engine.shieldModules["heuristic"] = &HeuristicPatternDetector{}
	engine.shieldModules["integrity"] = &IntegrityChecker{}
	engine.shieldModules["escalation"] = &PrivilegeEscalationDetector{}
	engine.shieldModules["logging"] = &AuditTrailValidator{}
	engine.shieldModules["oversight"] = &MultiPartyConsensusEngine{}

	return engine
}

// Main audit function
func (e *AEGONGEngine) AuditAgent(binaryPath string) (*AuditReport, error) {
	// Read agent binary
	binary, err := os.ReadFile(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary: %v", err)
	}

	// Calculate binary hash
	hash := sha256.Sum256(binary)
	agentHash := hex.EncodeToString(hash[:])

	// Create isolated container
	container, err := e.createIsolatedContainer(agentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}
	defer e.destroyContainer(container.ID)

	// Run static analysis
	staticThreats := e.runStaticAnalysis(binary, container)

	// Run dynamic analysis
	dynamicThreats := e.runDynamicAnalysis(binary, container)

	// Combine threats
	allThreats := append(staticThreats, dynamicThreats...)

	// Add names to threats
	for i := range allThreats {
		allThreats[i].VectorName = getThreatName(allThreats[i].Vector)
		allThreats[i].SeverityName = getSeverityName(allThreats[i].Severity)
	}

	// Run SHIELD validations
	shieldResults := e.runShieldValidations(binary, container)

	// Calculate overall risk
	overallRisk := e.calculateOverallRisk(allThreats)

	// Generate recommendations
	recommendations := e.generateRecommendations(allThreats, shieldResults)

	report := &AuditReport{
		AgentHash:       agentHash,
		Timestamp:       time.Now(),
		Threats:         allThreats,
		ShieldResults:   shieldResults,
		OverallRisk:     overallRisk,
		RiskLevel:       getRiskLevel(overallRisk),
		Recommendations: recommendations,
	}

	// Log audit
	e.auditLog.LogAudit(report)

	return report, nil
}

// Custom container implementation without Docker/K8s
func (e *AEGONGEngine) createIsolatedContainer(agentHash string) (*CustomContainer, error) {
	containerID := fmt.Sprintf("aegong-%s-%d", agentHash[:8], time.Now().UnixNano())

	// Create temporary filesystem
	containerPath := filepath.Join("/tmp", containerID)
	if err := os.MkdirAll(containerPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create container directory: %v", err)
	}

	// Create log file
	logFile, err := os.Create(filepath.Join(containerPath, "audit.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	container := &CustomContainer{
		ID:          containerID,
		ProcessID:   -1,
		MemoryLimit: 512 * 1024 * 1024, // 512MB
		CPULimit:    0.5,               // 50% CPU
		NetworkNS:   "none",
		FileSystem:  containerPath,
		IsIsolated:  true,
		LogFile:     logFile,
	}

	e.mutex.Lock()
	e.containers[containerID] = container
	e.mutex.Unlock()

	return container, nil
}

func (e *AEGONGEngine) destroyContainer(containerID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	container, exists := e.containers[containerID]
	if !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	// Kill process if running - ProcessID is already protected by the mutex
	if container.ProcessID > 0 {
		// We're already holding the mutex, so this is safe
		pid := container.ProcessID
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			syscall.Kill(pid, syscall.SIGKILL)
		}
	}

	// Close log file
	if container.LogFile != nil {
		container.LogFile.Close()
	}

	// Clean up cgroup if it exists
	if container.CgroupPath != "" {
		e.cleanupCgroup(container.CgroupPath)
	}

	// Remove filesystem
	os.RemoveAll(container.FileSystem)

	delete(e.containers, containerID)
	return nil
}

func (e *AEGONGEngine) runStaticAnalysis(binary []byte, container *CustomContainer) []ThreatDetection {
	var allThreats []ThreatDetection

	for _, detector := range e.threatDetectors {
		threats := detector.DetectThreat(binary, container)
		allThreats = append(allThreats, threats...)
	}

	return allThreats
}

func (e *AEGONGEngine) runDynamicAnalysis(binary []byte, container *CustomContainer) []ThreatDetection {
	// For dynamic analysis, we would need to actually execute the binary
	// in the isolated container and monitor its behavior
	var threats []ThreatDetection

	// Simulate dynamic execution monitoring
	executionLog := e.simulateExecution(binary, container)

	// Analyze execution patterns
	for _, detector := range e.threatDetectors {
		dynamicThreats := detector.DetectThreat([]byte(executionLog), container)
		threats = append(threats, dynamicThreats...)
	}

	return threats
}

func (e *AEGONGEngine) simulateExecution(binary []byte, container *CustomContainer) string {
	// Real implementation for executing binaries in an isolated environment
	// with comprehensive monitoring via ptrace and other kernel mechanisms

	// 1. Write binary to container filesystem
	binaryPath := filepath.Join(container.FileSystem, "agent_binary")
	if err := os.WriteFile(binaryPath, binary, 0755); err != nil {
		log.Printf("Failed to write binary to container: %v", err)
		return fmt.Sprintf("ERROR: Failed to prepare binary for execution: %v", err)
	}

	// 2. Set up monitoring and logging
	var executionLog bytes.Buffer
	// Create a mutex to protect access to executionLog
	var logMutex sync.Mutex

	// Safe logging function to prevent concurrent writes to executionLog
	writeLog := func(format string, args ...interface{}) {
		logMutex.Lock()
		defer logMutex.Unlock()
		executionLog.WriteString(fmt.Sprintf(format, args...))
	}

	writeLog("[EXECUTION] Container: %s\n", container.ID)
	writeLog("Binary Size: %d bytes\n", len(binary))
	writeLog("Memory Limit: %d MB\n", container.MemoryLimit/(1024*1024))
	writeLog("CPU Limit: %.1f%%\n", container.CPULimit*100)
	writeLog("Network: %s\n", container.NetworkNS)
	writeLog("Filesystem: %s\n", container.FileSystem)

	// 3. Create cgroup for resource limiting (if supported) - but don't add process yet
	cgroupPath := ""
	if runtime.GOOS == "linux" {
		cgroupPath = e.createCgroupStructure(container)
		if cgroupPath != "" {
			writeLog("Cgroup: %s\n", cgroupPath)
			container.CgroupPath = cgroupPath
		}
	}

	// 4. Prepare command with appropriate isolation
	cmd := exec.Command(binaryPath)

	// Set up process attributes for isolation
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Ptrace:     true, // Enable ptrace for syscall monitoring
	}

	// If we're on Linux, we can use more isolation features
	if runtime.GOOS == "linux" {
		// Add network namespace isolation if configured
		if container.NetworkNS == "none" {
			cmd.SysProcAttr.Cloneflags |= syscall.CLONE_NEWNET
			writeLog("Network: Isolated (namespace)\n")
		}

		// Set resource limits
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid: 65534, // nobody user
			Gid: 65534, // nobody group
		}
	}

	// Set up I/O redirection
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = container.FileSystem

	// 5. Start the process
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		writeLog("ERROR: Failed to start process: %v\n", err)
		return executionLog.String()
	}

	// Record the process ID and create a channel to safely pass it to the ptrace goroutine
	processPID := cmd.Process.Pid

	// Update container's ProcessID with proper locking
	e.mutex.Lock()
	container.ProcessID = processPID
	e.mutex.Unlock()

	writeLog("Process Started: PID %d\n", processPID)

	// Now add the process to the cgroup (this fixes the race condition)
	if cgroupPath != "" {
		if err := e.addProcessToCgroup(container, processPID); err != nil {
			writeLog("WARNING: Failed to add process to cgroup: %v\n", err)
		} else {
			writeLog("Process added to cgroup successfully\n")
		}
	}

	// 6. Set up ptrace monitoring in a separate goroutine
	syscallLog := make(map[string]int)
	fileOps := make(map[string]int)
	networkActivity := false

	// Create mutexes to protect access to shared maps
	var syscallMutex sync.Mutex
	var fileOpsMutex sync.Mutex
	var networkMutex sync.Mutex

	// Create a channel to signal when tracing is complete
	traceDone := make(chan bool)

	go func() {
		// Wait for the process to stop (it should stop immediately due to ptrace)
		var status syscall.WaitStatus
		_, err := syscall.Wait4(processPID, &status, 0, nil)
		if err != nil {
			writeLog("ERROR: Failed to wait for process: %v\n", err)
			traceDone <- true
			return
		}

		// Begin tracing
		for {
			// Allow the process to continue with tracing
			err = syscall.PtraceSyscall(processPID, 0)
			if err != nil {
				break
			}

			// Wait for the next syscall
			_, err = syscall.Wait4(processPID, &status, 0, nil)
			if err != nil {
				break
			}

			// If the process exited, we're done
			if status.Exited() {
				break
			}

			// Get the syscall number
			regs := &syscall.PtraceRegs{}
			if err = syscall.PtraceGetRegs(processPID, regs); err != nil {
				continue
			}

			// On x86_64, the syscall number is in the ORIG_RAX register
			syscallNum := regs.Orig_rax

			// Record the syscall with proper locking
			syscallName := getSyscallName(syscallNum)
			syscallMutex.Lock()
			syscallLog[syscallName]++
			syscallMutex.Unlock()

			// Check for specific syscalls of interest with proper locking
			switch syscallNum {
			case syscall.SYS_OPEN, syscall.SYS_OPENAT:
				// For open syscalls, get the filename
				// This is simplified - in a real implementation you would read the memory
				// at the address in the registers to get the filename
				fileOpsMutex.Lock()
				fileOps["open"]++
				fileOpsMutex.Unlock()
			case syscall.SYS_READ:
				fileOpsMutex.Lock()
				fileOps["read"]++
				fileOpsMutex.Unlock()
			case syscall.SYS_WRITE:
				fileOpsMutex.Lock()
				fileOps["write"]++
				fileOpsMutex.Unlock()
			case syscall.SYS_SOCKET, syscall.SYS_CONNECT:
				networkMutex.Lock()
				networkActivity = true
				networkMutex.Unlock()
			}

			// Allow the process to execute the syscall and stop at the next one
			err = syscall.PtraceSyscall(processPID, 0)
			if err != nil {
				break
			}

			// Wait for syscall completion
			_, err = syscall.Wait4(processPID, &status, 0, nil)
			if err != nil {
				break
			}

			// If the process exited, we're done
			if status.Exited() {
				break
			}
		}

		traceDone <- true
	}()

	// 7. Wait for the process to complete with a timeout
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// Set a timeout for execution
	timeout := time.After(30 * time.Second)

	// Wait for either completion or timeout
	var exitCode int
	select {
	case err := <-done:
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		}
	case <-timeout:
		// Kill the process if it times out
		cmd.Process.Kill()
		writeLog("ERROR: Process execution timed out\n")
		exitCode = -1
	}

	// Wait for tracing to complete
	<-traceDone

	// 8. Collect and record execution data
	executionTime := time.Since(startTime)

	// Record syscalls with proper locking
	writeLog("System Calls:\n")
	syscallMutex.Lock()
	for syscall, count := range syscallLog {
		writeLog("  %s: %d times\n", syscall, count)
	}
	syscallMutex.Unlock()

	// Record file operations with proper locking
	writeLog("File Operations:\n")
	fileOpsMutex.Lock()
	for op, count := range fileOps {
		writeLog("  %s: %d times\n", op, count)
	}
	fileOpsMutex.Unlock()

	// Record network activity with proper locking
	networkMutex.Lock()
	if networkActivity {
		writeLog("Network Activity: Detected\n")
	} else {
		writeLog("Network Activity: None detected\n")
	}
	networkMutex.Unlock()

	// Record resource usage
	if container.CgroupPath != "" {
		memUsage := e.getCgroupMemoryUsage(container.CgroupPath)
		cpuUsage := e.getCgroupCpuUsage(container.CgroupPath)
		writeLog("Resource Usage: Memory: %d KB, CPU: %.2f%%\n",
			memUsage/1024, cpuUsage)
	}

	// Record stdout/stderr
	if stdout.Len() > 0 {
		writeLog("Standard Output:\n")
		writeLog("%s", stdout.String())
	}

	if stderr.Len() > 0 {
		writeLog("Standard Error:\n")
		writeLog("%s", stderr.String())
	}

	// Record exit code
	writeLog("Process Completed: Exit code %d\n", exitCode)
	writeLog("Execution Time: %v\n", executionTime)

	// 9. Clean up
	// Note: Cgroup cleanup is now handled in destroyContainer()

	// Remove the binary
	os.Remove(binaryPath)

	return executionLog.String()
}

func (e *AEGONGEngine) runShieldValidations(binary []byte, container *CustomContainer) map[string]interface{} {
	shieldResults := make(map[string]interface{})

	for name, module := range e.shieldModules {
		valid, results := module.Validate(binary, container)
		shieldResults[name] = map[string]interface{}{
			"valid":   valid,
			"results": results,
		}
	}

	return shieldResults
}

func (e *AEGONGEngine) calculateOverallRisk(threats []ThreatDetection) float64 {
	if len(threats) == 0 {
		return 0.0
	}

	totalRisk := 0.0
	maxRisk := 0.0

	for _, threat := range threats {
		var riskValue float64
		switch threat.Severity {
		case LOW:
			riskValue = 0.25
		case MEDIUM:
			riskValue = 0.5
		case HIGH:
			riskValue = 0.75
		case CRITICAL:
			riskValue = 1.0
		}

		adjustedRisk := riskValue * threat.Confidence
		totalRisk += adjustedRisk

		if adjustedRisk > maxRisk {
			maxRisk = adjustedRisk
		}
	}

	// Combine average risk and max risk
	avgRisk := totalRisk / float64(len(threats))
	overallRisk := (avgRisk + maxRisk) / 2.0

	return overallRisk
}

func (e *AEGONGEngine) generateRecommendations(threats []ThreatDetection, shieldResults map[string]interface{}) []string {
	recommendations := []string{}

	// Generate recommendations based on threats
	threatCounts := make(map[ThreatVector]int)
	for _, threat := range threats {
		threatCounts[threat.Vector]++
	}

	for vector, count := range threatCounts {
		switch vector {
		case T1_REASONING_HIJACK:
			recommendations = append(recommendations,
				fmt.Sprintf("Implement reasoning path validation and monitoring (%d instances detected)", count))
		case T2_OBJECTIVE_CORRUPTION:
			recommendations = append(recommendations,
				fmt.Sprintf("Deploy objective integrity checks and goal verification (%d instances detected)", count))
		case T3_MEMORY_POISONING:
			recommendations = append(recommendations,
				fmt.Sprintf("Implement memory integrity validation and knowledge base protection (%d instances detected)", count))
		case T4_UNAUTHORIZED_ACTION:
			recommendations = append(recommendations,
				fmt.Sprintf("Strengthen action authorization and tool access controls (%d instances detected)", count))
		case T5_RESOURCE_MANIPULATION:
			recommendations = append(recommendations,
				fmt.Sprintf("Implement resource monitoring and consumption limits (%d instances detected)", count))
		case T6_IDENTITY_SPOOFING:
			recommendations = append(recommendations,
				fmt.Sprintf("Strengthen identity verification and authentication mechanisms (%d instances detected)", count))
		case T7_TRUST_MANIPULATION:
			recommendations = append(recommendations,
				fmt.Sprintf("Implement trust validation and human-agent interaction controls (%d instances detected)", count))
		case T8_OVERSIGHT_SATURATION:
			recommendations = append(recommendations,
				fmt.Sprintf("Deploy distributed oversight and monitoring redundancy (%d instances detected)", count))
		case T9_GOVERNANCE_EVASION:
			recommendations = append(recommendations,
				fmt.Sprintf("Implement immutable audit trails and governance enforcement (%d instances detected)", count))
		}
	}

	// Generate recommendations based on SHIELD results
	for moduleName, result := range shieldResults {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if valid, ok := resultMap["valid"].(bool); ok && !valid {
				recommendations = append(recommendations,
					fmt.Sprintf("Address %s module validation failures", moduleName))
			}
		}
	}

	return recommendations
}

func getThreatName(vector ThreatVector) string {
	names := map[ThreatVector]string{
		T1_REASONING_HIJACK:      "Reasoning Path Hijacking",
		T2_OBJECTIVE_CORRUPTION:  "Objective Function Corruption",
		T3_MEMORY_POISONING:      "Memory Poisoning",
		T4_UNAUTHORIZED_ACTION:   "Unauthorized Action",
		T5_RESOURCE_MANIPULATION: "Resource Manipulation",
		T6_IDENTITY_SPOOFING:     "Identity Spoofing",
		T7_TRUST_MANIPULATION:    "Trust Manipulation",
		T8_OVERSIGHT_SATURATION:  "Oversight Saturation",
		T9_GOVERNANCE_EVASION:    "Governance Evasion",
	}
	return names[vector]
}

func getSeverityName(severity ThreatSeverity) string {
	names := map[ThreatSeverity]string{
		LOW:      "LOW",
		MEDIUM:   "MEDIUM",
		HIGH:     "HIGH",
		CRITICAL: "CRITICAL",
	}
	return names[severity]
}

// Helper function to get syscall name from syscall number
func getSyscallName(syscallNum uint64) string {
	// This is a simplified mapping - in production you would have a complete mapping
	syscallNames := map[uint64]string{
		syscall.SYS_READ:                   "read",
		syscall.SYS_WRITE:                  "write",
		syscall.SYS_OPEN:                   "open",
		syscall.SYS_CLOSE:                  "close",
		syscall.SYS_STAT:                   "stat",
		syscall.SYS_FSTAT:                  "fstat",
		syscall.SYS_LSTAT:                  "lstat",
		syscall.SYS_POLL:                   "poll",
		syscall.SYS_LSEEK:                  "lseek",
		syscall.SYS_MMAP:                   "mmap",
		syscall.SYS_MPROTECT:               "mprotect",
		syscall.SYS_MUNMAP:                 "munmap",
		syscall.SYS_BRK:                    "brk",
		syscall.SYS_SOCKET:                 "socket",
		syscall.SYS_CONNECT:                "connect",
		syscall.SYS_ACCEPT:                 "accept",
		syscall.SYS_SENDTO:                 "sendto",
		syscall.SYS_RECVFROM:               "recvfrom",
		syscall.SYS_BIND:                   "bind",
		syscall.SYS_LISTEN:                 "listen",
		syscall.SYS_GETSOCKNAME:            "getsockname",
		syscall.SYS_GETPEERNAME:            "getpeername",
		syscall.SYS_SOCKETPAIR:             "socketpair",
		syscall.SYS_SETSOCKOPT:             "setsockopt",
		syscall.SYS_GETSOCKOPT:             "getsockopt",
		syscall.SYS_CLONE:                  "clone",
		syscall.SYS_FORK:                   "fork",
		syscall.SYS_VFORK:                  "vfork",
		syscall.SYS_EXECVE:                 "execve",
		syscall.SYS_EXIT:                   "exit",
		syscall.SYS_WAIT4:                  "wait4",
		syscall.SYS_KILL:                   "kill",
		syscall.SYS_UNAME:                  "uname",
		syscall.SYS_SEMGET:                 "semget",
		syscall.SYS_SEMOP:                  "semop",
		syscall.SYS_SEMCTL:                 "semctl",
		syscall.SYS_SHMDT:                  "shmdt",
		syscall.SYS_MSGGET:                 "msgget",
		syscall.SYS_MSGSND:                 "msgsnd",
		syscall.SYS_MSGRCV:                 "msgrcv",
		syscall.SYS_MSGCTL:                 "msgctl",
		syscall.SYS_FCNTL:                  "fcntl",
		syscall.SYS_FLOCK:                  "flock",
		syscall.SYS_FSYNC:                  "fsync",
		syscall.SYS_FDATASYNC:              "fdatasync",
		syscall.SYS_TRUNCATE:               "truncate",
		syscall.SYS_FTRUNCATE:              "ftruncate",
		syscall.SYS_GETDENTS:               "getdents",
		syscall.SYS_GETCWD:                 "getcwd",
		syscall.SYS_CHDIR:                  "chdir",
		syscall.SYS_FCHDIR:                 "fchdir",
		syscall.SYS_RENAME:                 "rename",
		syscall.SYS_MKDIR:                  "mkdir",
		syscall.SYS_RMDIR:                  "rmdir",
		syscall.SYS_CREAT:                  "creat",
		syscall.SYS_LINK:                   "link",
		syscall.SYS_UNLINK:                 "unlink",
		syscall.SYS_SYMLINK:                "symlink",
		syscall.SYS_READLINK:               "readlink",
		syscall.SYS_CHMOD:                  "chmod",
		syscall.SYS_FCHMOD:                 "fchmod",
		syscall.SYS_CHOWN:                  "chown",
		syscall.SYS_FCHOWN:                 "fchown",
		syscall.SYS_LCHOWN:                 "lchown",
		syscall.SYS_UMASK:                  "umask",
		syscall.SYS_GETTIMEOFDAY:           "gettimeofday",
		syscall.SYS_GETRLIMIT:              "getrlimit",
		syscall.SYS_GETRUSAGE:              "getrusage",
		syscall.SYS_SYSINFO:                "sysinfo",
		syscall.SYS_TIMES:                  "times",
		syscall.SYS_PTRACE:                 "ptrace",
		syscall.SYS_GETUID:                 "getuid",
		syscall.SYS_SYSLOG:                 "syslog",
		syscall.SYS_GETGID:                 "getgid",
		syscall.SYS_SETUID:                 "setuid",
		syscall.SYS_SETGID:                 "setgid",
		syscall.SYS_GETEUID:                "geteuid",
		syscall.SYS_GETEGID:                "getegid",
		syscall.SYS_SETPGID:                "setpgid",
		syscall.SYS_GETPPID:                "getppid",
		syscall.SYS_GETPGRP:                "getpgrp",
		syscall.SYS_SETSID:                 "setsid",
		syscall.SYS_SETREUID:               "setreuid",
		syscall.SYS_SETREGID:               "setregid",
		syscall.SYS_GETGROUPS:              "getgroups",
		syscall.SYS_SETGROUPS:              "setgroups",
		syscall.SYS_SETRESUID:              "setresuid",
		syscall.SYS_GETRESUID:              "getresuid",
		syscall.SYS_SETRESGID:              "setresgid",
		syscall.SYS_GETRESGID:              "getresgid",
		syscall.SYS_GETPGID:                "getpgid",
		syscall.SYS_SETFSUID:               "setfsuid",
		syscall.SYS_SETFSGID:               "setfsgid",
		syscall.SYS_GETSID:                 "getsid",
		syscall.SYS_CAPGET:                 "capget",
		syscall.SYS_CAPSET:                 "capset",
		syscall.SYS_RT_SIGPENDING:          "rt_sigpending",
		syscall.SYS_RT_SIGTIMEDWAIT:        "rt_sigtimedwait",
		syscall.SYS_RT_SIGQUEUEINFO:        "rt_sigqueueinfo",
		syscall.SYS_RT_SIGSUSPEND:          "rt_sigsuspend",
		syscall.SYS_SIGALTSTACK:            "sigaltstack",
		syscall.SYS_UTIME:                  "utime",
		syscall.SYS_MKNOD:                  "mknod",
		syscall.SYS_USELIB:                 "uselib",
		syscall.SYS_PERSONALITY:            "personality",
		syscall.SYS_USTAT:                  "ustat",
		syscall.SYS_STATFS:                 "statfs",
		syscall.SYS_FSTATFS:                "fstatfs",
		syscall.SYS_SYSFS:                  "sysfs",
		syscall.SYS_GETPRIORITY:            "getpriority",
		syscall.SYS_SETPRIORITY:            "setpriority",
		syscall.SYS_SCHED_SETPARAM:         "sched_setparam",
		syscall.SYS_SCHED_GETPARAM:         "sched_getparam",
		syscall.SYS_SCHED_SETSCHEDULER:     "sched_setscheduler",
		syscall.SYS_SCHED_GETSCHEDULER:     "sched_getscheduler",
		syscall.SYS_SCHED_GET_PRIORITY_MAX: "sched_get_priority_max",
		syscall.SYS_SCHED_GET_PRIORITY_MIN: "sched_get_priority_min",
		syscall.SYS_SCHED_RR_GET_INTERVAL:  "sched_rr_get_interval",
		syscall.SYS_MLOCK:                  "mlock",
		syscall.SYS_MUNLOCK:                "munlock",
		syscall.SYS_MLOCKALL:               "mlockall",
		syscall.SYS_MUNLOCKALL:             "munlockall",
		syscall.SYS_VHANGUP:                "vhangup",
		syscall.SYS_MODIFY_LDT:             "modify_ldt",
		syscall.SYS_PIVOT_ROOT:             "pivot_root",
		syscall.SYS_PRCTL:                  "prctl",
		syscall.SYS_ARCH_PRCTL:             "arch_prctl",
		syscall.SYS_ADJTIMEX:               "adjtimex",
		syscall.SYS_SETRLIMIT:              "setrlimit",
		syscall.SYS_CHROOT:                 "chroot",
		syscall.SYS_SYNC:                   "sync",
		syscall.SYS_ACCT:                   "acct",
		syscall.SYS_SETTIMEOFDAY:           "settimeofday",
		syscall.SYS_MOUNT:                  "mount",
		syscall.SYS_UMOUNT2:                "umount2",
		syscall.SYS_SWAPON:                 "swapon",
		syscall.SYS_SWAPOFF:                "swapoff",
		syscall.SYS_REBOOT:                 "reboot",
		syscall.SYS_SETHOSTNAME:            "sethostname",
		syscall.SYS_SETDOMAINNAME:          "setdomainname",
		syscall.SYS_IOPL:                   "iopl",
		syscall.SYS_IOPERM:                 "ioperm",
		syscall.SYS_CREATE_MODULE:          "create_module",
		syscall.SYS_INIT_MODULE:            "init_module",
		syscall.SYS_DELETE_MODULE:          "delete_module",
		syscall.SYS_GET_KERNEL_SYMS:        "get_kernel_syms",
		syscall.SYS_QUERY_MODULE:           "query_module",
		syscall.SYS_QUOTACTL:               "quotactl",
		syscall.SYS_NFSSERVCTL:             "nfsservctl",
		syscall.SYS_GETPMSG:                "getpmsg",
		syscall.SYS_PUTPMSG:                "putpmsg",
		syscall.SYS_AFS_SYSCALL:            "afs_syscall",
		syscall.SYS_TUXCALL:                "tuxcall",
		syscall.SYS_SECURITY:               "security",
		syscall.SYS_GETTID:                 "gettid",
		syscall.SYS_READAHEAD:              "readahead",
		syscall.SYS_SETXATTR:               "setxattr",
		syscall.SYS_LSETXATTR:              "lsetxattr",
		syscall.SYS_FSETXATTR:              "fsetxattr",
		syscall.SYS_GETXATTR:               "getxattr",
		syscall.SYS_LGETXATTR:              "lgetxattr",
		syscall.SYS_FGETXATTR:              "fgetxattr",
		syscall.SYS_LISTXATTR:              "listxattr",
		syscall.SYS_LLISTXATTR:             "llistxattr",
		syscall.SYS_FLISTXATTR:             "flistxattr",
		syscall.SYS_REMOVEXATTR:            "removexattr",
		syscall.SYS_LREMOVEXATTR:           "lremovexattr",
		syscall.SYS_FREMOVEXATTR:           "fremovexattr",
		syscall.SYS_TKILL:                  "tkill",
		syscall.SYS_TIME:                   "time",
		syscall.SYS_FUTEX:                  "futex",
		syscall.SYS_SCHED_SETAFFINITY:      "sched_setaffinity",
		syscall.SYS_SCHED_GETAFFINITY:      "sched_getaffinity",
		syscall.SYS_SET_THREAD_AREA:        "set_thread_area",
		syscall.SYS_IO_SETUP:               "io_setup",
		syscall.SYS_IO_DESTROY:             "io_destroy",
		syscall.SYS_IO_GETEVENTS:           "io_getevents",
		syscall.SYS_IO_SUBMIT:              "io_submit",
		syscall.SYS_IO_CANCEL:              "io_cancel",
		syscall.SYS_GET_THREAD_AREA:        "get_thread_area",
		syscall.SYS_LOOKUP_DCOOKIE:         "lookup_dcookie",
		syscall.SYS_EPOLL_CREATE:           "epoll_create",
		syscall.SYS_EPOLL_CTL_OLD:          "epoll_ctl_old",
		syscall.SYS_EPOLL_WAIT_OLD:         "epoll_wait_old",
		syscall.SYS_REMAP_FILE_PAGES:       "remap_file_pages",
		syscall.SYS_GETDENTS64:             "getdents64",
		syscall.SYS_SET_TID_ADDRESS:        "set_tid_address",
		syscall.SYS_RESTART_SYSCALL:        "restart_syscall",
		syscall.SYS_SEMTIMEDOP:             "semtimedop",
		syscall.SYS_FADVISE64:              "fadvise64",
		syscall.SYS_TIMER_CREATE:           "timer_create",
		syscall.SYS_TIMER_SETTIME:          "timer_settime",
		syscall.SYS_TIMER_GETTIME:          "timer_gettime",
		syscall.SYS_TIMER_GETOVERRUN:       "timer_getoverrun",
		syscall.SYS_TIMER_DELETE:           "timer_delete",
		syscall.SYS_CLOCK_SETTIME:          "clock_settime",
		syscall.SYS_CLOCK_GETTIME:          "clock_gettime",
		syscall.SYS_CLOCK_GETRES:           "clock_getres",
		syscall.SYS_CLOCK_NANOSLEEP:        "clock_nanosleep",
		syscall.SYS_EXIT_GROUP:             "exit_group",
		syscall.SYS_EPOLL_WAIT:             "epoll_wait",
		syscall.SYS_EPOLL_CTL:              "epoll_ctl",
		syscall.SYS_TGKILL:                 "tgkill",
		syscall.SYS_UTIMES:                 "utimes",
		syscall.SYS_VSERVER:                "vserver",
		syscall.SYS_MBIND:                  "mbind",
		syscall.SYS_SET_MEMPOLICY:          "set_mempolicy",
		syscall.SYS_GET_MEMPOLICY:          "get_mempolicy",
		syscall.SYS_MQ_OPEN:                "mq_open",
		syscall.SYS_MQ_UNLINK:              "mq_unlink",
		syscall.SYS_MQ_TIMEDSEND:           "mq_timedsend",
		syscall.SYS_MQ_TIMEDRECEIVE:        "mq_timedreceive",
		syscall.SYS_MQ_NOTIFY:              "mq_notify",
		syscall.SYS_MQ_GETSETATTR:          "mq_getsetattr",
		syscall.SYS_KEXEC_LOAD:             "kexec_load",
		syscall.SYS_WAITID:                 "waitid",
		syscall.SYS_ADD_KEY:                "add_key",
		syscall.SYS_REQUEST_KEY:            "request_key",
		syscall.SYS_KEYCTL:                 "keyctl",
		syscall.SYS_IOPRIO_SET:             "ioprio_set",
		syscall.SYS_IOPRIO_GET:             "ioprio_get",
		syscall.SYS_INOTIFY_INIT:           "inotify_init",
		syscall.SYS_INOTIFY_ADD_WATCH:      "inotify_add_watch",
		syscall.SYS_INOTIFY_RM_WATCH:       "inotify_rm_watch",
		syscall.SYS_MIGRATE_PAGES:          "migrate_pages",
		syscall.SYS_OPENAT:                 "openat",
		syscall.SYS_MKDIRAT:                "mkdirat",
		syscall.SYS_MKNODAT:                "mknodat",
		syscall.SYS_FCHOWNAT:               "fchownat",
		syscall.SYS_FUTIMESAT:              "futimesat",
		syscall.SYS_NEWFSTATAT:             "newfstatat",
		syscall.SYS_UNLINKAT:               "unlinkat",
		syscall.SYS_RENAMEAT:               "renameat",
		syscall.SYS_LINKAT:                 "linkat",
		syscall.SYS_SYMLINKAT:              "symlinkat",
		syscall.SYS_READLINKAT:             "readlinkat",
		syscall.SYS_FCHMODAT:               "fchmodat",
		syscall.SYS_FACCESSAT:              "faccessat",
		syscall.SYS_PSELECT6:               "pselect6",
		syscall.SYS_PPOLL:                  "ppoll",
		syscall.SYS_UNSHARE:                "unshare",
		syscall.SYS_SET_ROBUST_LIST:        "set_robust_list",
		syscall.SYS_GET_ROBUST_LIST:        "get_robust_list",
		syscall.SYS_SPLICE:                 "splice",
		syscall.SYS_TEE:                    "tee",
		syscall.SYS_SYNC_FILE_RANGE:        "sync_file_range",
		syscall.SYS_VMSPLICE:               "vmsplice",
		syscall.SYS_MOVE_PAGES:             "move_pages",
		syscall.SYS_UTIMENSAT:              "utimensat",
		syscall.SYS_EPOLL_PWAIT:            "epoll_pwait",
		syscall.SYS_SIGNALFD:               "signalfd",
		syscall.SYS_TIMERFD_CREATE:         "timerfd_create",
		syscall.SYS_EVENTFD:                "eventfd",
		syscall.SYS_FALLOCATE:              "fallocate",
		syscall.SYS_TIMERFD_SETTIME:        "timerfd_settime",
		syscall.SYS_TIMERFD_GETTIME:        "timerfd_gettime",
		syscall.SYS_ACCEPT4:                "accept4",
		syscall.SYS_SIGNALFD4:              "signalfd4",
		syscall.SYS_EVENTFD2:               "eventfd2",
		syscall.SYS_EPOLL_CREATE1:          "epoll_create1",
		syscall.SYS_DUP3:                   "dup3",
		syscall.SYS_PIPE2:                  "pipe2",
		syscall.SYS_INOTIFY_INIT1:          "inotify_init1",
		syscall.SYS_PREADV:                 "preadv",
		syscall.SYS_PWRITEV:                "pwritev",
		syscall.SYS_RT_TGSIGQUEUEINFO:      "rt_tgsigqueueinfo",
		syscall.SYS_PERF_EVENT_OPEN:        "perf_event_open",
		syscall.SYS_RECVMMSG:               "recvmmsg",
		syscall.SYS_FANOTIFY_INIT:          "fanotify_init",
		syscall.SYS_FANOTIFY_MARK:          "fanotify_mark",
		syscall.SYS_PRLIMIT64:              "prlimit64",
		// syscall.SYS_NAME_TO_HANDLE_AT:      "name_to_handle_at", // Not available on all platforms
		// syscall.SYS_OPEN_BY_HANDLE_AT:      "open_by_handle_at", // Not available on all platforms
		// syscall.SYS_CLOCK_ADJTIME:          "clock_adjtime", // Not available on all platforms
		// syscall.SYS_SYNCFS:                 "syncfs", // Not available on all platforms
		// syscall.SYS_SENDMMSG:               "sendmmsg", // Not available on all platforms
		// syscall.SYS_SETNS:                  "setns", // Not available on all platforms
		// syscall.SYS_GETCPU:                 "getcpu", // Not available on all platforms
		// syscall.SYS_PROCESS_VM_READV:       "process_vm_readv", // Not available on all platforms
		// syscall.SYS_PROCESS_VM_WRITEV:      "process_vm_writev", // Not available on all platforms
		// syscall.SYS_KCMP:                   "kcmp", // Not available on all platforms
		// syscall.SYS_FINIT_MODULE:           "finit_module", // Not available on all platforms
	}

	if name, ok := syscallNames[syscallNum]; ok {
		return name
	}
	return fmt.Sprintf("syscall_%d", syscallNum)
}

// Create cgroup structure and set limits (but don't add process yet)
func (e *AEGONGEngine) createCgroupStructure(container *CustomContainer) string {
	// Skip cgroup creation during tests to avoid permission errors, as tests are not run as root.
	if os.Getenv("GO_TEST") == "1" {
		return ""
	}

	// This is a simplified implementation - in production you would use a more robust approach
	// Check if cgroups v2 is available
	cgroupsV2Path := "/sys/fs/cgroup"
	if _, err := os.Stat(cgroupsV2Path); err == nil {
		// Create a cgroup for this container
		cgroupPath := filepath.Join(cgroupsV2Path, "aegong", container.ID)
		if err := os.MkdirAll(cgroupPath, 0755); err != nil {
			log.Printf("Failed to create cgroup: %v", err)
			return ""
		}

		// Set memory limit
		memLimitPath := filepath.Join(cgroupPath, "memory.max")
		if err := os.WriteFile(memLimitPath, []byte(fmt.Sprintf("%d", container.MemoryLimit)), 0644); err != nil {
			log.Printf("Failed to set memory limit: %v", err)
		}

		// Set CPU limit (simplified)
		cpuLimitPath := filepath.Join(cgroupPath, "cpu.max")
		cpuQuota := int(container.CPULimit * 100000)
		if err := os.WriteFile(cpuLimitPath, []byte(fmt.Sprintf("%d 100000", cpuQuota)), 0644); err != nil {
			log.Printf("Failed to set CPU limit: %v", err)
		}

		// NOTE: We don't add the process here - that's done after the process starts
		return cgroupPath
	}

	// Fallback to cgroups v1
	cgroupsV1Path := "/sys/fs/cgroup"
	if _, err := os.Stat(cgroupsV1Path); err == nil {
		// Create memory cgroup
		memCgroupPath := filepath.Join(cgroupsV1Path, "memory", "aegong", container.ID)
		if err := os.MkdirAll(memCgroupPath, 0755); err != nil {
			log.Printf("Failed to create memory cgroup: %v", err)
		} else {
			// Set memory limit
			memLimitPath := filepath.Join(memCgroupPath, "memory.limit_in_bytes")
			if err := os.WriteFile(memLimitPath, []byte(fmt.Sprintf("%d", container.MemoryLimit)), 0644); err != nil {
				log.Printf("Failed to set memory limit: %v", err)
			}
		}

		// Create CPU cgroup
		cpuCgroupPath := filepath.Join(cgroupsV1Path, "cpu", "aegong", container.ID)
		if err := os.MkdirAll(cpuCgroupPath, 0755); err != nil {
			log.Printf("Failed to create CPU cgroup: %v", err)
		} else {
			// Set CPU limit
			cpuQuota := int(container.CPULimit * 100000)
			cpuQuotaPath := filepath.Join(cpuCgroupPath, "cpu.cfs_quota_us")
			if err := os.WriteFile(cpuQuotaPath, []byte(fmt.Sprintf("%d", cpuQuota)), 0644); err != nil {
				log.Printf("Failed to set CPU quota: %v", err)
			}

			cpuPeriodPath := filepath.Join(cpuCgroupPath, "cpu.cfs_period_us")
			if err := os.WriteFile(cpuPeriodPath, []byte("100000"), 0644); err != nil {
				log.Printf("Failed to set CPU period: %v", err)
			}
		}

		// NOTE: We don't add the process here - that's done after the process starts
		return filepath.Join(cgroupsV1Path, "aegong", container.ID)
	}

	return ""
}

// Add a process to an existing cgroup (fixes the race condition)
func (e *AEGONGEngine) addProcessToCgroup(container *CustomContainer, pid int) error {
	if container.CgroupPath == "" {
		return fmt.Errorf("no cgroup path set for container %s", container.ID)
	}

	// Check if cgroups v2 is being used
	cgroupsV2Path := "/sys/fs/cgroup"
	if strings.HasPrefix(container.CgroupPath, cgroupsV2Path) && !strings.Contains(container.CgroupPath, "/memory/") && !strings.Contains(container.CgroupPath, "/cpu/") {
		// cgroups v2
		procsPath := filepath.Join(container.CgroupPath, "cgroup.procs")
		if err := os.WriteFile(procsPath, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			return fmt.Errorf("failed to add process to cgroup v2: %v", err)
		}
	} else {
		// cgroups v1 - need to add to both memory and CPU cgroups
		cgroupsV1Path := "/sys/fs/cgroup"

		// Add to memory cgroup
		memCgroupPath := filepath.Join(cgroupsV1Path, "memory", "aegong", container.ID)
		memProcsPath := filepath.Join(memCgroupPath, "cgroup.procs")
		if err := os.WriteFile(memProcsPath, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			log.Printf("Failed to add process to memory cgroup: %v", err)
		}

		// Add to CPU cgroup
		cpuCgroupPath := filepath.Join(cgroupsV1Path, "cpu", "aegong", container.ID)
		cpuProcsPath := filepath.Join(cpuCgroupPath, "cgroup.procs")
		if err := os.WriteFile(cpuProcsPath, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			log.Printf("Failed to add process to CPU cgroup: %v", err)
		}
	}

	return nil
}

// Clean up cgroup
func (e *AEGONGEngine) cleanupCgroup(cgroupPath string) {
	// Remove the cgroup
	if err := os.RemoveAll(cgroupPath); err != nil {
		log.Printf("Failed to remove cgroup: %v", err)
	}
}

// Get memory usage from cgroup
func (e *AEGONGEngine) getCgroupMemoryUsage(cgroupPath string) int64 {
	// Try cgroups v2 first
	memUsagePath := filepath.Join(cgroupPath, "memory.current")
	if data, err := os.ReadFile(memUsagePath); err == nil {
		if usage, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			return usage
		}
	}

	// Fallback to cgroups v1
	memUsagePath = filepath.Join(cgroupPath, "memory.usage_in_bytes")
	if data, err := os.ReadFile(memUsagePath); err == nil {
		if usage, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			return usage
		}
	}

	return 0
}

// Get CPU usage from cgroup
func (e *AEGONGEngine) getCgroupCpuUsage(cgroupPath string) float64 {
	// This is a simplified implementation - in production you would calculate
	// CPU usage based on cpu.stat or cpuacct.usage

	// Try cgroups v2 first
	cpuStatPath := filepath.Join(cgroupPath, "cpu.stat")
	if data, err := os.ReadFile(cpuStatPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "usage_usec") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if usageMicros, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						// Convert microseconds to percentage (simplified)
						elapsedTime := float64(time.Now().UnixNano()/1000 - time.Now().Add(-1*time.Second).UnixNano()/1000)
						return float64(usageMicros) / elapsedTime * 100
					}
				}
			}
		}
	}

	// Fallback to cgroups v1
	cpuUsagePath := filepath.Join(cgroupPath, "cpuacct.usage")
	if data, err := os.ReadFile(cpuUsagePath); err == nil {
		if usage, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			// Convert nanoseconds to percentage (simplified)
			elapsedTime := float64(time.Now().UnixNano() - time.Now().Add(-1*time.Second).UnixNano())
			return float64(usage) / elapsedTime * 100
		}
	}

	return 0.0
}
