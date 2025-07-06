// Voice Integration for Aegong Agent Auditor

// Add this method to the AegongInterface class
AegongInterface.prototype.checkVoiceReport = async function(reportHash) {
    const voiceControls = document.getElementById('voiceControls');
    const voiceLoading = document.getElementById('voiceLoading');
    const audioElement = document.getElementById('aegongVoice');
    const audioSource = document.getElementById('aegongVoiceSource');
    const playVoiceBtn = document.getElementById('playVoiceBtn');

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
                    playVoiceBtn.addEventListener('click', () => {
                        audioElement.play();
                    });
                }
            } else {
                // No audio URL, hide the whole section
                voiceControls.style.display = 'none';
            }
        } else {
            console.log('Voice report not available');
            voiceControls.style.display = 'none';
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
    }
};

console.log("Voice integration loaded");