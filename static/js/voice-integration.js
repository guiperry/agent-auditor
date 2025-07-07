/**
 * AEGONG Voice Integration
 * Handles the automatic playback of AEGONG voice reports in the frontend
 */

class AEGONGVoicePlayer {
    constructor() {
        this.audioElement = null;
        this.isPlaying = false;
        this.reportHash = null;
        this.voiceEnabled = true; // Voice enabled by default
    }

    /**
     * Initialize the voice player
     */
    initialize() {
        // Create audio element if it doesn't exist
        if (!this.audioElement) {
            this.audioElement = document.createElement('audio');
            this.audioElement.id = 'aegong-voice-player';
            this.audioElement.style.display = 'none';
            document.body.appendChild(this.audioElement);
            
            // Add event listeners
            this.audioElement.addEventListener('ended', () => {
                this.isPlaying = false;
                this.showNextStepsTooltip();
            });
        }
        
        // Add voice toggle button to the UI
        this.addVoiceToggleButton();
        
        // Check for voice preference in localStorage
        const savedPreference = localStorage.getItem('aegong-voice-enabled');
        if (savedPreference !== null) {
            this.voiceEnabled = savedPreference === 'true';
            this.updateToggleButtonState();
        }
    }
    
    /**
     * Add voice toggle button to the UI
     */
    addVoiceToggleButton() {
        const toggleButton = document.createElement('button');
        toggleButton.id = 'voice-toggle-button';
        toggleButton.className = 'voice-toggle-button';
        toggleButton.innerHTML = this.voiceEnabled ? 
            '<i class="fas fa-volume-up"></i> AEGONG Voice ON' : 
            '<i class="fas fa-volume-mute"></i> AEGONG Voice OFF';
        
        toggleButton.addEventListener('click', () => {
            this.voiceEnabled = !this.voiceEnabled;
            localStorage.setItem('aegong-voice-enabled', this.voiceEnabled);
            this.updateToggleButtonState();
            
            if (!this.voiceEnabled && this.isPlaying) {
                this.stopPlayback();
            }
        });
        
        // Add to the header or controls section
        const controlsSection = document.querySelector('.report-controls') || 
                               document.querySelector('header') ||
                               document.body;
        
        controlsSection.appendChild(toggleButton);
    }
    
    /**
     * Update the toggle button state
     */
    updateToggleButtonState() {
        const toggleButton = document.getElementById('voice-toggle-button');
        if (toggleButton) {
            toggleButton.innerHTML = this.voiceEnabled ? 
                '<i class="fas fa-volume-up"></i> AEGONG Voice ON' : 
                '<i class="fas fa-volume-mute"></i> AEGONG Voice OFF';
            
            toggleButton.className = this.voiceEnabled ? 
                'voice-toggle-button active' : 
                'voice-toggle-button';
        }
    }

    /**
     * Play the AEGONG voice report for a specific agent
     * @param {string} reportHash - The hash of the agent report
     */
    playReport(reportHash) {
        if (!this.voiceEnabled) return;
        
        this.reportHash = reportHash;
        
        // First check if the voice report exists by making a HEAD request
        fetch(`/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`, { method: 'HEAD' })
            .then(response => {
                if (response.ok) {
                    // Voice report exists, play it
                    const audioUrl = `/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`;
                    this.audioElement.src = audioUrl;
                    
                    // Play with a slight delay to ensure UI is ready
                    setTimeout(() => {
                        this.audioElement.play()
                            .then(() => {
                                this.isPlaying = true;
                                this.showPlayingIndicator();
                            })
                            .catch(error => {
                                console.error('Failed to play AEGONG voice report:', error);
                                // Try to generate it via the API
                                this.tryGenerateVoiceReport(reportHash);
                            });
                    }, 1000);
                } else {
                    // Voice report doesn't exist, try to generate it via the API
                    console.log('Voice report not found, trying to generate it...');
                    this.tryGenerateVoiceReport(reportHash);
                }
            })
            .catch(error => {
                console.log('Error checking for voice report:', error);
                // Try the API as a fallback
                this.tryGenerateVoiceReport(reportHash);
            });
    }
    
    /**
     * Try to generate a voice report via the API
     * @param {string} reportHash - The hash of the agent report
     */
    tryGenerateVoiceReport(reportHash) {
        fetch(`/api/voice/${reportHash.substring(0, 8)}`)
            .then(response => {
                if (response.ok) {
                    return response.json();
                } else {
                    throw new Error('Failed to generate voice report');
                }
            })
            .then(data => {
                if (data.audio_url) {
                    console.log('Voice report generated successfully');
                    // Now try to play it
                    this.audioElement.src = data.audio_url;
                    
                    setTimeout(() => {
                        this.audioElement.play()
                            .then(() => {
                                this.isPlaying = true;
                                this.showPlayingIndicator();
                            })
                            .catch(error => {
                                console.error('Failed to play generated voice report:', error);
                            });
                    }, 1000);
                } else {
                    console.log('No audio URL returned from API');
                }
            })
            .catch(error => {
                console.log('Error generating voice report:', error);
            });
    }
    
    /**
     * Stop the current playback
     */
    stopPlayback() {
        if (this.audioElement && this.isPlaying) {
            this.audioElement.pause();
            this.audioElement.currentTime = 0;
            this.isPlaying = false;
            this.hidePlayingIndicator();
        }
    }
    
    /**
     * Show an indicator that AEGONG is speaking
     */
    showPlayingIndicator() {
        // Remove any existing indicator
        this.hidePlayingIndicator();
        
        // Create a new indicator
        const indicator = document.createElement('div');
        indicator.id = 'aegong-speaking-indicator';
        indicator.className = 'aegong-speaking';
        indicator.innerHTML = '<i class="fas fa-broadcast-tower"></i> AEGONG is speaking...';
        
        // Add a stop button
        const stopButton = document.createElement('button');
        stopButton.className = 'stop-voice-button';
        stopButton.innerHTML = '<i class="fas fa-stop"></i>';
        stopButton.addEventListener('click', () => this.stopPlayback());
        indicator.appendChild(stopButton);
        
        // Add to the document
        document.body.appendChild(indicator);
    }
    
    /**
     * Hide the speaking indicator
     */
    hidePlayingIndicator() {
        const indicator = document.getElementById('aegong-speaking-indicator');
        if (indicator) {
            indicator.remove();
        }
    }
    
    /**
     * Show a message that voice is available
     */
    showVoiceAvailableMessage() {
        const message = document.createElement('div');
        message.className = 'voice-available-message';
        message.innerHTML = `
            <i class="fas fa-volume-up"></i>
            <p>AEGONG voice report is available for this agent.</p>
            <button id="play-voice-report">Play Voice Report</button>
        `;
        
        // Add click handler for the play button
        message.querySelector('#play-voice-report').addEventListener('click', () => {
            message.remove();
            this.playReport(this.reportHash);
        });
        
        // Add to the document
        document.body.appendChild(message);
        
        // Auto-remove after 10 seconds
        setTimeout(() => {
            if (document.body.contains(message)) {
                message.remove();
            }
        }, 10000);
    }
    
    /**
     * Show a tooltip with next steps after the voice report ends
     */
    showNextStepsTooltip() {
        const tooltip = document.createElement('div');
        tooltip.className = 'next-steps-tooltip';
        tooltip.innerHTML = `
            <h3>Next Steps</h3>
            <p>Based on AEGONG's analysis, you should:</p>
            <ul>
                <li>Review the detailed threat analysis in the report</li>
                <li>Implement the recommended security measures</li>
                <li>Re-audit your agent after making changes</li>
            </ul>
            <button id="close-tooltip">Got it</button>
        `;
        
        // Add click handler for the close button
        tooltip.querySelector('#close-tooltip').addEventListener('click', () => {
            tooltip.remove();
        });
        
        // Add to the document
        document.body.appendChild(tooltip);
        
        // Auto-remove after 20 seconds
        setTimeout(() => {
            if (document.body.contains(tooltip)) {
                tooltip.remove();
            }
        }, 20000);
    }
}

// Legacy integration with the AegongInterface class
// This is for backward compatibility with the existing UI
if (typeof AegongInterface !== 'undefined') {
    // Add this method to the AegongInterface class
    AegongInterface.prototype.checkVoiceReport = async function(reportHash) {
        const voiceControls = document.getElementById('voiceControls');
        
        // First check if the voice controls element exists
        if (!voiceControls) {
            console.log('Voice controls element not found in the DOM');
            return;
        }
        
        const voiceLoading = document.getElementById('voiceLoading');
        const audioElement = document.getElementById('aegongVoice');
        const audioSource = document.getElementById('aegongVoiceSource');
        const playVoiceBtn = document.getElementById('playVoiceBtn');

        // First check if the voice report file exists directly
        try {
            const fileCheck = await fetch(`/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`, { method: 'HEAD' });
            if (!fileCheck.ok) {
                // File doesn't exist, hide the voice controls and return early
                voiceControls.style.display = 'none';
                return;
            }
        } catch (error) {
            // Error checking for file, hide the voice controls and return early
            console.log('Voice report file not available');
            voiceControls.style.display = 'none';
            return;
        }

        // Show loading indicator
        voiceControls.style.display = 'block';
        voiceLoading.style.display = 'flex';
        audioElement.style.display = 'none';
        playVoiceBtn.style.display = 'none';

        try {
            const response = await fetch(`/api/voice/${reportHash.substring(0, 8)}`);
            
            if (response.ok) {
                const data = await response.json();
                
                if (data.audio_url) {
                    // Hide loader, show player
                    voiceLoading.style.display = 'none';
                    audioElement.style.display = 'block';
                    playVoiceBtn.style.display = 'inline-block';

                    // Set audio source and load
                    audioSource.src = data.audio_url;
                    audioElement.load();
                    
                    // Setup voice playback button
                    if (playVoiceBtn) {
                        // Remove any existing event listeners to prevent duplicates
                        const newPlayBtn = playVoiceBtn.cloneNode(true);
                        playVoiceBtn.parentNode.replaceChild(newPlayBtn, playVoiceBtn);
                        
                        newPlayBtn.addEventListener('click', () => {
                            audioElement.play();
                        });
                    }
                } else {
                    // No audio URL, hide the whole section
                    voiceControls.style.display = 'none';
                }
            } else {
                console.log('Voice report not available from API');
                
                // Try to use the file directly if it exists
                const audioUrl = `/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`;
                
                // Hide loader, show player
                voiceLoading.style.display = 'none';
                audioElement.style.display = 'block';
                playVoiceBtn.style.display = 'inline-block';

                // Set audio source and load
                audioSource.src = audioUrl;
                audioElement.load();
                
                // Setup voice playback button
                if (playVoiceBtn) {
                    // Remove any existing event listeners to prevent duplicates
                    const newPlayBtn = playVoiceBtn.cloneNode(true);
                    playVoiceBtn.parentNode.replaceChild(newPlayBtn, playVoiceBtn);
                    
                    newPlayBtn.addEventListener('click', () => {
                        audioElement.play();
                    });
                }
            }
        } catch (error) {
            console.error('Error checking for voice report:', error);
            voiceControls.style.display = 'none';
        }
    };

    // Modify the showResults method to check for voice reports
    const originalShowResults = AegongInterface.prototype.showResults;
    AegongInterface.prototype.showResults = function(report) {
        // Call the original method
        originalShowResults.call(this, report);
        
        // Check for voice report
        if (report.agent_hash) {
            this.checkVoiceReport(report.agent_hash);
            
            // If the new voice player is available, use it too
            if (window.aegongVoice && typeof window.aegongVoice.playReport === 'function') {
                try {
                    window.aegongVoice.playReport(report.agent_hash);
                } catch (error) {
                    console.warn('Error playing voice report with new player:', error);
                }
            }
        }
    };
}

// Initialize the voice player when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Create and initialize the voice player
    window.aegongVoice = new AEGONGVoicePlayer();
    window.aegongVoice.initialize();
    
    // Check if we're on a report page and play the report
    const reportContainer = document.querySelector('.report-container');
    if (reportContainer) {
        const reportHash = reportContainer.dataset.reportHash;
        if (reportHash) {
            window.aegongVoice.playReport(reportHash);
        }
    }
    
    console.log("AEGONG Voice Integration loaded");
});

// Add CSS for the voice player UI
const style = document.createElement('style');
style.textContent = `
    .voice-toggle-button {
        background-color: #333;
        color: #fff;
        border: none;
        border-radius: 4px;
        padding: 8px 12px;
        margin: 10px;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 8px;
        transition: background-color 0.3s;
    }
    
    .voice-toggle-button.active {
        background-color: #007bff;
    }
    
    .voice-toggle-button:hover {
        background-color: #555;
    }
    
    .voice-toggle-button.active:hover {
        background-color: #0069d9;
    }
    
    .aegong-speaking {
        position: fixed;
        bottom: 20px;
        right: 20px;
        background-color: rgba(0, 0, 0, 0.8);
        color: #fff;
        padding: 10px 15px;
        border-radius: 30px;
        display: flex;
        align-items: center;
        gap: 10px;
        z-index: 1000;
        animation: pulse 2s infinite;
    }
    
    @keyframes pulse {
        0% { opacity: 0.8; }
        50% { opacity: 1; }
        100% { opacity: 0.8; }
    }
    
    .stop-voice-button {
        background-color: #dc3545;
        color: white;
        border: none;
        border-radius: 50%;
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        margin-left: 10px;
    }
    
    .voice-available-message {
        position: fixed;
        top: 20px;
        right: 20px;
        background-color: #343a40;
        color: white;
        padding: 15px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        z-index: 1000;
        max-width: 300px;
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    
    .voice-available-message i {
        font-size: 24px;
        margin-bottom: 10px;
        color: #007bff;
    }
    
    .voice-available-message button {
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        padding: 8px 16px;
        margin-top: 10px;
        cursor: pointer;
        transition: background-color 0.3s;
    }
    
    .voice-available-message button:hover {
        background-color: #0069d9;
    }
    
    .next-steps-tooltip {
        position: fixed;
        bottom: 20px;
        left: 20px;
        background-color: #343a40;
        color: white;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        z-index: 1000;
        max-width: 350px;
    }
    
    .next-steps-tooltip h3 {
        margin-top: 0;
        color: #007bff;
    }
    
    .next-steps-tooltip ul {
        padding-left: 20px;
        margin-bottom: 15px;
    }
    
    .next-steps-tooltip button {
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        padding: 8px 16px;
        cursor: pointer;
        float: right;
    }
    
    .next-steps-tooltip button:hover {
        background-color: #0069d9;
    }
`;
document.head.appendChild(style);/**
 * AEGONG Voice Integration
 * Handles the automatic playback of AEGONG voice reports in the frontend
 */

class AEGONGVoicePlayer {
    constructor() {
        this.audioElement = null;
        this.isPlaying = false;
        this.reportHash = null;
        this.voiceEnabled = true; // Voice enabled by default
    }

    /**
     * Initialize the voice player
     */
    initialize() {
        // Create audio element if it doesn't exist
        if (!this.audioElement) {
            this.audioElement = document.createElement('audio');
            this.audioElement.id = 'aegong-voice-player';
            this.audioElement.style.display = 'none';
            document.body.appendChild(this.audioElement);
            
            // Add event listeners
            this.audioElement.addEventListener('ended', () => {
                this.isPlaying = false;
                this.showNextStepsTooltip();
            });
        }
        
        // Add voice toggle button to the UI
        this.addVoiceToggleButton();
        
        // Check for voice preference in localStorage
        const savedPreference = localStorage.getItem('aegong-voice-enabled');
        if (savedPreference !== null) {
            this.voiceEnabled = savedPreference === 'true';
            this.updateToggleButtonState();
        }
    }
    
    /**
     * Add voice toggle button to the UI
     */
    addVoiceToggleButton() {
        const toggleButton = document.createElement('button');
        toggleButton.id = 'voice-toggle-button';
        toggleButton.className = 'voice-toggle-button';
        toggleButton.innerHTML = this.voiceEnabled ? 
            '<i class="fas fa-volume-up"></i> AEGONG Voice ON' : 
            '<i class="fas fa-volume-mute"></i> AEGONG Voice OFF';
        
        toggleButton.addEventListener('click', () => {
            this.voiceEnabled = !this.voiceEnabled;
            localStorage.setItem('aegong-voice-enabled', this.voiceEnabled);
            this.updateToggleButtonState();
            
            if (!this.voiceEnabled && this.isPlaying) {
                this.stopPlayback();
            }
        });
        
        // Add to the header or controls section
        const controlsSection = document.querySelector('.report-controls') || 
                               document.querySelector('header') ||
                               document.body;
        
        controlsSection.appendChild(toggleButton);
    }
    
    /**
     * Update the toggle button state
     */
    updateToggleButtonState() {
        const toggleButton = document.getElementById('voice-toggle-button');
        if (toggleButton) {
            toggleButton.innerHTML = this.voiceEnabled ? 
                '<i class="fas fa-volume-up"></i> AEGONG Voice ON' : 
                '<i class="fas fa-volume-mute"></i> AEGONG Voice OFF';
            
            toggleButton.className = this.voiceEnabled ? 
                'voice-toggle-button active' : 
                'voice-toggle-button';
        }
    }

    /**
     * Play the AEGONG voice report for a specific agent
     * @param {string} reportHash - The hash of the agent report
     */
    playReport(reportHash) {
        if (!this.voiceEnabled) return;
        
        this.reportHash = reportHash;
        
        // First check if the voice report exists by making a HEAD request
        fetch(`/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`, { method: 'HEAD' })
            .then(response => {
                if (response.ok) {
                    // Voice report exists, play it
                    const audioUrl = `/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`;
                    this.audioElement.src = audioUrl;
                    
                    // Play with a slight delay to ensure UI is ready
                    setTimeout(() => {
                        this.audioElement.play()
                            .then(() => {
                                this.isPlaying = true;
                                this.showPlayingIndicator();
                            })
                            .catch(error => {
                                console.error('Failed to play AEGONG voice report:', error);
                                // Try to generate it via the API
                                this.tryGenerateVoiceReport(reportHash);
                            });
                    }, 1000);
                } else {
                    // Voice report doesn't exist, try to generate it via the API
                    console.log('Voice report not found, trying to generate it...');
                    this.tryGenerateVoiceReport(reportHash);
                }
            })
            .catch(error => {
                console.log('Error checking for voice report:', error);
                // Try the API as a fallback
                this.tryGenerateVoiceReport(reportHash);
            });
    }
    
    /**
     * Try to generate a voice report via the API
     * @param {string} reportHash - The hash of the agent report
     */
    tryGenerateVoiceReport(reportHash) {
        fetch(`/api/voice/${reportHash.substring(0, 8)}`)
            .then(response => {
                if (response.ok) {
                    return response.json();
                } else {
                    throw new Error('Failed to generate voice report');
                }
            })
            .then(data => {
                if (data.audio_url) {
                    console.log('Voice report generated successfully');
                    // Now try to play it
                    this.audioElement.src = data.audio_url;
                    
                    setTimeout(() => {
                        this.audioElement.play()
                            .then(() => {
                                this.isPlaying = true;
                                this.showPlayingIndicator();
                            })
                            .catch(error => {
                                console.error('Failed to play generated voice report:', error);
                            });
                    }, 1000);
                } else {
                    console.log('No audio URL returned from API');
                }
            })
            .catch(error => {
                console.log('Error generating voice report:', error);
            });
    }
    
    /**
     * Stop the current playback
     */
    stopPlayback() {
        if (this.audioElement && this.isPlaying) {
            this.audioElement.pause();
            this.audioElement.currentTime = 0;
            this.isPlaying = false;
            this.hidePlayingIndicator();
        }
    }
    
    /**
     * Show an indicator that AEGONG is speaking
     */
    showPlayingIndicator() {
        // Remove any existing indicator
        this.hidePlayingIndicator();
        
        // Create a new indicator
        const indicator = document.createElement('div');
        indicator.id = 'aegong-speaking-indicator';
        indicator.className = 'aegong-speaking';
        indicator.innerHTML = '<i class="fas fa-broadcast-tower"></i> AEGONG is speaking...';
        
        // Add a stop button
        const stopButton = document.createElement('button');
        stopButton.className = 'stop-voice-button';
        stopButton.innerHTML = '<i class="fas fa-stop"></i>';
        stopButton.addEventListener('click', () => this.stopPlayback());
        indicator.appendChild(stopButton);
        
        // Add to the document
        document.body.appendChild(indicator);
    }
    
    /**
     * Hide the speaking indicator
     */
    hidePlayingIndicator() {
        const indicator = document.getElementById('aegong-speaking-indicator');
        if (indicator) {
            indicator.remove();
        }
    }
    
    /**
     * Show a message that voice is available
     */
    showVoiceAvailableMessage() {
        const message = document.createElement('div');
        message.className = 'voice-available-message';
        message.innerHTML = `
            <i class="fas fa-volume-up"></i>
            <p>AEGONG voice report is available for this agent.</p>
            <button id="play-voice-report">Play Voice Report</button>
        `;
        
        // Add click handler for the play button
        message.querySelector('#play-voice-report').addEventListener('click', () => {
            message.remove();
            this.playReport(this.reportHash);
        });
        
        // Add to the document
        document.body.appendChild(message);
        
        // Auto-remove after 10 seconds
        setTimeout(() => {
            if (document.body.contains(message)) {
                message.remove();
            }
        }, 10000);
    }
    
    /**
     * Show a tooltip with next steps after the voice report ends
     */
    showNextStepsTooltip() {
        const tooltip = document.createElement('div');
        tooltip.className = 'next-steps-tooltip';
        tooltip.innerHTML = `
            <h3>Next Steps</h3>
            <p>Based on AEGONG's analysis, you should:</p>
            <ul>
                <li>Review the detailed threat analysis in the report</li>
                <li>Implement the recommended security measures</li>
                <li>Re-audit your agent after making changes</li>
            </ul>
            <button id="close-tooltip">Got it</button>
        `;
        
        // Add click handler for the close button
        tooltip.querySelector('#close-tooltip').addEventListener('click', () => {
            tooltip.remove();
        });
        
        // Add to the document
        document.body.appendChild(tooltip);
        
        // Auto-remove after 20 seconds
        setTimeout(() => {
            if (document.body.contains(tooltip)) {
                tooltip.remove();
            }
        }, 20000);
    }
}

// Legacy integration with the AegongInterface class
// This is for backward compatibility with the existing UI
if (typeof AegongInterface !== 'undefined') {
    // Add this method to the AegongInterface class
    AegongInterface.prototype.checkVoiceReport = async function(reportHash) {
        const voiceControls = document.getElementById('voiceControls');
        
        // First check if the voice controls element exists
        if (!voiceControls) {
            console.log('Voice controls element not found in the DOM');
            return;
        }
        
        const voiceLoading = document.getElementById('voiceLoading');
        const audioElement = document.getElementById('aegongVoice');
        const audioSource = document.getElementById('aegongVoiceSource');
        const playVoiceBtn = document.getElementById('playVoiceBtn');

        // First check if the voice report file exists directly
        try {
            const fileCheck = await fetch(`/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`, { method: 'HEAD' });
            if (!fileCheck.ok) {
                // File doesn't exist, hide the voice controls and return early
                voiceControls.style.display = 'none';
                return;
            }
        } catch (error) {
            // Error checking for file, hide the voice controls and return early
            console.log('Voice report file not available');
            voiceControls.style.display = 'none';
            return;
        }

        // Show loading indicator
        voiceControls.style.display = 'block';
        voiceLoading.style.display = 'flex';
        audioElement.style.display = 'none';
        playVoiceBtn.style.display = 'none';

        try {
            const response = await fetch(`/api/voice/${reportHash.substring(0, 8)}`);
            
            if (response.ok) {
                const data = await response.json();
                
                if (data.audio_url) {
                    // Hide loader, show player
                    voiceLoading.style.display = 'none';
                    audioElement.style.display = 'block';
                    playVoiceBtn.style.display = 'inline-block';

                    // Set audio source and load
                    audioSource.src = data.audio_url;
                    audioElement.load();
                    
                    // Setup voice playback button
                    if (playVoiceBtn) {
                        // Remove any existing event listeners to prevent duplicates
                        const newPlayBtn = playVoiceBtn.cloneNode(true);
                        playVoiceBtn.parentNode.replaceChild(newPlayBtn, playVoiceBtn);
                        
                        newPlayBtn.addEventListener('click', () => {
                            audioElement.play();
                        });
                    }
                } else {
                    // No audio URL, hide the whole section
                    voiceControls.style.display = 'none';
                }
            } else {
                console.log('Voice report not available from API');
                
                // Try to use the file directly if it exists
                const audioUrl = `/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`;
                
                // Hide loader, show player
                voiceLoading.style.display = 'none';
                audioElement.style.display = 'block';
                playVoiceBtn.style.display = 'inline-block';

                // Set audio source and load
                audioSource.src = audioUrl;
                audioElement.load();
                
                // Setup voice playback button
                if (playVoiceBtn) {
                    // Remove any existing event listeners to prevent duplicates
                    const newPlayBtn = playVoiceBtn.cloneNode(true);
                    playVoiceBtn.parentNode.replaceChild(newPlayBtn, playVoiceBtn);
                    
                    newPlayBtn.addEventListener('click', () => {
                        audioElement.play();
                    });
                }
            }
        } catch (error) {
            console.error('Error checking for voice report:', error);
            voiceControls.style.display = 'none';
        }
    };

    // Modify the showResults method to check for voice reports
    const originalShowResults = AegongInterface.prototype.showResults;
    AegongInterface.prototype.showResults = function(report) {
        // Call the original method
        originalShowResults.call(this, report);
        
        // Check for voice report
        if (report.agent_hash) {
            this.checkVoiceReport(report.agent_hash);
            
            // If the new voice player is available, use it too
            if (window.aegongVoice && typeof window.aegongVoice.playReport === 'function') {
                try {
                    window.aegongVoice.playReport(report.agent_hash);
                } catch (error) {
                    console.warn('Error playing voice report with new player:', error);
                }
            }
        }
    };
}

// Initialize the voice player when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Create and initialize the voice player
    window.aegongVoice = new AEGONGVoicePlayer();
    window.aegongVoice.initialize();
    
    // Check if we're on a report page and play the report
    const reportContainer = document.querySelector('.report-container');
    if (reportContainer) {
        const reportHash = reportContainer.dataset.reportHash;
        if (reportHash) {
            window.aegongVoice.playReport(reportHash);
        }
    }
    
    console.log("AEGONG Voice Integration loaded");
});

// Add CSS for the voice player UI
const style = document.createElement('style');
style.textContent = `
    .voice-toggle-button {
        background-color: #333;
        color: #fff;
        border: none;
        border-radius: 4px;
        padding: 8px 12px;
        margin: 10px;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 8px;
        transition: background-color 0.3s;
    }
    
    .voice-toggle-button.active {
        background-color: #007bff;
    }
    
    .voice-toggle-button:hover {
        background-color: #555;
    }
    
    .voice-toggle-button.active:hover {
        background-color: #0069d9;
    }
    
    .aegong-speaking {
        position: fixed;
        bottom: 20px;
        right: 20px;
        background-color: rgba(0, 0, 0, 0.8);
        color: #fff;
        padding: 10px 15px;
        border-radius: 30px;
        display: flex;
        align-items: center;
        gap: 10px;
        z-index: 1000;
        animation: pulse 2s infinite;
    }
    
    @keyframes pulse {
        0% { opacity: 0.8; }
        50% { opacity: 1; }
        100% { opacity: 0.8; }
    }
    
    .stop-voice-button {
        background-color: #dc3545;
        color: white;
        border: none;
        border-radius: 50%;
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        margin-left: 10px;
    }
    
    .voice-available-message {
        position: fixed;
        top: 20px;
        right: 20px;
        background-color: #343a40;
        color: white;
        padding: 15px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        z-index: 1000;
        max-width: 300px;
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    
    .voice-available-message i {
        font-size: 24px;
        margin-bottom: 10px;
        color: #007bff;
    }
    
    .voice-available-message button {
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        padding: 8px 16px;
        margin-top: 10px;
        cursor: pointer;
        transition: background-color 0.3s;
    }
    
    .voice-available-message button:hover {
        background-color: #0069d9;
    }
    
    .next-steps-tooltip {
        position: fixed;
        bottom: 20px;
        left: 20px;
        background-color: #343a40;
        color: white;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        z-index: 1000;
        max-width: 350px;
    }
    
    .next-steps-tooltip h3 {
        margin-top: 0;
        color: #007bff;
    }
    
    .next-steps-tooltip ul {
        padding-left: 20px;
        margin-bottom: 15px;
    }
    
    .next-steps-tooltip button {
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        padding: 8px 16px;
        cursor: pointer;
        float: right;
    }
    
    .next-steps-tooltip button:hover {
        background-color: #0069d9;
    }
`;
document.head.appendChild(style);