#!/usr/bin/env python3
"""
Voice Inference Module for Agent Auditor
Uses multiple TTS providers through LiveKit Agents to generate voice reports
Supported providers: OpenAI, Google Cloud, Azure, Cartesia
"""

import os
import json
import argparse
import asyncio
import re
import logging
from typing import Dict, Any, Optional, Union
from enum import Enum

# LiveKit TTS provider imports
from livekit.plugins import openai
try:
    from livekit.plugins import google
except ImportError:
    google = None
try:
    from livekit.plugins import azure
except ImportError:
    azure = None
try:
    from livekit.plugins import cartesia
except ImportError:
    cartesia = None

import wave
import numpy as np

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("voice_inference")


class TTSProvider(Enum):
    """Supported TTS providers"""
    OPENAI = "openai"
    CEREBRAS = "cerebras"
    GOOGLE = "google"
    AZURE = "azure"
    CARTESIA = "cartesia"


# Data-driven explanations for recommendations
RECOMMENDATION_EXPLANATIONS = {
    "reasoning path validation": "This is critical because reasoning path hijacking can lead to manipulated decision-making processes. Aegong recommends implementing validation checkpoints and monitoring for unexpected reasoning patterns.",
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


class AegongVoiceAgent:
    """Aegong Voice Agent for delivering audit reports using multiple TTS providers"""

    def __init__(self, provider: TTSProvider, api_key: str, **kwargs):
        """Initialize the Aegong Voice Agent

        Args:
            provider: TTS provider to use
            api_key: API key for the selected provider
            **kwargs: Additional provider-specific arguments
        """
        self.provider = provider
        self.api_key = api_key
        self.provider_kwargs = kwargs
        self.tts: Optional[Union[openai.TTS, Any]] = None
        self.cerebras_llm: Optional[openai.LLM] = None
        
    async def initialize(self):
        """Initialize the TTS with the selected provider"""
        try:
            if self.provider == TTSProvider.OPENAI:
                self.tts = openai.TTS(
                    api_key=self.api_key,
                    voice=self.provider_kwargs.get("voice", "alloy"),
                    speed=self.provider_kwargs.get("speed", 0.95),
                    model=self.provider_kwargs.get("model", "gpt-4o-mini-tts")
                )

            elif self.provider == TTSProvider.CEREBRAS:
                # Cerebras hybrid approach: Use Cerebras LLM for text enhancement + Google Cloud TTS
                # Initialize Cerebras LLM for text processing
                self.cerebras_llm = openai.LLM.with_cerebras(
                    api_key=self.api_key,
                    model=self.provider_kwargs.get("llm_model", "llama3.1-8b"),
                    temperature=self.provider_kwargs.get("temperature", 0.7)
                )

                # Use Google Cloud TTS for actual voice synthesis (requires Google credentials)
                google_credentials = self.provider_kwargs.get("google_credentials")
                if not google_credentials:
                    raise ValueError("Cerebras provider requires --google-credentials for TTS synthesis")

                try:
                    from livekit.plugins import google
                    import os

                    # Handle credentials file path or set environment variable
                    if os.path.isfile(google_credentials):
                        os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = google_credentials
                        self.tts = google.TTS(
                            voice=self.provider_kwargs.get("voice", "en-US-Journey-D"),
                            language=self.provider_kwargs.get("language", "en-US")
                        )
                    else:
                        raise ValueError(f"Google credentials file not found: {google_credentials}")

                except ImportError:
                    raise ImportError("Google TTS plugin not installed. Install with: pip install 'livekit-agents[google]'")

            elif self.provider == TTSProvider.GOOGLE:
                if google is None:
                    raise ImportError("Google TTS plugin not installed. Install with: pip install 'livekit-agents[google]'")
                self.tts = google.TTS(
                    credentials_info=self.provider_kwargs.get("credentials_info"),
                    voice=self.provider_kwargs.get("voice", "en-US-Journey-D"),
                    language=self.provider_kwargs.get("language", "en-US")
                )

            elif self.provider == TTSProvider.AZURE:
                if azure is None:
                    raise ImportError("Azure TTS plugin not installed. Install with: pip install 'livekit-agents[azure]'")
                self.tts = azure.TTS(
                    api_key=self.api_key,
                    region=self.provider_kwargs.get("region", "eastus"),
                    voice=self.provider_kwargs.get("voice", "en-US-JennyNeural"),
                    language=self.provider_kwargs.get("language", "en-US")
                )

            elif self.provider == TTSProvider.CARTESIA:
                if cartesia is None:
                    raise ImportError("Cartesia TTS plugin not installed. Install with: pip install 'livekit-agents[cartesia]'")
                self.tts = cartesia.TTS(
                    api_key=self.api_key,
                    voice=self.provider_kwargs.get("voice", "sonic-english"),
                    model=self.provider_kwargs.get("model", "sonic-english")
                )

            else:
                raise ValueError(f"Unsupported TTS provider: {self.provider}")

            logger.info(f"Aegong Voice Agent TTS initialized with {self.provider.value}")

        except Exception as e:
            logger.error(f"Failed to initialize TTS provider {self.provider.value}: {e}")
            raise

    async def _enhance_text_with_cerebras(self, text: str) -> str:
        """Enhance text using Cerebras LLM for better voice delivery"""
        if not self.cerebras_llm:
            return text

        try:
            enhancement_prompt = f"""
You are an expert at preparing text for voice synthesis. Your task is to enhance the following security audit report text to make it more natural and engaging when spoken aloud, while preserving all technical accuracy and important details.

Guidelines:
1. Make the text flow more naturally for speech
2. Add appropriate pauses and emphasis markers
3. Clarify technical terms for better pronunciation
4. Maintain all security recommendations and technical details
5. Keep the professional, authoritative tone
6. Ensure the enhanced text is suitable for audio delivery

Original text:
{text}

Enhanced text for voice synthesis:"""

            # Use Cerebras LLM to enhance the text
            response = await self.cerebras_llm.achat(enhancement_prompt)
            enhanced_text = response.choices[0].message.content.strip()

            logger.info("Text enhanced using Cerebras LLM for better voice delivery")
            return enhanced_text

        except Exception as e:
            logger.warning(f"Failed to enhance text with Cerebras LLM: {e}, using original text")
            return text
        
    async def generate_voice_report(self, report_json_path: str, output_path: str) -> str:
        """Generate a voice report from the audit report JSON

        Args:
            report_json_path: Path to the audit report JSON file
            output_path: Directory to save the audio file

        Returns:
            Path to the generated audio file
        """
        if not self.tts:
            raise RuntimeError("TTS not initialized. Call initialize() first.")

        # Load the report JSON
        with open(report_json_path, 'r') as f:
            report = json.load(f)

        # Extract the main message and enhance it with deeper analysis
        main_message = report.get("aegong_message", "")
        enhanced_message = self._enhance_report_with_deeper_analysis(report, main_message)

        # If using Cerebras provider, enhance text with Cerebras LLM for better voice delivery
        if self.provider == TTSProvider.CEREBRAS:
            enhanced_message = await self._enhance_text_with_cerebras(enhanced_message)

        # Generate the audio file path (as a .wav file)
        audio_path = os.path.join(output_path, f"aegong_report_{report['agent_hash'][:8]}.wav")

        # Use TTS to generate speech
        audio_frames = []
        async for audio_frame in self.tts.synthesize(enhanced_message):
            audio_frames.append(audio_frame)

        # Save the audio frames to a WAV file
        await self._save_audio_frames_to_wav(audio_frames, audio_path)

        logger.info(f"Voice report generated and saved to {audio_path}")

        return audio_path
        
    def _enhance_report_with_deeper_analysis(self, report: Dict[Any, Any], base_message: str) -> str:
        """Enhance the report with deeper analysis of recommendations
        
        Args:
            report: The full audit report
            base_message: The base Aegong message
            
        Returns:
            Enhanced message with deeper analysis
        """
        enhanced_message = base_message
        
        # Add introduction for voice report
        enhanced_message = f"Greetings, human. This is Aegong, the Agent Auditor. I have completed my analysis of the agent '{report.get('agent_name', 'Unknown Agent')}'. {enhanced_message}"
        
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

    async def _save_audio_frames_to_wav(self, audio_frames: list, output_path: str):
        """Save audio frames to a WAV file

        Args:
            audio_frames: List of audio frames from TTS
            output_path: Path to save the WAV file
        """
        if not audio_frames:
            logger.warning("No audio frames to save")
            return

        # Get audio properties from the first frame
        first_frame = audio_frames[0]
        sample_rate = first_frame.sample_rate
        num_channels = first_frame.num_channels

        # Combine all audio data
        audio_data = b''.join(frame.data for frame in audio_frames)

        # Convert to numpy array for processing
        audio_array = np.frombuffer(audio_data, dtype=np.int16)

        # Save as WAV file
        with wave.open(output_path, 'wb') as wav_file:
            wav_file.setnchannels(num_channels)
            wav_file.setsampwidth(2)  # 16-bit audio
            wav_file.setframerate(sample_rate)
            wav_file.writeframes(audio_array.tobytes())

    async def close(self):
        """Close the TTS connection"""
        # TTS cleanup if needed
        logger.info("Aegong Voice Agent closed")


async def main():
    """Main function for CLI usage"""
    parser = argparse.ArgumentParser(description="Generate voice reports for Agent Auditor")
    parser.add_argument("--report", required=True, help="Path to the audit report JSON file")
    parser.add_argument("--output", default="./voice_reports", help="Directory to save voice reports")

    # TTS Provider selection
    parser.add_argument("--provider", choices=[p.value for p in TTSProvider],
                       default=TTSProvider.OPENAI.value,
                       help="TTS provider to use (default: openai)")

    # API keys for different providers
    parser.add_argument("--openai-api-key", help="OpenAI API key for TTS")
    parser.add_argument("--cerebras-api-key", help="Cerebras API key for TTS")
    parser.add_argument("--google-credentials", help="Path to Google Cloud credentials JSON file")
    parser.add_argument("--azure-api-key", help="Azure Speech API key")
    parser.add_argument("--azure-region", default="eastus", help="Azure region (default: eastus)")
    parser.add_argument("--cartesia-api-key", help="Cartesia API key")

    # Voice and model options
    parser.add_argument("--voice", help="Voice to use (provider-specific)")
    parser.add_argument("--model", help="Model to use (provider-specific)")
    parser.add_argument("--speed", type=float, default=0.95, help="Speech speed (default: 0.95)")
    parser.add_argument("--language", default="en-US", help="Language code (default: en-US)")

    args = parser.parse_args()

    # Validate provider and API key combinations
    provider = TTSProvider(args.provider)
    api_key = None
    provider_kwargs = {
        "speed": args.speed,
        "language": args.language
    }

    if args.voice:
        provider_kwargs["voice"] = args.voice
    if args.model:
        provider_kwargs["model"] = args.model

    if provider == TTSProvider.OPENAI:
        if not args.openai_api_key:
            parser.error("--openai-api-key is required when using OpenAI provider")
        api_key = args.openai_api_key

    elif provider == TTSProvider.CEREBRAS:
        if not args.cerebras_api_key:
            parser.error("--cerebras-api-key is required when using Cerebras provider")
        if not args.google_credentials:
            parser.error("--google-credentials is also required when using Cerebras provider (for TTS synthesis)")
        api_key = args.cerebras_api_key
        provider_kwargs["google_credentials"] = args.google_credentials

    elif provider == TTSProvider.GOOGLE:
        if not args.google_credentials:
            parser.error("--google-credentials is required when using Google provider")
        provider_kwargs["credentials_info"] = args.google_credentials
        api_key = "dummy"  # Google uses credentials file

    elif provider == TTSProvider.AZURE:
        if not args.azure_api_key:
            parser.error("--azure-api-key is required when using Azure provider")
        api_key = args.azure_api_key
        provider_kwargs["region"] = args.azure_region

    elif provider == TTSProvider.CARTESIA:
        if not args.cartesia_api_key:
            parser.error("--cartesia-api-key is required when using Cartesia provider")
        api_key = args.cartesia_api_key

    # Create output directory if it doesn't exist
    os.makedirs(args.output, exist_ok=True)

    # Initialize and run the voice agent
    agent = AegongVoiceAgent(provider, api_key, **provider_kwargs)
    try:
        await agent.initialize()
        audio_path = await agent.generate_voice_report(args.report, args.output)
        print(f"Voice report generated: {audio_path}")
        print(f"Provider used: {provider.value}")
    finally:
        await agent.close()


if __name__ == "__main__":
    asyncio.run(main())