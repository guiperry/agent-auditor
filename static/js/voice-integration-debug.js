/**
 * AEGONG Voice Integration with Debug Features
 * Handles the automatic playback of AEGONG voice reports in the frontend
 * Includes enhanced error handling and debugging
 */

class AEGONGVoicePlayer {
    constructor() {
        this.audioElement = null;
        this.isPlaying = false;
        this.reportHash = null;
        this.voiceEnabled = true; // Voice enabled by default
        this.debugMode = true; // Enable debug mode
    }

    /**
     * Initialize the voice player
     */
    initialize() {
        console.log('Initializing AEGONG Voice Player with debugging');
        
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
            
            // Add error listener
            this.audioElement.addEventListener('error', (e) => {
                console.error('Audio element error:', e);
                this.showErrorMessage(`Audio playback error: ${e.target.error ? e.target.error.message : 'Unknown error'}`);
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
        
        // Add debug info
        if (this.debugMode) {
            this.showDebugMessage('AEGONG Voice Player initialized');
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
            
            if (this.debugMode) {
                this.showDebugMessage(`Voice ${this.voiceEnabled ? 'enabled' : 'disabled'}`);
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
        if (!this.voiceEnabled) {
            if (this.debugMode) {
                this.showDebugMessage('Voice playback skipped - voice is disabled');
            }
            return;
        }
        
        this.reportHash = reportHash;
        
        if (this.debugMode) {
            this.showDebugMessage(`Attempting to play voice report for ${reportHash}`);
        }
        
        // First check if the voice report exists by making a HEAD request
        fetch(`/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`, { method: 'HEAD' })
            .then(response => {
                if (this.debugMode) {
                    this.showDebugMessage(`Voice file check response: ${response.status} ${response.statusText}`);
                }
                
                if (response.ok) {
                    // Voice report exists, play it
                    const audioUrl = `/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`;
                    this.audioElement.src = audioUrl;
                    
                    if (this.debugMode) {
                        this.showDebugMessage(`Playing voice file from: ${audioUrl}`);
                    }
                    
                    // Play with a slight delay to ensure UI is ready
                    setTimeout(() => {
                        this.audioElement.play()
                            .then(() => {
                                this.isPlaying = true;
                                this.showPlayingIndicator();
                                
                                if (this.debugMode) {
                                    this.showDebugMessage('Voice playback started successfully');
                                }
                            })
                            .catch(error => {
                                console.error('Failed to play AEGONG voice report:', error);
                                this.showErrorMessage(`Failed to play voice report: ${error.message}`);
                                
                                // Try to generate it via the API
                                if (this.debugMode) {
                                    this.showDebugMessage('Playback failed, trying to generate new voice report');
                                }
                                this.tryGenerateVoiceReport(reportHash);
                            });
                    }, 1000);
                } else {
                    // Voice report doesn't exist, try to generate it via the API
                    if (this.debugMode) {
                        this.showDebugMessage('Voice file not found, trying to generate it');
                    }
                    this.tryGenerateVoiceReport(reportHash);
                }
            })
            .catch(error => {
                console.log('Error checking for voice report:', error);
                this.showErrorMessage(`Error checking for voice file: ${error.message}`);
                
                // Try the API as a fallback
                if (this.debugMode) {
                    this.showDebugMessage('Error checking for voice file, trying API as fallback');
                }
                this.tryGenerateVoiceReport(reportHash);
            });
    }
    
    /**
     * Try to generate a voice report via the API
     * @param {string} reportHash - The hash of the agent report
     */
    tryGenerateVoiceReport(reportHash) {
        if (this.debugMode) {
            this.showDebugMessage(`Attempting to generate voice report for ${reportHash}`);
        }
        
        // Show a notification that we're generating a voice report
        const notification = document.createElement('div');
        notification.className = 'voice-notification';
        notification.innerHTML = `
            <div class="voice-notification-content">
                <i class="fas fa-cog fa-spin"></i>
                <span>Generating AEGONG voice report...</span>
            </div>
        `;
        document.body.appendChild(notification);
        
        fetch(`/api/voice/${reportHash.substring(0, 8)}`)
            .then(response => {
                if (this.debugMode) {
                    this.showDebugMessage(`Voice API response status: ${response.status} ${response.statusText}`);
                }
                
                if (response.ok) {
                    return response.json();
                } else {
                    // Try to get the error message from the response
                    return response.text().then(text => {
                        try {
                            const errorJson = JSON.parse(text);
                            throw new Error(errorJson.error || 'Failed to generate voice report');
                        } catch (e) {
                            throw new Error(`Failed to generate voice report: ${text || response.statusText}`);
                        }
                    });
                }
            })
            .then(data => {
                // Remove the notification
                notification.remove();
                
                if (data.audio_url) {
                    if (this.debugMode) {
                        this.showDebugMessage(`Voice report generated successfully: ${data.audio_url}`);
                    }
                    
                    // Now try to play it
                    this.audioElement.src = data.audio_url;
                    
                    setTimeout(() => {
                        this.audioElement.play()
                            .then(() => {
                                this.isPlaying = true;
                                this.showPlayingIndicator();
                                
                                if (this.debugMode) {
                                    this.showDebugMessage('Generated voice report playing successfully');
                                }
                            })
                            .catch(error => {
                                console.error('Failed to play generated voice report:', error);
                                this.showErrorMessage(`Failed to play generated voice report: ${error.message}`);
                            });
                    }, 1000);
                } else {
                    console.log('No audio URL returned from API');
                    this.showErrorMessage('No audio URL returned from API');
                    
                    if (this.debugMode) {
                        this.showDebugMessage('API returned success but no audio URL was provided');
                    }
                }
            })
            .catch(error => {
                // Remove the notification
                notification.remove();
                
                console.error('Error generating voice report:', error);
                this.showErrorMessage(`Error generating voice report: ${error.message}`);
                
                if (this.debugMode) {
                    this.showDebugMessage(`Voice generation error: ${error.message}`);
                }
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
            
            if (this.debugMode) {
                this.showDebugMessage('Voice playback stopped');
            }
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
    
    /**
     * Show an error message to the user
     * @param {string} message - The error message to show
     */
    showErrorMessage(message) {
        const errorMsg = document.createElement('div');
        errorMsg.className = 'voice-error-message';
        errorMsg.innerHTML = `
            <div class="voice-error-content">
                <i class="fas fa-exclamation-triangle"></i>
                <span>${message}</span>
                <button class="voice-error-close">&times;</button>
            </div>
        `;
        
        // Add close button functionality
        errorMsg.querySelector('.voice-error-close').addEventListener('click', () => {
            errorMsg.remove();
        });
        
        document.body.appendChild(errorMsg);
        
        // Auto-remove after 10 seconds
        setTimeout(() => {
            if (document.body.contains(errorMsg)) {
                errorMsg.remove();
            }
        }, 10000);
    }
    
    /**
     * Show a debug message (only in debug mode)
     * @param {string} message - The debug message to show
     */
    showDebugMessage(message) {
        if (!this.debugMode) return;
        
        const debugMsg = document.createElement('div');
        debugMsg.className = 'voice-debug-message';
        debugMsg.innerHTML = `
            <div class="voice-debug-content">
                <i class="fas fa-bug"></i>
                <span>${message}</span>
                <button class="voice-debug-close">&times;</button>
            </div>
        `;
        
        // Add close button functionality
        debugMsg.querySelector('.voice-debug-close').addEventListener('click', () => {
            debugMsg.remove();
        });
        
        document.body.appendChild(debugMsg);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (document.body.contains(debugMsg)) {
                debugMsg.remove();
            }
        }, 5000);
        
        // Also log to console
        console.log(`[AEGONG Voice Debug] ${message}`);
    }
}

// Legacy integration with the AegongInterface class
// This is for backward compatibility with the existing UI
if (typeof AegongInterface !== 'undefined') {
    // Add this method to the AegongInterface class
    AegongInterface.prototype.checkVoiceReport = async function(reportHash) {
        console.log(`Legacy voice integration checking for report: ${reportHash}`);
        
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
            console.log(`Checking if voice file exists: /voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`);
            const fileCheck = await fetch(`/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`, { method: 'HEAD' });
            console.log(`File check response: ${fileCheck.status} ${fileCheck.statusText}`);
            
            if (!fileCheck.ok) {
                // File doesn't exist, hide the voice controls and return early
                console.log('Voice report file not found, hiding controls');
                voiceControls.style.display = 'none';
                return;
            }
        } catch (error) {
            // Error checking for file, hide the voice controls and return early
            console.log('Error checking for voice report file:', error);
            voiceControls.style.display = 'none';
            return;
        }

        // Show loading indicator
        voiceControls.style.display = 'block';
        voiceLoading.style.display = 'flex';
        audioElement.style.display = 'none';
        playVoiceBtn.style.display = 'none';

        try {
            console.log(`Fetching voice report from API: /api/voice/${reportHash.substring(0, 8)}`);
            const response = await fetch(`/api/voice/${reportHash.substring(0, 8)}`);
            console.log(`API response: ${response.status} ${response.statusText}`);
            
            if (response.ok) {
                const data = await response.json();
                console.log('API response data:', data);
                
                if (data.audio_url) {
                    console.log(`Audio URL received: ${data.audio_url}`);
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
                            console.log('Play button clicked, playing audio');
                            audioElement.play()
                                .then(() => console.log('Audio playback started'))
                                .catch(err => console.error('Audio playback error:', err));
                        });
                    }
                } else {
                    // No audio URL, hide the whole section
                    console.log('No audio URL in API response, hiding controls');
                    voiceControls.style.display = 'none';
                }
            } else {
                console.log('Voice report not available from API, trying direct file access');
                
                // Try to use the file directly if it exists
                const audioUrl = `/voice_reports/aegong_report_${reportHash.substring(0, 8)}.wav`;
                console.log(`Using direct file URL: ${audioUrl}`);
                
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
                        console.log('Play button clicked, playing audio directly');
                        audioElement.play()
                            .then(() => console.log('Direct audio playback started'))
                            .catch(err => console.error('Direct audio playback error:', err));
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
            console.log(`Report displayed, checking for voice report: ${report.agent_hash}`);
            this.checkVoiceReport(report.agent_hash);
            
            // If the new voice player is available, use it too
            if (window.aegongVoice && typeof window.aegongVoice.playReport === 'function') {
                try {
                    console.log('Using new voice player');
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
    console.log('DOM loaded, initializing AEGONG Voice Player');
    
    // Create and initialize the voice player
    window.aegongVoice = new AEGONGVoicePlayer();
    window.aegongVoice.initialize();
    
    // Check if we're on a report page and play the report
    const reportContainer = document.querySelector('.report-container');
    if (reportContainer) {
        const reportHash = reportContainer.dataset.reportHash;
        if (reportHash) {
            console.log(`Found report container with hash: ${reportHash}`);
            window.aegongVoice.playReport(reportHash);
        } else {
            console.log('Report container found but no report hash');
        }
    } else {
        console.log('No report container found on page');
    }
    
    console.log("AEGONG Voice Integration with debugging loaded");
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
    
    /* Error message styles */
    .voice-error-message {
        position: fixed;
        top: 20px;
        right: 20px;
        background-color: #dc3545;
        color: white;
        padding: 15px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        z-index: 1001;
        max-width: 350px;
        animation: fadeIn 0.3s ease-in-out;
    }
    
    .voice-error-content {
        display: flex;
        align-items: center;
        gap: 10px;
    }
    
    .voice-error-content i {
        font-size: 20px;
    }
    
    .voice-error-close {
        background: none;
        border: none;
        color: white;
        font-size: 20px;
        cursor: pointer;
        margin-left: auto;
    }
    
    /* Debug message styles */
    .voice-debug-message {
        position: fixed;
        bottom: 70px;
        right: 20px;
        background-color: #17a2b8;
        color: white;
        padding: 10px 15px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        z-index: 1001;
        max-width: 400px;
        font-family: monospace;
        font-size: 12px;
        animation: fadeIn 0.3s ease-in-out;
    }
    
    .voice-debug-content {
        display: flex;
        align-items: center;
        gap: 10px;
    }
    
    .voice-debug-content i {
        font-size: 16px;
    }
    
    .voice-debug-close {
        background: none;
        border: none;
        color: white;
        font-size: 16px;
        cursor: pointer;
        margin-left: auto;
    }
    
    /* Notification styles */
    .voice-notification {
        position: fixed;
        top: 20px;
        right: 20px;
        background-color: #17a2b8;
        color: white;
        padding: 15px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        z-index: 1001;
        max-width: 350px;
        animation: fadeIn 0.3s ease-in-out;
    }
    
    .voice-notification-content {
        display: flex;
        align-items: center;
        gap: 10px;
    }
    
    @keyframes fadeIn {
        from { opacity: 0; transform: translateY(-20px); }
        to { opacity: 1; transform: translateY(0); }
    }
`;
document.head.appendChild(style);