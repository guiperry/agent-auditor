class AegongInterface {
    constructor() {
        this.ws = null;
        this.currentReport = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connectWebSocket();
        this.loadHistory();
        this.updateStatus('Ready', 'ready');
    }

    setupEventListeners() {
        // File upload
        const fileInput = document.getElementById('fileInput');
        const uploadArea = document.getElementById('uploadArea');
        const uploadBtn = document.getElementById('uploadBtn');

        uploadArea.addEventListener('click', () => fileInput.click());
        uploadArea.addEventListener('dragover', this.handleDragOver.bind(this));
        uploadArea.addEventListener('dragleave', this.handleDragLeave.bind(this));
        uploadArea.addEventListener('drop', this.handleDrop.bind(this));

        fileInput.addEventListener('change', this.handleFileSelect.bind(this));
        uploadBtn.addEventListener('click', this.uploadFile.bind(this));

        // Tab switching
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', () => this.switchTab(btn.dataset.tab));
        });

        // Action buttons
        document.getElementById('newAnalysisBtn').addEventListener('click', this.resetInterface.bind(this));
        document.getElementById('downloadReportBtn').addEventListener('click', this.downloadReport.bind(this));
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleWebSocketMessage(message);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            setTimeout(() => this.connectWebSocket(), 5000);
        };
    }

    handleWebSocketMessage(message) {
        if (message.type === 'aegong_message') {
            this.showAegongMessage(message.message);
        }
    }

    showAegongMessage(message) {
        // Could be used for real-time updates during analysis
        console.log('Aegong says:', message);
    }

    handleDragOver(e) {
        e.preventDefault();
        document.getElementById('uploadArea').classList.add('dragover');
    }

    handleDragLeave(e) {
        e.preventDefault();
        document.getElementById('uploadArea').classList.remove('dragover');
    }

    handleDrop(e) {
        e.preventDefault();
        document.getElementById('uploadArea').classList.remove('dragover');
        
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            document.getElementById('fileInput').files = files;
            this.handleFileSelect();
        }
    }

    handleFileSelect() {
        const fileInput = document.getElementById('fileInput');
        const uploadBtn = document.getElementById('uploadBtn');
        
        if (fileInput.files.length > 0) {
            const file = fileInput.files[0];
            uploadBtn.disabled = false;
            uploadBtn.querySelector('.btn-text').textContent = `Upload ${file.name}`;
        }
    }

    async uploadFile() {
        const fileInput = document.getElementById('fileInput');
        const uploadBtn = document.getElementById('uploadBtn');
        
        if (fileInput.files.length === 0) return;

        const file = fileInput.files[0];
        const formData = new FormData();
        formData.append('agent', file);

        // Show loading state
        uploadBtn.disabled = true;
        uploadBtn.querySelector('.btn-text').style.display = 'none';
        uploadBtn.querySelector('.btn-loader').hidden = false;

        try {
            const response = await fetch('/api/upload', {
                method: 'POST',
                body: formData
            });

            const result = await response.json();
            
            if (response.ok) {
                this.startAnalysis(result.filename);
            } else {
                throw new Error(result.error || 'Upload failed');
            }
        } catch (error) {
            console.error('Upload error:', error);
            this.updateStatus('Upload failed', 'error');
            
            // Reset button
            uploadBtn.disabled = false;
            uploadBtn.querySelector('.btn-text').style.display = 'inline';
            uploadBtn.querySelector('.btn-loader').hidden = true;
        }
    }

    async startAnalysis(filename) {
        // Hide upload section and show analysis
        document.getElementById('uploadSection').hidden = true;
        document.getElementById('analysisSection').hidden = false;

        this.updateStatus('Analyzing...', 'analyzing');
        
        // Simulate progress
        this.simulateProgress();

        try {
            const response = await fetch(`/api/audit/${filename}`, {
                method: 'POST'
            });

            const result = await response.json();
            
            if (response.ok) {
                this.currentReport = result;
                this.showResults(result);
                this.loadHistory(); // Refresh history
            } else {
                // Check if this is a "not an agent" error
                if (result.error === "Not an AI agent" && result.validation) {
                    this.showNotAgentError(result.validation, filename);
                } else {
                    throw new Error(result.error || result.message || 'Analysis failed');
                }
            }
        } catch (error) {
            console.error('Analysis error:', error);
            this.updateStatus('Analysis failed', 'error');
            
            // Return to upload screen with error message
            setTimeout(() => {
                this.resetInterface();
                this.showErrorMessage('Analysis failed: ' + error.message);
            }, 1000);
        }
    }
    
    showNotAgentError(validation, filename) {
        // Stop the progress animation
        document.getElementById('progressFill').style.width = '100%';
        document.getElementById('analysisStatus').textContent = 'Validation failed';
        
        // Create error message
        const errorSection = document.createElement('div');
        errorSection.className = 'error-section section-card fade-in';
        errorSection.innerHTML = `
            <h3>⚠️ Not an AI Agent</h3>
            <p>The uploaded file "${filename}" does not appear to be an AI agent based on our validation criteria.</p>
            
            <div class="validation-details">
                <p><strong>File Type:</strong> ${validation.agent_type.toUpperCase()}</p>
                <p><strong>Confidence:</strong> ${Math.round(validation.confidence * 100)}%</p>
                
                <div class="validation-capabilities">
                    <strong>Detected Capabilities:</strong>
                    ${validation.capabilities.length > 0 ? 
                        `<ul>${validation.capabilities.map(cap => `<li>${this.formatCapabilityName(cap)}</li>`).join('')}</ul>` : 
                        '<p>No agent capabilities detected</p>'}
                </div>
                
                <div class="validation-reasons">
                    <strong>Validation Notes:</strong>
                    <ul>
                        ${validation.reasons.map(reason => `<li>${reason}</li>`).join('')}
                    </ul>
                </div>
                
                <p class="validation-explanation">
                    <strong>Why this matters:</strong> Aegong can only analyze AI agents that have the minimum required capabilities:
                    perception (input), action (output), and either reasoning (decision-making) or memory (state).
                </p>
            </div>
            
            <button id="returnToUploadBtn" class="btn primary-btn">Upload a Different File</button>
        `;
        
        // Replace analysis section with error
        const analysisSection = document.getElementById('analysisSection');
        analysisSection.innerHTML = '';
        analysisSection.appendChild(errorSection);
        
        // Add button event listener
        document.getElementById('returnToUploadBtn').addEventListener('click', () => {
            this.resetInterface();
        });
        
        // Update status
        this.updateStatus('Validation failed', 'error');
    }
    
    showErrorMessage(message) {
        const uploadArea = document.getElementById('uploadArea');
        const errorMsg = document.createElement('div');
        errorMsg.className = 'upload-error';
        errorMsg.textContent = message;
        
        // Remove any existing error messages
        const existingError = uploadArea.querySelector('.upload-error');
        if (existingError) {
            existingError.remove();
        }
        
        // Add the error message
        uploadArea.appendChild(errorMsg);
        
        // Add some basic styles if they don't exist
        if (!document.getElementById('errorStyles')) {
            const styleEl = document.createElement('style');
            styleEl.id = 'errorStyles';
            styleEl.textContent = `
                .upload-error {
                    color: #e74c3c;
                    background-color: rgba(231, 76, 60, 0.1);
                    border: 1px solid #e74c3c;
                    border-radius: 4px;
                    padding: 10px;
                    margin-top: 15px;
                    text-align: center;
                }
            `;
            document.head.appendChild(styleEl);
        }
    }

    simulateProgress() {
        const progressFill = document.getElementById('progressFill');
        const analysisStatus = document.getElementById('analysisStatus');
        
        const steps = [
            { progress: 5, text: 'Aegong is awakening his sensors...' },
            { progress: 15, text: 'Validating agent capabilities...' },
            { progress: 25, text: 'Scanning binary structure...' },
            { progress: 40, text: 'Analyzing threat vectors...' },
            { progress: 60, text: 'Running SHIELD validations...' },
            { progress: 80, text: 'Calculating risk assessment...' },
            { progress: 95, text: 'Preparing Aegong\'s verdict...' },
            { progress: 100, text: 'Analysis complete!' }
        ];

        let currentStep = 0;
        const interval = setInterval(() => {
            if (currentStep < steps.length) {
                const step = steps[currentStep];
                progressFill.style.width = `${step.progress}%`;
                analysisStatus.textContent = step.text;
                currentStep++;
            } else {
                clearInterval(interval);
            }
        }, 700);
    }

    showResults(report) {
        // Hide analysis section and show results
        document.getElementById('analysisSection').hidden = true;
        document.getElementById('resultsSection').hidden = false;

        this.updateStatus('Analysis complete', 'complete');
        
        // Store the current report
        this.currentReport = report;

        // Populate overview
        document.getElementById('agentName').textContent = report.agent_name || 'Unknown Agent';
        document.getElementById('threatCount').textContent = report.threats.length;
        document.getElementById('riskScore').textContent = `${Math.round(report.overall_risk * 100)}%`;
        
        // Update risk badge
        const riskBadge = document.getElementById('riskBadge');
        const riskLevel = document.getElementById('riskLevel');
        riskLevel.textContent = report.risk_level;
        riskBadge.className = `risk-badge risk-${report.risk_level.toLowerCase()}`;

        // Show Aegong's message
        document.getElementById('aegongText').textContent = report.aegong_message;

        // Populate threats
        this.populateThreats(report.threats);

        // Populate SHIELD results
        this.populateShields(report.shield_results);

        // Populate recommendations
        this.populateRecommendations(report.recommendations);

        // Update SHIELD score
        const shieldPassed = this.countPassedShields(report.shield_results);
        document.getElementById('shieldScore').textContent = `${shieldPassed}/6`;
        
        // Display validation results if available
        if (report.details && report.details.validation) {
            this.displayValidationResults(report.details.validation);
        }
        
        // Add report hash to the results container for voice player
        const resultsSection = document.getElementById('resultsSection');
        resultsSection.classList.add('report-container');
        resultsSection.dataset.reportHash = report.agent_hash;
        
        // Trigger AEGONG voice playback if available
        if (window.aegongVoice && typeof window.aegongVoice.playReport === 'function') {
            try {
                // First check if the voice report file exists
                fetch(`/voice_reports/aegong_report_${report.agent_hash.substring(0, 8)}.wav`, { method: 'HEAD' })
                    .then(response => {
                        if (response.ok) {
                            // Voice report exists, play it
                            window.aegongVoice.playReport(report.agent_hash);
                        } else {
                            // Voice report doesn't exist, don't try to play it
                            console.log('Voice report not available for this agent');
                        }
                    })
                    .catch(error => {
                        console.log('Error checking for voice report:', error);
                    });
            } catch (error) {
                console.warn('Error playing voice report:', error);
            }
        }
    }
    
    displayValidationResults(validation) {
        // Check if we have a validation section in the HTML
        let validationSection = document.getElementById('validationSection');
        
        // If not, create it
        if (!validationSection) {
            // Find the overview section to insert after it
            const overviewSection = document.querySelector('.overview-section');
            
            // Create validation section
            validationSection = document.createElement('div');
            validationSection.id = 'validationSection';
            validationSection.className = 'validation-section section-card fade-in';
            
            // Insert after overview
            if (overviewSection && overviewSection.parentNode) {
                overviewSection.parentNode.insertBefore(validationSection, overviewSection.nextSibling);
            }
        }
        
        // Format confidence as percentage
        const confidencePercent = Math.round(validation.confidence * 100);
        
        // Create HTML content
        let html = `
            <h3>Agent Validation</h3>
            <div class="validation-details">
                <p><strong>Agent Type:</strong> ${validation.agent_type.toUpperCase()}</p>
                <p><strong>Confidence:</strong> ${confidencePercent}%</p>
                <div class="validation-capabilities">
                    <strong>Detected Capabilities:</strong>
                    <ul>
        `;
        
        // Add capabilities
        validation.capabilities.forEach(capability => {
            html += `<li>${this.formatCapabilityName(capability)}</li>`;
        });
        
        html += `
                    </ul>
                </div>
                <div class="validation-reasons">
                    <strong>Validation Notes:</strong>
                    <ul>
        `;
        
        // Add reasons
        validation.reasons.forEach(reason => {
            html += `<li>${reason}</li>`;
        });
        
        html += `
                    </ul>
                </div>
            </div>
        `;
        
        // Set the HTML content
        validationSection.innerHTML = html;
        
        // Add some basic styles if they don't exist
        if (!document.getElementById('validationStyles')) {
            const styleEl = document.createElement('style');
            styleEl.id = 'validationStyles';
            styleEl.textContent = `
                .validation-section {
                    margin-top: 20px;
                    padding: 20px;
                }
                .validation-capabilities, .validation-reasons {
                    margin-top: 10px;
                }
                .validation-capabilities ul, .validation-reasons ul {
                    margin-top: 5px;
                    padding-left: 20px;
                }
            `;
            document.head.appendChild(styleEl);
        }
    }
    
    formatCapabilityName(capability) {
        // Format capability names for display
        switch(capability) {
            case 'perception':
                return 'Perception (Input/Sensing)';
            case 'action':
                return 'Action (Output/Response)';
            case 'reasoning':
                return 'Reasoning (Decision-making)';
            case 'memory':
                return 'Memory (State Management)';
            case 'autonomy':
                return 'Autonomy (Independent Operation)';
            case 'ai_libraries':
                return 'AI/ML Libraries';
            case 'agent_class':
                return 'Agent-specific Components';
            default:
                return capability.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
        }
    }

    populateThreats(threats) {
        const threatsList = document.getElementById('threatsList');
        threatsList.innerHTML = '';

        if (threats.length === 0) {
            threatsList.innerHTML = '<p class="no-threats">🎉 No threats detected! Aegong is pleased.</p>';
            return;
        }

        threats.forEach(threat => {
            const threatItem = document.createElement('div');
            threatItem.className = 'threat-item fade-in';
            
            threatItem.innerHTML = `
                <div class="threat-header">
                    <div class="threat-title">${threat.vector_name}</div>
                    <div class="severity-badge severity-${threat.severity_name.toLowerCase()}">
                        ${threat.severity_name}
                    </div>
                </div>
                <p><strong>Confidence:</strong> ${Math.round(threat.confidence * 100)}%</p>
                <div class="threat-evidence">
                    <strong>Evidence:</strong>
                    <ul>
                        ${threat.evidence.map(e => `<li>${e}</li>`).join('')}
                    </ul>
                </div>
            `;
            
            threatsList.appendChild(threatItem);
        });
    }

    populateShields(shieldResults) {
        const shieldsGrid = document.getElementById('shieldsGrid');
        shieldsGrid.innerHTML = '';

        Object.entries(shieldResults).forEach(([name, result]) => {
            const shieldItem = document.createElement('div');
            shieldItem.className = 'shield-item fade-in';
            
            const isValid = result.valid;
            const statusClass = isValid ? 'shield-pass' : 'shield-fail';
            const statusText = isValid ? 'PASS' : 'FAIL';
            
            shieldItem.innerHTML = `
                <div class="shield-header">
                    <div class="shield-name">${name}</div>
                    <div class="shield-status ${statusClass}">${statusText}</div>
                </div>
                <div class="shield-details">
                    ${this.formatShieldDetails(result.results)}
                </div>
            `;
            
            shieldsGrid.appendChild(shieldItem);
        });
    }

    formatShieldDetails(details) {
        return Object.entries(details)
            .map(([key, value]) => {
                const formattedKey = key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
                let formattedValue = value;
                
                if (typeof value === 'number') {
                    formattedValue = value.toFixed(2);
                } else if (typeof value === 'boolean') {
                    formattedValue = value ? '✅' : '❌';
                }
                
                return `<p><strong>${formattedKey}:</strong> ${formattedValue}</p>`;
            })
            .join('');
    }

    populateRecommendations(recommendations) {
        const recommendationsList = document.getElementById('recommendationsList');
        recommendationsList.innerHTML = '';

        if (recommendations.length === 0) {
            recommendationsList.innerHTML = '<p class="no-recommendations">🎯 No specific recommendations. Agent appears secure!</p>';
            return;
        }

        recommendations.forEach(rec => {
            const recItem = document.createElement('div');
            recItem.className = 'recommendation-item fade-in';
            recItem.textContent = rec;
            recommendationsList.appendChild(recItem);
        });
    }

    countPassedShields(shieldResults) {
        return Object.values(shieldResults).filter(result => result.valid).length;
    }

    switchTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabName);
        });

        // Update tab panes
        document.querySelectorAll('.tab-pane').forEach(pane => {
            pane.classList.toggle('active', pane.id === tabName);
        });
    }

    async loadHistory() {
        try {
            const response = await fetch('/api/reports');
            const reports = await response.json();
            
            const historyList = document.getElementById('historyList');
            historyList.innerHTML = '';

            if (reports.length === 0) {
                historyList.innerHTML = '<p class="no-history">📝 No previous audits found. Upload an agent to begin!</p>';
                return;
            }

            reports.reverse().forEach(report => {
                const historyItem = document.createElement('div');
                historyItem.className = 'history-item';
                historyItem.onclick = () => this.loadReport(report.hash);
                
                const date = new Date(report.timestamp).toLocaleDateString();
                const riskClass = `risk-${report.risk_level.toLowerCase()}`;
                
                historyItem.innerHTML = `
                    <div class="history-info">
                        <h4>${report.agent_name}</h4>
                        <p>${report.threat_count} threats detected</p>
                    </div>
                    <div class="history-meta">
                        <div class="history-risk ${riskClass}">${report.risk_level}</div>
                        <div class="history-date">${date}</div>
                    </div>
                `;
                
                historyList.appendChild(historyItem);
            });
        } catch (error) {
            console.error('Failed to load history:', error);
        }
    }

    async loadReport(hash) {
        try {
            const response = await fetch(`/api/report/${hash}`);
            const report = await response.json();
            
            this.currentReport = report;
            this.showResults(report);
            
            // Show results section
            document.getElementById('uploadSection').hidden = true;
            document.getElementById('analysisSection').hidden = true;
            document.getElementById('resultsSection').hidden = false;
        } catch (error) {
            console.error('Failed to load report:', error);
        }
    }

    downloadReport() {
        if (!this.currentReport) return;

        const dataStr = JSON.stringify(this.currentReport, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        
        const link = document.createElement('a');
        link.href = URL.createObjectURL(dataBlob);
        link.download = `aegong_report_${this.currentReport.agent_hash.substring(0, 8)}.json`;
        link.click();
    }

    resetInterface() {
        // Reset file input
        document.getElementById('fileInput').value = '';
        
        // Reset upload button
        const uploadBtn = document.getElementById('uploadBtn');
        uploadBtn.disabled = true;
        uploadBtn.querySelector('.btn-text').textContent = 'Select File First';
        uploadBtn.querySelector('.btn-text').style.display = 'inline';
        uploadBtn.querySelector('.btn-loader').hidden = true;

        // Show upload section, hide others
        document.getElementById('uploadSection').hidden = false;
        document.getElementById('analysisSection').hidden = true;
        document.getElementById('resultsSection').hidden = true;

        // Reset progress
        document.getElementById('progressFill').style.width = '0%';
        
        this.updateStatus('Ready', 'ready');
        this.currentReport = null;
    }

    updateStatus(text, type) {
        const statusText = document.getElementById('statusText');
        const statusDot = document.getElementById('statusDot');
        
        statusText.textContent = text;
        
        // Update status dot color based on type
        const colors = {
            ready: '#4ecdc4',
            analyzing: '#ff9800',
            complete: '#4caf50',
            error: '#f44336'
        };
        
        statusDot.style.background = colors[type] || '#4ecdc4';
    }
}

// Initialize the interface when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new AegongInterface();
});