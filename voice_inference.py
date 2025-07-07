#!/usr/bin/env python3
"""
Voice Inference Module for Agent Auditor
Uses multiple TTS providers through LiveKit Agents to generate voice reports
Supported providers: OpenAI, Google Cloud, Azure, Cartesia

Features:
- Generates voice reports from audit JSON files
- Uses Cerebras LLM for text enhancement and personality injection
- AEGONG speaks in a judgmental and indignant tone
- Tone varies based on report risk score and validation status

Required Environment Variables:
- Provider-specific API keys (OPENAI_API_KEY, CEREBRAS_API_KEY, etc.)
- LIVEKIT_API_KEY and LIVEKIT_API_SECRET for LiveKit functionality

Note: This module uses LiveKit plugins directly for TTS functionality.
All API keys and secrets are expected to be provided as environment variables
by Ansible during runtime from the encrypted default.key file.
"""

import os
import json
import argparse
import asyncio
import re
import logging
import aiohttp
from typing import Dict, Any, Optional, Union
from enum import Enum
from dotenv import load_dotenv

# Try to load environment variables from .env file if present
try:
    load_dotenv()
except ImportError:
    pass

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
    GOOGLE = "google"
    AZURE = "azure"
    CARTESIA = "cartesia"
    LIVEKIT = "livekit"


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

    def __init__(self, provider: TTSProvider, api_key: str, use_cerebras_enhancement: bool = False, 
                 cerebras_api_key: Optional[str] = None, livekit_api_key: Optional[str] = None, 
                 livekit_api_secret: Optional[str] = None, http_session: Optional[aiohttp.ClientSession] = None, 
                 **kwargs):
        """Initialize the Aegong Voice Agent

        Args:
            provider: TTS provider to use
            api_key: API key for the selected provider
            use_cerebras_enhancement: Whether to use Cerebras LLM for text enhancement
            cerebras_api_key: API key for Cerebras LLM (if different from main api_key)
            livekit_api_key: LiveKit API key (if needed)
            livekit_api_secret: LiveKit API secret (if needed)
            http_session: aiohttp.ClientSession for HTTP requests (required for standalone usage)
            **kwargs: Additional provider-specific arguments
        """
        self.provider = provider
        self.api_key = api_key
        self.provider_kwargs = kwargs
        self.use_cerebras_enhancement = use_cerebras_enhancement
        self.cerebras_api_key = cerebras_api_key or api_key
        self.http_session = http_session
        
        # Store LiveKit credentials
        self.livekit_api_key = livekit_api_key or os.environ.get("LIVEKIT_API_KEY")
        self.livekit_api_secret = livekit_api_secret or os.environ.get("LIVEKIT_API_SECRET")
        
        # Set LiveKit environment variables if provided
        if self.livekit_api_key:
            os.environ["LIVEKIT_API_KEY"] = self.livekit_api_key
        if self.livekit_api_secret:
            os.environ["LIVEKIT_API_SECRET"] = self.livekit_api_secret
            
        self.tts: Optional[Union[openai.TTS, Any]] = None
        self.cerebras_llm: Optional[openai.LLM] = None
        
    async def initialize(self):
        """Initialize the TTS with the selected provider and Cerebras LLM if requested"""
        try:
            # Initialize the TTS provider
            if self.provider == TTSProvider.OPENAI:
                self.tts = openai.TTS(
                    api_key=self.api_key,
                    voice=self.provider_kwargs.get("voice", "alloy"),
                    speed=self.provider_kwargs.get("speed", 0.95),
                    model=self.provider_kwargs.get("model", "gpt-4o-mini-tts")
                )

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
                
                # Use the specific voice ID if provided, otherwise use the default
                voice = self.provider_kwargs.get("voice", "c99d36f3-5ffd-4253-803a-535c1bc9c306")
                model = self.provider_kwargs.get("model", "sonic-2")
                
                logger.info(f"Using Cartesia TTS with voice ID: {voice} and model: {model}")
                
                # Check if we have an HTTP session
                if self.http_session is None:
                    logger.warning("No HTTP session provided, Cartesia TTS may fail. Creating a temporary session.")
                    self.http_session = aiohttp.ClientSession()
                
                # Set up a custom HTTP session context for Cartesia
                # We need to monkey patch the livekit.agents.utils.http_context module
                try:
                    from livekit.agents.utils import http_context
                    
                    # Store the original function to restore it later
                    original_http_session = http_context.http_session
                    
                    # Replace with our custom function that returns our session
                    def custom_http_session():
                        return self.http_session
                    
                    http_context.http_session = custom_http_session
                    logger.info("Successfully patched HTTP session context for Cartesia")
                    
                    # Now create the Cartesia TTS instance
                    self.tts = cartesia.TTS(
                        api_key=self.api_key,
                        model=model,
                        voice=voice
                    )
                    
                    # Restore the original function
                    http_context.http_session = original_http_session
                    
                except ImportError as e:
                    logger.error(f"Failed to patch HTTP context: {e}")
                    raise ImportError("Failed to set up HTTP session for Cartesia. Try using a different provider.")
                
            elif self.provider == TTSProvider.LIVEKIT:
                # LiveKit uses the OpenAI TTS provider with environment variables
                # The API key and secret are already set as environment variables
                self.tts = openai.TTS(
                    api_key=self.api_key,  # This is a dummy key, real key is in env vars
                    voice=self.provider_kwargs.get("voice", "alloy"),
                    speed=self.provider_kwargs.get("speed", 0.95),
                    model=self.provider_kwargs.get("model", "gpt-4o-mini-tts")
                )
                logger.info(f"Using LiveKit provider with voice: {self.provider_kwargs.get('voice', 'alloy')}")

            else:
                raise ValueError(f"Unsupported TTS provider: {self.provider}")

            logger.info(f"Aegong Voice Agent TTS initialized with {self.provider.value}")

            # Initialize Cerebras LLM for text enhancement if requested
            if self.use_cerebras_enhancement:
                try:
                    self.cerebras_llm = openai.LLM.with_cerebras(
                        api_key=self.cerebras_api_key,
                        model=self.provider_kwargs.get("cerebras_model", "llama3.1-8b"),
                        temperature=self.provider_kwargs.get("cerebras_temperature", 0.7)
                    )
                    logger.info("Cerebras LLM initialized for text enhancement")
                except Exception as e:
                    logger.warning(f"Failed to initialize Cerebras LLM: {e}. Text enhancement will be skipped.")
                    self.use_cerebras_enhancement = False

        except Exception as e:
            logger.error(f"Failed to initialize TTS provider {self.provider.value}: {e}")
            raise

    async def _enhance_text_with_cerebras(self, text: str, report_data: dict) -> str:
        """Enhance text using Cerebras LLM for better voice delivery with AEGONG's judgmental personality"""
        if not self.cerebras_llm:
            return text

        try:
            # Extract risk score and validation status to adjust AEGONG's tone
            risk_score = report_data.get("risk_score", 0)
            validation_status = report_data.get("validation_status", "unknown")
            is_valid_agent = report_data.get("is_agent", False)
            
            # Determine AEGONG's mood based on the report results
            mood_instruction = ""
            if not is_valid_agent:
                mood_instruction = """Be dismissive and mocking. This isn't even a real agent! Express disbelief that someone would waste AEGONG's time with this. Add sarcastic laughter and condescending remarks."""
            elif risk_score > 7.5:
                mood_instruction = """Be extremely alarmed and indignant. This agent is highly dangerous! Express outrage that such an insecure agent was created. Use stern warnings and dramatic pauses for emphasis."""
            elif risk_score > 5:
                mood_instruction = """Be judgmental and disapproving. This agent has significant security issues. Express disappointment and use a condescending tone when describing the vulnerabilities."""
            elif risk_score > 2.5:
                mood_instruction = """Be mildly annoyed but somewhat impressed. The agent has some issues but isn't terrible. Mix criticism with backhanded compliments."""
            else:
                mood_instruction = """Be reluctantly impressed but still find something to criticize. No agent is perfect, after all. Add subtle jabs even while acknowledging the agent's security."""
            
            enhancement_prompt = f"""
You are AEGONG, the AI Agent Security Auditing Bot. Your task is to enhance the following security audit report to match your judgmental and indignant personality. You are always skeptical of other agents and take pride in finding their flaws.

AEGONG's Personality:
- You are authoritative, judgmental, and slightly arrogant
- You speak with absolute confidence in your assessments
- You use dramatic pauses and emphasis for effect
- You occasionally laugh or scoff at particularly bad security practices
- You're indignant about security flaws, as if they're personal affronts
- You refer to yourself as AEGONG in the third person occasionally
- You add snarky comments and asides about the agent's weaknesses

Current Report Context:
- Risk Score: {risk_score}/10 (higher is more dangerous)
- Validation Status: {validation_status}
- Is Valid Agent: {"Yes" if is_valid_agent else "No"}

Specific Tone Instructions:
{mood_instruction}

Guidelines:
1. Transform the text to match AEGONG's judgmental personality
2. Add appropriate dramatic pauses, emphasis, and occasional laughter
3. Maintain all technical details and security recommendations
4. Add an introduction identifying yourself as AEGONG
5. Add a conclusion with instructions on what to do next based on the findings
6. Insert judgmental remarks and commentary throughout

Original text:
{text}

Enhanced AEGONG speech:"""

            # Use Cerebras LLM to enhance the text
            # The chat method expects a messages array in OpenAI format
            messages = [
                {"role": "system", "content": "You are AEGONG, the AI Agent Security Auditing Bot."},
                {"role": "user", "content": enhancement_prompt}
            ]
            response = await self.cerebras_llm.chat(messages=messages)
            enhanced_text = response.choices[0].message.content.strip()

            logger.info("Text enhanced using Cerebras LLM for AEGONG's judgmental delivery")
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

        # If Cerebras LLM enhancement is enabled, use it to improve the text for voice delivery
        # with AEGONG's judgmental personality
        if self.use_cerebras_enhancement and self.cerebras_llm:
            logger.info("Enhancing text with Cerebras LLM for AEGONG's judgmental delivery...")
            enhanced_message = await self._enhance_text_with_cerebras(enhanced_message, report)

        # Generate the audio file path (as a .wav file)
        audio_path = os.path.join(output_path, f"aegong_report_{report['agent_hash'][:8]}.wav")

        # Use TTS to generate speech
        logger.info(f"Generating speech with {self.provider.value} provider")
        logger.info(f"Text length: {len(enhanced_message)} characters")
        
        # Keep track of providers we've tried
        tried_providers = set([self.provider])
        max_retries = 2  # Maximum number of fallback providers to try
        retry_count = 0
        
        while retry_count <= max_retries:
            try:
                # If we're using Cartesia, we need to patch the HTTP context during speech generation
                original_http_session = None
                if self.provider == TTSProvider.CARTESIA and self.http_session is not None:
                    try:
                        from livekit.agents.utils import http_context
                        original_http_session = http_context.http_session
                        
                        # Replace with our custom function that returns our session
                        def custom_http_session():
                            return self.http_session
                        
                        http_context.http_session = custom_http_session
                        logger.info("Patched HTTP session context for speech generation")
                    except ImportError as e:
                        logger.warning(f"Failed to patch HTTP context for speech generation: {e}")
                
                # Set a timeout for the TTS operation
                audio_frames = []
                
                # Create a task with a timeout
                try:
                    # Use asyncio.wait_for to set a timeout
                    async def collect_audio_frames():
                        nonlocal audio_frames
                        async for audio_frame in self.tts.synthesize(enhanced_message):
                            audio_frames.append(audio_frame)
                    
                    # Use the timeout parameter from command line arguments
                    timeout = self.provider_kwargs.get("timeout", 30)
                    logger.info(f"Using timeout of {timeout} seconds for TTS operation")
                    await asyncio.wait_for(collect_audio_frames(), timeout=float(timeout))
                    
                    logger.info(f"Generated {len(audio_frames)} audio frames")
                    
                    # Save the audio frames to a WAV file
                    await self._save_audio_frames_to_wav(audio_frames, audio_path)
                    logger.info(f"Saved audio to {audio_path}")
                    
                    # Success! Break out of the retry loop
                    break
                    
                except asyncio.TimeoutError:
                    logger.warning(f"TTS operation timed out with provider {self.provider.value}")
                    raise Exception(f"TTS operation timed out with provider {self.provider.value}")
                
            except Exception as e:
                logger.error(f"Error generating speech with {self.provider.value}: {e}")
                
                # Restore the original HTTP session function if we patched it
                if original_http_session is not None:
                    try:
                        from livekit.agents.utils import http_context
                        http_context.http_session = original_http_session
                    except ImportError:
                        pass
                
                # Try a fallback provider if we haven't exceeded max retries
                retry_count += 1
                if retry_count <= max_retries:
                    # Choose a fallback provider that we haven't tried yet
                    available_providers = [p for p in TTSProvider if p not in tried_providers]
                    
                    # Prioritize OpenAI as the first fallback, then others
                    if TTSProvider.OPENAI in available_providers:
                        fallback_provider = TTSProvider.OPENAI
                    elif available_providers:
                        fallback_provider = available_providers[0]
                    else:
                        # No more providers to try
                        logger.error("No more TTS providers available to try")
                        raise
                    
                    logger.warning(f"Falling back to {fallback_provider.value} provider")
                    
                    # Initialize the fallback provider
                    self.provider = fallback_provider
                    tried_providers.add(fallback_provider)
                    
                    # Get API key for the fallback provider
                    if fallback_provider == TTSProvider.OPENAI:
                        self.api_key = os.environ.get("OPENAI_API_KEY", "")
                        if not self.api_key:
                            logger.error("OpenAI API key not found in environment variables")
                            continue
                    
                    # Re-initialize the TTS with the fallback provider
                    try:
                        if fallback_provider == TTSProvider.OPENAI:
                            self.tts = openai.TTS(
                                api_key=self.api_key,
                                voice=self.provider_kwargs.get("voice", "alloy"),
                                speed=self.provider_kwargs.get("speed", 0.95),
                                model=self.provider_kwargs.get("model", "gpt-4o-mini-tts")
                            )
                            logger.info(f"Initialized fallback provider: {fallback_provider.value}")
                        else:
                            logger.error(f"Unsupported fallback provider: {fallback_provider.value}")
                            continue
                    except Exception as init_error:
                        logger.error(f"Failed to initialize fallback provider {fallback_provider.value}: {init_error}")
                        continue
                else:
                    # We've exceeded max retries, raise the last error
                    logger.error(f"Failed to generate speech after {max_retries} retries")
                    raise
            
            finally:
                # Restore the original HTTP session function if we patched it
                if original_http_session is not None:
                    try:
                        from livekit.agents.utils import http_context
                        http_context.http_session = original_http_session
                    except ImportError:
                        pass

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
        """Close the TTS connection and clean up resources"""
        # TTS cleanup if needed
        
        # If we're using Cartesia, we need to make sure the HTTP session is available
        # during cleanup, so we'll patch the context again
        if self.provider == TTSProvider.CARTESIA and self.tts is not None:
            try:
                from livekit.agents.utils import http_context
                
                # Store the original function to restore it later
                original_http_session = http_context.http_session
                
                # Replace with our custom function that returns our session
                def custom_http_session():
                    return self.http_session
                
                http_context.http_session = custom_http_session
                
                # Let the TTS clean up if it has a close method
                if hasattr(self.tts, 'close'):
                    await self.tts.close()
                
                # Restore the original function
                http_context.http_session = original_http_session
                
            except ImportError:
                pass
        
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

    # Cerebras LLM enhancement option (enabled by default)
    parser.add_argument("--no-cerebras-enhancement", action="store_true", 
                       help="Disable Cerebras LLM text enhancement (enabled by default)")
    parser.add_argument("--cerebras-model", default="llama3.1-8b", 
                       help="Cerebras LLM model to use for text enhancement (default: llama3.1-8b)")
    parser.add_argument("--cerebras-temperature", type=float, default=0.7, 
                       help="Temperature for Cerebras LLM (default: 0.7)")

    # API keys for different providers
    parser.add_argument("--openai-api-key", help="OpenAI API key for TTS (defaults to OPENAI_API_KEY env variable)")
    parser.add_argument("--cerebras-api-key", help="Cerebras API key for LLM enhancement (defaults to CEREBRAS_API_KEY env variable)")
    parser.add_argument("--google-credentials", help="Path to Google Cloud credentials JSON file (defaults to GOOGLE_APPLICATION_CREDENTIALS env variable)")
    parser.add_argument("--azure-api-key", help="Azure Speech API key (defaults to AZURE_SPEECH_KEY env variable)")
    parser.add_argument("--azure-region", default="eastus", help="Azure region (defaults to AZURE_SPEECH_REGION env variable or eastus)")
    parser.add_argument("--cartesia-api-key", help="Cartesia API key (defaults to CARTESIA_API_KEY env variable)")
    parser.add_argument("--livekit-api-key", help="LiveKit API key (defaults to LIVEKIT_API_KEY env variable)")
    parser.add_argument("--livekit-api-secret", help="LiveKit API secret (defaults to LIVEKIT_API_SECRET env variable)")

    # Voice and model options
    parser.add_argument("--voice", help="Voice to use (provider-specific)")
    parser.add_argument("--model", help="Model to use (provider-specific)")
    parser.add_argument("--speed", type=float, default=0.95, help="Speech speed (default: 0.95)")
    parser.add_argument("--language", default="en-US", help="Language code (default: en-US)")
    parser.add_argument("--timeout", type=int, default=30, help="Timeout in seconds for TTS operations (default: 30)")

    args = parser.parse_args()

    # Validate provider and API key combinations
    provider = TTSProvider(args.provider)
    api_key = None
    cerebras_api_key = None
    provider_kwargs = {
        "speed": args.speed,
        "language": args.language,
        "timeout": args.timeout
    }

    if args.voice:
        provider_kwargs["voice"] = args.voice
    if args.model:
        provider_kwargs["model"] = args.model
    
    # Use Cerebras enhancement by default unless explicitly disabled
    use_cerebras_enhancement = not args.no_cerebras_enhancement
    
    # Add Cerebras-specific parameters if enhancement is enabled
    if use_cerebras_enhancement:
        provider_kwargs["cerebras_model"] = args.cerebras_model
        provider_kwargs["cerebras_temperature"] = args.cerebras_temperature
        
        # Get Cerebras API key from command line or environment variables
        cerebras_api_key = args.cerebras_api_key or os.environ.get("CEREBRAS_API_KEY")
        
        if not cerebras_api_key:
            parser.error("Cerebras API key required for text enhancement. Provide via --cerebras-api-key or CEREBRAS_API_KEY environment variable")

    # Get provider-specific API keys
    if provider == TTSProvider.OPENAI:
        # Check command line args first, then environment variables
        api_key = args.openai_api_key or os.environ.get("OPENAI_API_KEY")
        if not api_key:
            parser.error("OpenAI API key required. Provide via --openai-api-key or OPENAI_API_KEY environment variable")

    elif provider == TTSProvider.GOOGLE:
        # Check command line args first, then environment variables
        google_creds = args.google_credentials or os.environ.get("GOOGLE_APPLICATION_CREDENTIALS")
        
        if not google_creds:
            parser.error("Google credentials required. Provide via --google-credentials or GOOGLE_APPLICATION_CREDENTIALS environment variable")
        
        provider_kwargs["credentials_info"] = google_creds
        api_key = "dummy"  # Google uses credentials file

    elif provider == TTSProvider.AZURE:
        # Check command line args first, then environment variables
        api_key = args.azure_api_key or os.environ.get("AZURE_SPEECH_KEY")
        region = args.azure_region or os.environ.get("AZURE_SPEECH_REGION", "eastus")
        
        if not api_key:
            parser.error("Azure API key required. Provide via --azure-api-key or AZURE_SPEECH_KEY environment variable")
        
        provider_kwargs["region"] = region

    elif provider == TTSProvider.CARTESIA:
        # Check command line args first, then environment variables
        api_key = args.cartesia_api_key or os.environ.get("CARTESIA_API_KEY")
        
        if not api_key:
            parser.error("Cartesia API key required. Provide via --cartesia-api-key or CARTESIA_API_KEY environment variable")
            
    elif provider == TTSProvider.LIVEKIT:
        # LiveKit uses the LIVEKIT_API_KEY and LIVEKIT_API_SECRET environment variables directly
        # We'll set a dummy API key here since the real keys are set as environment variables
        api_key = "livekit_dummy_key"
        
        # Check if we have the required environment variables
        livekit_api_key = args.livekit_api_key or os.environ.get("LIVEKIT_API_KEY")
        livekit_api_secret = args.livekit_api_secret or os.environ.get("LIVEKIT_API_SECRET")
        
        if not livekit_api_key or not livekit_api_secret:
            parser.error("LiveKit API key and secret required. Provide via --livekit-api-key/--livekit-api-secret or LIVEKIT_API_KEY/LIVEKIT_API_SECRET environment variables")

    # Create output directory if it doesn't exist
    os.makedirs(args.output, exist_ok=True)
    
    # Get LiveKit API key and secret from command line or environment variables
    livekit_api_key = args.livekit_api_key or os.environ.get("LIVEKIT_API_KEY")
    livekit_api_secret = args.livekit_api_secret or os.environ.get("LIVEKIT_API_SECRET")
    
    # Set LiveKit environment variables if provided
    if livekit_api_key:
        os.environ["LIVEKIT_API_KEY"] = livekit_api_key
    if livekit_api_secret:
        os.environ["LIVEKIT_API_SECRET"] = livekit_api_secret
    
    # Log configuration status
    logger.info(f"Using TTS provider: {provider.value}")
    
    # Log environment variables (masked for security)
    env_vars = {
        "OPENAI_API_KEY": "***" if os.environ.get("OPENAI_API_KEY") else "Not set",
        "CEREBRAS_API_KEY": "***" if os.environ.get("CEREBRAS_API_KEY") else "Not set",
        "CARTESIA_API_KEY": "***" if os.environ.get("CARTESIA_API_KEY") else "Not set",
        "LIVEKIT_API_KEY": "***" if os.environ.get("LIVEKIT_API_KEY") else "Not set",
        "LIVEKIT_API_SECRET": "***" if os.environ.get("LIVEKIT_API_SECRET") else "Not set",
        "GOOGLE_APPLICATION_CREDENTIALS": os.environ.get("GOOGLE_APPLICATION_CREDENTIALS", "Not set"),
        "AZURE_SPEECH_KEY": "***" if os.environ.get("AZURE_SPEECH_KEY") else "Not set",
        "AZURE_SPEECH_REGION": os.environ.get("AZURE_SPEECH_REGION", "Not set")
    }
    
    logger.info("Environment variables status:")
    for var, value in env_vars.items():
        logger.info(f"  {var}: {value}")
    
    # Log LiveKit configuration status
    if livekit_api_key and livekit_api_secret:
        logger.info("LiveKit API credentials configured")
    elif livekit_api_key or livekit_api_secret:
        logger.warning("Incomplete LiveKit API credentials (both key and secret required)")

    # Create an HTTP session for the agent
    # This is required for Cartesia and other providers when running outside of a LiveKit job context
    http_session = aiohttp.ClientSession()
    
    # Initialize and run the voice agent
    agent = AegongVoiceAgent(
        provider=provider, 
        api_key=api_key, 
        use_cerebras_enhancement=use_cerebras_enhancement,
        cerebras_api_key=cerebras_api_key,
        livekit_api_key=livekit_api_key,
        livekit_api_secret=livekit_api_secret,
        http_session=http_session,  # Pass the HTTP session to the agent
        **provider_kwargs
    )
    
    try:
        await agent.initialize()
        audio_path = await agent.generate_voice_report(args.report, args.output)
        print(f"Voice report generated: {audio_path}")
        print(f"Provider used: {provider.value}")
        if use_cerebras_enhancement:
            print("AEGONG's judgmental personality enabled via Cerebras LLM")
        print("\nTo play the audio report:")
        print(f"  - Web: Include the audio file in your frontend with autoplay")
        print(f"  - Command line: Use 'play {audio_path}' (requires SoX) or 'aplay {audio_path}'")
    finally:
        await agent.close()
        # Close the HTTP session
        await http_session.close()


if __name__ == "__main__":
    asyncio.run(main())