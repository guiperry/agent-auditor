class AegonInterface {
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
        if (message.type === 'aegon_message') {
            this.showAegonMessage(message.message);
        }
    }

    showAegonMessage(message) {
        // Could be used for real-time updates during analysis
        console.log('Aegon says:', message);
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

            const report = await response.json();
            
            if (response.ok) {
                this.currentReport = report;
                this.showResults(report);
                this.loadHistory(); // Refresh history
            } else {
                throw new Error(report.error || 'Analysis failed');
            }
        } catch (error) {
            console.error('Analysis error:', error);
            this.updateStatus('Analysis failed', 'error');
        }
    }

    simulateProgress() {
        const progressFill = document.getElementById('progressFill');
        const analysisStatus = document.getElementById('analysisStatus');
        
        const steps = [
            { progress: 10, text: 'Aegon is awakening his sensors...' },
            { progress: 25, text: 'Scanning binary structure...' },
            { progress: 40, text: 'Analyzing threat vectors...' },
            { progress: 60, text: 'Running SHIELD validations...' },
            { progress: 80, text: 'Calculating risk assessment...' },
            { progress: 95, text: 'Preparing Aegon\'s verdict...' },
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
        }, 800);
    }

    showResults(report) {
        // Hide analysis section and show results
        document.getElementById('analysisSection').hidden = true;
        document.getElementById('resultsSection').hidden = false;

        this.updateStatus('Analysis complete', 'complete');

        // Populate overview
        document.getElementById('agentName').textContent = report.agent_name || 'Unknown Agent';
        document.getElementById('threatCount').textContent = report.threats.length;
        document.getElementById('riskScore').textContent = `${Math.round(report.overall_risk * 100)}%`;
        
        // Update risk badge
        const riskBadge = document.getElementById('riskBadge');
        const riskLevel = document.getElementById('riskLevel');
        riskLevel.textContent = report.risk_level;
        riskBadge.className = `risk-badge risk-${report.risk_level.toLowerCase()}`;

        // Show Aegon's message
        document.getElementById('aegonText').textContent = report.aegon_message;

        // Populate threats
        this.populateThreats(report.threats);

        // Populate SHIELD results
        this.populateShields(report.shield_results);

        // Populate recommendations
        this.populateRecommendations(report.recommendations);

        // Update SHIELD score
        const shieldPassed = this.countPassedShields(report.shield_results);
        document.getElementById('shieldScore').textContent = `${shieldPassed}/6`;
    }

    populateThreats(threats) {
        const threatsList = document.getElementById('threatsList');
        threatsList.innerHTML = '';

        if (threats.length === 0) {
            threatsList.innerHTML = '<p class="no-threats">🎉 No threats detected! Aegon is pleased.</p>';
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
        link.download = `aegon_report_${this.currentReport.agent_hash.substring(0, 8)}.json`;
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
    new AegonInterface();
});