#!/usr/bin/env python3
"""
Example usage of the multi-provider TTS system for Agent Auditor
This script demonstrates how to use different TTS providers programmatically
"""

import asyncio
import os
from voice_inference import AegongVoiceAgent, TTSProvider


async def example_openai():
    """Example using OpenAI TTS"""
    print("=== OpenAI TTS Example ===")
    
    # You would need a real API key for this to work
    api_key = os.getenv("OPENAI_API_KEY", "your-openai-api-key-here")
    
    agent = AegongVoiceAgent(
        provider=TTSProvider.OPENAI,
        api_key=api_key,
        voice="alloy",
        speed=0.95,
        model="gpt-4o-mini-tts"
    )
    
    try:
        await agent.initialize()
        print("✅ OpenAI TTS initialized successfully")
        
        # Generate voice report (would fail without real API key)
        # audio_path = await agent.generate_voice_report("reports/report_14468589.json", "./voice_reports")
        # print(f"Audio generated: {audio_path}")
        
    except Exception as e:
        print(f"❌ OpenAI TTS failed: {e}")
    finally:
        await agent.close()


async def example_google():
    """Example using Google Cloud TTS"""
    print("\n=== Google Cloud TTS Example ===")
    
    # You would need real Google Cloud credentials for this to work
    credentials_path = os.getenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/credentials.json")
    
    agent = AegongVoiceAgent(
        provider=TTSProvider.GOOGLE,
        api_key="dummy",  # Google uses credentials file
        credentials_info=credentials_path,
        voice="en-US-Journey-D",
        language="en-US"
    )
    
    try:
        await agent.initialize()
        print("✅ Google Cloud TTS initialized successfully")
        
    except ImportError as e:
        print(f"❌ Google Cloud TTS plugin not installed: {e}")
        print("Install with: pip install 'livekit-agents[google]'")
    except Exception as e:
        print(f"❌ Google Cloud TTS failed: {e}")
    finally:
        await agent.close()


async def example_azure():
    """Example using Azure Speech TTS"""
    print("\n=== Azure Speech TTS Example ===")
    
    # You would need real Azure credentials for this to work
    api_key = os.getenv("AZURE_SPEECH_KEY", "your-azure-api-key-here")
    region = os.getenv("AZURE_SPEECH_REGION", "eastus")
    
    agent = AegongVoiceAgent(
        provider=TTSProvider.AZURE,
        api_key=api_key,
        region=region,
        voice="en-US-JennyNeural",
        language="en-US"
    )
    
    try:
        await agent.initialize()
        print("✅ Azure Speech TTS initialized successfully")
        
    except ImportError as e:
        print(f"❌ Azure TTS plugin not installed: {e}")
        print("Install with: pip install 'livekit-agents[azure]'")
    except Exception as e:
        print(f"❌ Azure Speech TTS failed: {e}")
    finally:
        await agent.close()


async def example_cerebras():
    """Example using Cerebras Hybrid TTS (Cerebras LLM + Google Cloud TTS)"""
    print("\n=== Cerebras Hybrid TTS Example ===")

    # You would need both Cerebras API key and Google Cloud credentials for this to work
    cerebras_key = os.getenv("CEREBRAS_API_KEY", "your-cerebras-api-key-here")
    google_credentials = os.getenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/credentials.json")

    agent = AegongVoiceAgent(
        provider=TTSProvider.CEREBRAS,
        api_key=cerebras_key,
        google_credentials=google_credentials,  # Required for TTS synthesis
        voice="en-US-Journey-D",
        language="en-US",
        llm_model="llama3.1-8b",  # Cerebras LLM model
        temperature=0.7
    )

    try:
        await agent.initialize()
        print("✅ Cerebras Hybrid TTS initialized successfully")
        print("   - Cerebras LLM for text enhancement")
        print("   - Google Cloud TTS for voice synthesis")

    except Exception as e:
        print(f"❌ Cerebras Hybrid TTS failed: {e}")
        if "google_credentials" in str(e).lower():
            print("   Make sure to provide both --cerebras-api-key and --google-credentials")
        elif "google" in str(e).lower():
            print("   Install Google TTS plugin: pip install 'livekit-agents[google]'")
    finally:
        await agent.close()


async def example_cartesia():
    """Example using Cartesia TTS"""
    print("\n=== Cartesia TTS Example ===")

    # You would need a real Cartesia API key for this to work
    api_key = os.getenv("CARTESIA_API_KEY", "your-cartesia-api-key-here")

    agent = AegongVoiceAgent(
        provider=TTSProvider.CARTESIA,
        api_key=api_key,
        voice="sonic-english",
        model="sonic-english"
    )

    try:
        await agent.initialize()
        print("✅ Cartesia TTS initialized successfully")

    except ImportError as e:
        print(f"❌ Cartesia TTS plugin not installed: {e}")
        print("Install with: pip install 'livekit-agents[cartesia]'")
    except Exception as e:
        print(f"❌ Cartesia TTS failed: {e}")
    finally:
        await agent.close()


async def main():
    """Run all examples"""
    print("Agent Auditor Multi-Provider TTS Examples")
    print("=" * 50)

    await example_openai()
    await example_cerebras()
    await example_google()
    await example_azure()
    await example_cartesia()
    
    print("\n" + "=" * 50)
    print("Examples completed!")
    print("\nTo use with real API keys:")
    print("1. Set environment variables for your chosen provider")
    print("2. Install the required plugin packages")
    print("3. Run the voice_inference.py script with your provider")
    print("\nSee TTS_PROVIDERS.md for detailed setup instructions.")


if __name__ == "__main__":
    asyncio.run(main())
