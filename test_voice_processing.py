#!/usr/bin/env python3
"""
Test script to verify voice inference JSON processing without requiring API keys
"""

import json
import asyncio
from voice_inference import AegongVoiceAgent

async def test_json_processing():
    """Test the JSON processing logic without TTS"""
    
    # Create a mock agent (we won't initialize TTS)
    agent = AegongVoiceAgent("dummy_key")
    
    # Test with the sample report
    report_path = "reports/report_14468589.json"
    
    try:
        # Load and process the JSON
        with open(report_path, 'r') as f:
            report_data = json.load(f)
        
        print("✅ Successfully loaded JSON report")
        print(f"Agent: {report_data.get('agent_name', 'Unknown')}")
        print(f"Risk Level: {report_data.get('risk_level', 'Unknown')}")
        print(f"Overall Risk: {report_data.get('overall_risk', 'Unknown')}")
        
        # Test the message enhancement logic
        aegong_message = report_data.get('aegong_message', '')
        if aegong_message:
            print("✅ Found Aegong message")
            print(f"Message length: {len(aegong_message)} characters")
            print(f"First 100 chars: {aegong_message[:100]}...")
            
            # Test the enhancement logic that would be used for TTS
            enhanced_message = f"""
🤖 Greetings! This is Aegong, your digital security guardian, delivering an audit report.

{aegong_message}

This concludes Aegong's security assessment. Stay vigilant in the digital realm!
"""
            print("✅ Successfully enhanced message for TTS")
            print(f"Enhanced message length: {len(enhanced_message)} characters")
        else:
            print("❌ No Aegong message found in report")
            
    except FileNotFoundError:
        print(f"❌ Report file not found: {report_path}")
    except json.JSONDecodeError as e:
        print(f"❌ Invalid JSON in report: {e}")
    except Exception as e:
        print(f"❌ Unexpected error: {e}")

if __name__ == "__main__":
    asyncio.run(test_json_processing())
