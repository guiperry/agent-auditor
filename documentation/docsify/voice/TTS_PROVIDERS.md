# TTS Providers Guide for Agent Auditor Voice Inference

The Agent Auditor voice inference system now supports multiple Text-to-Speech (TTS) providers. This guide explains how to set up and use each provider.

## Supported Providers

### 1. OpenAI TTS (Default)
- **High-quality voices**: alloy, echo, fable, onyx, nova, shimmer
- **Models**: gpt-4o-mini-tts, gpt-4o-tts
- **Features**: Fast, reliable, good quality

**Setup:**
```bash
# Already included in base requirements
pip install 'livekit-agents[openai]~=1.0'
```

**Usage:**
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider openai \
  --openai-api-key YOUR_OPENAI_API_KEY \
  --voice alloy \
  --speed 0.95
```

### 2. Cerebras Hybrid TTS
- **Approach**: Cerebras LLM for text enhancement + Google Cloud TTS for voice synthesis
- **Features**: Fast LLM processing, enhanced text for better voice delivery
- **Requirements**: Both Cerebras API key (for LLM) and Google Cloud credentials (for TTS)
- **Benefits**: Leverages Cerebras' fast inference to optimize text before high-quality Google TTS

**Setup:**
```bash
# Uses OpenAI plugin for Cerebras LLM and Google plugin for TTS
pip install 'livekit-agents[openai]~=1.0'
pip install 'livekit-agents[google]'
```

**Usage:**
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider cerebras \
  --cerebras-api-key YOUR_CEREBRAS_API_KEY \
  --google-credentials /path/to/credentials.json \
  --voice en-US-Journey-D \
  --language en-US
```

### 3. Google Cloud TTS
- **High-quality voices**: en-US-Journey-D, en-US-Wavenet-D, etc.
- **Features**: Excellent quality, many languages, SSML support

**Setup:**
```bash
pip install 'livekit-agents[google]'
# Set up Google Cloud credentials
export GOOGLE_APPLICATION_CREDENTIALS="path/to/credentials.json"
```

**Usage:**
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider google \
  --google-credentials /path/to/credentials.json \
  --voice en-US-Journey-D \
  --language en-US
```

### 4. Azure Speech TTS
- **High-quality voices**: en-US-JennyNeural, en-US-AriaNeural, etc.
- **Features**: Neural voices, SSML support, many languages

**Setup:**
```bash
pip install 'livekit-agents[azure]'
```

**Usage:**
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider azure \
  --azure-api-key YOUR_AZURE_API_KEY \
  --azure-region eastus \
  --voice en-US-JennyNeural \
  --language en-US
```

### 5. Cartesia TTS
- **High-quality voices**: sonic-english, sonic-multilingual
- **Features**: Very fast, low latency, high quality

**Setup:**
```bash
pip install 'livekit-agents[cartesia]'
```

**Usage:**
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider cartesia \
  --cartesia-api-key YOUR_CARTESIA_API_KEY \
  --voice sonic-english \
  --model sonic-english
```

## Command Line Options

### Required Arguments
- `--report`: Path to the audit report JSON file
- Provider-specific API key (see examples above)

### Optional Arguments
- `--provider`: TTS provider (openai, cerebras, google, azure, cartesia) [default: openai]
- `--output`: Output directory for voice files [default: ./voice_reports]
- `--voice`: Voice to use (provider-specific)
- `--model`: Model to use (provider-specific)
- `--speed`: Speech speed [default: 0.95]
- `--language`: Language code [default: en-US]

### Provider-Specific Options
- `--azure-region`: Azure region [default: eastus]
- `--google-credentials`: Path to Google Cloud credentials JSON

## API Key Setup

### OpenAI
1. Get API key from https://platform.openai.com/api-keys
2. Set environment variable: `export OPENAI_API_KEY="your-key"`
3. Or use `--openai-api-key` argument

### Cerebras (Hybrid)
1. Get Cerebras API key from https://inference.cerebras.ai/
2. Set up Google Cloud credentials (see Google Cloud section above)
3. Set environment variables:
   - `export CEREBRAS_API_KEY="your-cerebras-key"`
   - `export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"`
4. Or use `--cerebras-api-key` and `--google-credentials` arguments

### Google Cloud
1. Create a Google Cloud project
2. Enable the Text-to-Speech API
3. Create a service account and download credentials JSON
4. Set `GOOGLE_APPLICATION_CREDENTIALS` environment variable

### Azure
1. Create an Azure Speech resource
2. Get the API key and region from Azure portal
3. Use `--azure-api-key` and `--azure-region` arguments

### Cartesia
1. Get API key from https://cartesia.ai/
2. Use `--cartesia-api-key` argument

## Examples

### Basic OpenAI usage:
```bash
python3 voice_inference.py --report reports/report_14468589.json --openai-api-key sk-...
```

### Enhanced Cerebras hybrid (LLM + TTS):
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider cerebras \
  --cerebras-api-key csk-... \
  --google-credentials ./gcp-credentials.json \
  --voice en-US-Journey-D
```

### High-quality Google Cloud:
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider google \
  --google-credentials ./gcp-credentials.json \
  --voice en-US-Wavenet-D
```

### Fast Cartesia:
```bash
python3 voice_inference.py \
  --report reports/report_14468589.json \
  --provider cartesia \
  --cartesia-api-key ctsa-... \
  --voice sonic-english
```

## Troubleshooting

### Import Errors
If you get import errors for specific providers, install the required plugin:
```bash
pip install 'livekit-agents[google]'  # For Google
pip install 'livekit-agents[azure]'   # For Azure
pip install 'livekit-agents[cartesia]' # For Cartesia
```

### Authentication Errors
- Verify your API keys are correct
- Check that credentials files exist and are readable
- Ensure environment variables are set correctly

### Voice/Model Errors
- Check provider documentation for available voices
- Some voices may not be available in all regions
- Try default voices if custom ones fail

## Note on LLM-Only Providers

Several AI providers specialize in LLM inference but do not offer TTS services:

### Cerebras
- **Status**: LLM provider only (Llama models)
- **TTS Support**: ✅ Hybrid solution available (Cerebras LLM + Google Cloud TTS)
- **Benefits**: Fast text enhancement with Cerebras + high-quality TTS with Google

### Anthropic (Claude)
- **Status**: LLM provider only
- **TTS Support**: ❌ Not available
- **Alternative**: Use Azure TTS (Microsoft's neural voices)

### DeepSeek
- **Status**: LLM provider only
- **TTS Support**: ❌ Not available
- **Alternative**: Use Google Cloud TTS (Google's WaveNet/Journey voices)

### Recommended TTS Providers
- **For fastest response**: Use Cartesia TTS
- **For highest quality**: Use Google Cloud TTS or Azure TTS
- **For reliability**: Use OpenAI TTS


---

<div class="footer-links">
<a href="#/legal/CODE_OF_CONDUCT.md" class="footer-link">Contributor Covenant Code of Conduct</a> | <a href="#/legal/PRIVACY_POLICY.md" class="footer-link">PRIVACY_POLICY.md</a> | <a href="#/legal/TERMS_AND_CONDITIONS.md" class="footer-link">TERMS AND CONDITIONS</a>

© 2025 Agent Auditor
</div>
