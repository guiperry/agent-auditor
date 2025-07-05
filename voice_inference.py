#!/usr/bin/env python3
"""
Voice Inference Module for Agent Auditor
Uses Cerebras through LiveKit Agents to generate voice reports
"""

import os
import json
import argparse
import asyncio
import re
import logging
from typing import Dict, Any, Optional

# LiveKit and Cerebras imports
from livekit import rtc
from livekit.agents import Agent
from livekit.agents.integrations.cerebras import CerebrasIntegration
from livekit.plugins.core import WaveSink

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("voice_inference")

# Data-driven explanations for recommendations
RECOMMENDATION_EXPLANATIONS = {
    "reasoning path validation": "This is critical because reasoning path hijacking can lead to manipulated decision-making processes. I recommend implementing validation checkpoints and monitoring for unexpected reasoning patterns.",
    "objective integrity": "Objective corruption can cause the agent to pursue harmful goals. Implement goal consistency verification and regular objective validation checks.",
    "memory integrity": "Memory poisoning allows persistent manipulation of the agent's knowledge base. Deploy cryptographic verification of memory states and implement read-only reference knowledge.",
    "action authorization": "Unauthorized actions bypass security controls. Implement fine-grained permission systems and action verification protocols.",
    "resource monitoring": "Resource manipulation can lead to denial of service or resource theft. Deploy adaptive resource limits and usage anomaly detection.",
    "identity verification": "Identity spoofing allows impersonation attacks. Implement multi-factor identity verification and credential rotation.",
    "trust validation": "Trust manipulation exploits human-agent interactions. Deploy trust boundary enforcement and interaction verification protocols.",
    "distributed oversight": "Oversight saturation overwhelms monitoring systems. Implement hierarchical monitoring with priority-based alert systems.",
    "immutable audit": "Governance evasion bypasses security policies. Deploy blockchain-based audit trails and policy enforcement verification.",
    "module validation": "Shield module failures indicate potential security gaps. Review and strengthen the affected security modules.",
}
DEFAULT_EXPLANATION = "This requires immediate attention to maintain agent security integrity."


class AegonVoiceAgent:
    """Aegon Voice Agent for delivering audit reports using Cerebras TTS"""
    
    def __init__(self, api_key: str, api_secret: str, ws_url: str):
        """Initialize the Aegon Voice Agent
        
        Args:
            api_key: LiveKit API key
            api_secret: LiveKit API secret
            ws_url: LiveKit WebSocket URL
        """
        self.api_key = api_key
        self.api_secret = api_secret
        self.ws_url = ws_url
        self.agent: Optional[Agent] = None
        self.cerebras: Optional[CerebrasIntegration] = None
        
    async def initialize(self):
        """Initialize the LiveKit agent with Cerebras integration"""
        # Create the agent
        self.agent = Agent(
            identity="aegon",
            name="Aegon",
            api_key=self.api_key,
            api_secret=self.api_secret,
            ws_url=self.ws_url
        )
        
        # Initialize Cerebras integration
        self.cerebras = CerebrasIntegration(
            self.agent,
            voice="male-deep",  # Use a deep male voice for Aegon
            language="en-US",
            speaking_rate=0.95,  # Slightly slower for clarity
            pitch=0.0,  # Neutral pitch
        )
        
        # Connect the agent
        await self.agent.connect()
        logger.info("Aegon Voice Agent initialized and connected")
        
    async def generate_voice_report(self, report_json_path: str, output_path: str) -> str:
        """Generate a voice report from the audit report JSON
        
        Args:
            report_json_path: Path to the audit report JSON file
            output_path: Directory to save the audio file
            
        Returns:
            Path to the generated audio file
        """
        if not self.agent or not self.cerebras:
            raise RuntimeError("Agent not initialized. Call initialize() first.")
            
        # Load the report JSON
        with open(report_json_path, 'r') as f:
            report = json.load(f)
            
        # Extract the main message and enhance it with deeper analysis
        main_message = report.get("aegon_message", "")
        enhanced_message = self._enhance_report_with_deeper_analysis(report, main_message)
        
        # Generate the audio file path (as a .wav file)
        audio_path = os.path.join(output_path, f"aegon_report_{report['agent_hash'][:8]}.wav")
        
        # Use Cerebras to generate speech stream
        audio_stream = await self.cerebras.speak(enhanced_message)
        
        # Save the audio stream to a WAV file
        # The Cerebras integration uses a sample rate of 24000 and 1 channel (mono)
        wave_sink = WaveSink(audio_path, sample_rate=24000, num_channels=1)
        
        try:
            async for frame in audio_stream:
                await wave_sink.write_frame(frame)
        finally:
            await wave_sink.close()
        
        logger.info(f"Voice report generated and saved to {audio_path}")
        
        return audio_path
        
    def _enhance_report_with_deeper_analysis(self, report: Dict[Any, Any], base_message: str) -> str:
        """Enhance the report with deeper analysis of recommendations
        
        Args:
            report: The full audit report
            base_message: The base Aegon message
            
        Returns:
            Enhanced message with deeper analysis
        """
        enhanced_message = base_message
        
        # Add introduction for voice report
        enhanced_message = f"Greetings, human. This is Aegon, the Agent Auditor. I have completed my analysis of the agent '{report.get('agent_name', 'Unknown Agent')}'. {enhanced_message}"
        
        # Add detailed recommendation analysis if recommendations exist
        if recommendations := report.get("recommendations", []):
            enhanced_message += "\n\nI have prepared detailed recommendations to address the security concerns:"
            
            for i, recommendation in enumerate(recommendations, 1):
                # Extract the core recommendation text for matching
                core_rec = re.sub(r'\s*\(\d+ instances detected\)', '', recommendation).strip().lower()
                
                # Find the matching explanation from our data-driven dictionary
                explanation = DEFAULT_EXPLANATION
                for keyword, detail in RECOMMENDATION_EXPLANATIONS.items():
                    if keyword in core_rec:
                        explanation = detail
                        break

                enhanced_message += f"\n\n{i}. {recommendation}. {explanation}"
        
        # Add conclusion
        enhanced_message += "\n\nI remain vigilant, protecting the digital realm one audit at a time. This concludes my voice report."
        
        return enhanced_message
    
    async def close(self):
        """Close the agent connection"""
        if self.agent:
            await self.agent.disconnect()
            logger.info("Aegon Voice Agent disconnected")


async def main():
    """Main function for CLI usage"""
    parser = argparse.ArgumentParser(description="Generate voice reports for Agent Auditor")
    parser.add_argument("--report", required=True, help="Path to the audit report JSON file")
    parser.add_argument("--output", default="./voice_reports", help="Directory to save voice reports")
    parser.add_argument("--api-key", required=True, help="LiveKit API key")
    parser.add_argument("--api-secret", required=True, help="LiveKit API secret")
    parser.add_argument("--ws-url", required=True, help="LiveKit WebSocket URL")
    
    args = parser.parse_args()
    
    # Create output directory if it doesn't exist
    os.makedirs(args.output, exist_ok=True)
    
    # Initialize and run the voice agent
    agent = AegonVoiceAgent(args.api_key, args.api_secret, args.ws_url)
    try:
        await agent.initialize()
        audio_path = await agent.generate_voice_report(args.report, args.output)
        print(f"Voice report generated: {audio_path}")
    finally:
        await agent.close()


if __name__ == "__main__":
    asyncio.run(main())