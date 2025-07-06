# Secure API Key Management

Aegong includes a secure key management system for storing and loading API keys during runtime. This system encrypts sensitive API keys and credentials, allowing them to be securely stored in version control while requiring a passphrase to decrypt them at runtime.

## Key Features

- **AES-256-GCM Encryption**: Industry-standard encryption for all stored keys
- **Passphrase-based Security**: Keys can only be accessed with the correct passphrase
- **Environment Variable Integration**: Passphrase is loaded from environment variables for security
- **Multiple Provider Support**: Store keys for various TTS providers (OpenAI, Cerebras, Google, Azure, Cartesia)
- **Runtime Decryption**: Keys are only decrypted in memory during runtime

## Setup Instructions

### 1. Generate an Encrypted Key File

Use the included utility to create an encrypted key file:

```bash
# Compile the key generation utility
go build -o generate-keys generate_key_file.go key_manager.go

# Run the utility to create a new key file
./generate-keys -output default.key
```

You will be prompted to:
1. Enter a secure passphrase (this will be needed to decrypt the keys)
2. Confirm the passphrase
3. Enter key-value pairs for your API keys

Example key names to include:
- `openai` - OpenAI API key
- `cerebras` - Cerebras API key
- `google_credentials_path` - Path to Google Cloud credentials JSON file
- `azure` - Azure Speech API key
- `azure_region` - Azure region (e.g., "eastus")
- `cartesia` - Cartesia API key
- `livekit_api_key` - LiveKit API key
- `livekit_api_secret` - LiveKit API secret

### 2. Configure the Voice Integration

Update your `voice_config.json` file to use the key file:

```json
{
    "enabled": true,
    "provider": "openai",
    "key_file": "default.key",
    "key_pass_env": "AEGONG_KEY_PASS",
    "output_dir": "voice_reports",
    "default_voice": "alloy",
    "ws_url": "wss://your-livekit-instance.example.com"
}
```

Configuration options:
- `enabled`: Set to `true` to enable voice reports
- `provider`: TTS provider to use (`openai`, `cerebras`, `google`, `azure`, `cartesia`, or `livekit`)
- `key_file`: Path to your encrypted key file
- `key_pass_env`: Name of the environment variable containing the decryption passphrase
- `output_dir`: Directory to store generated voice reports
- `default_voice`: Default voice to use (provider-specific)
- `ws_url`: WebSocket URL for LiveKit integration

### 3. Set the Passphrase Environment Variable

Before running the application, set the passphrase environment variable:

```bash
# Linux/macOS
export AEGONG_KEY_PASS="your-secure-passphrase"

# Windows (PowerShell)
$env:AEGONG_KEY_PASS="your-secure-passphrase"
```

For production environments, consider using a secure method to set this environment variable, such as:
- Docker secrets
- Kubernetes secrets
- Cloud provider secret management services

### 4. Test the Key Manager

You can test that your keys are correctly stored and can be retrieved:

```bash
# Compile the test utility
go build -o test-keys test_key_manager.go key_manager.go

# List all available keys
./test-keys -key-file default.key -list

# Retrieve a specific key
./test-keys -key-file default.key -key-name openai
```

## Security Best Practices

1. **Use a Strong Passphrase**: Choose a long, complex passphrase that is difficult to guess
2. **Protect the Passphrase**: Never store the passphrase in plain text or commit it to version control
3. **Restrict Key File Access**: Set appropriate file permissions on the key file (e.g., `chmod 600 default.key`)
4. **Regular Key Rotation**: Periodically update your API keys and regenerate the key file
5. **Backup Securely**: Keep secure backups of your key file and passphrase

## Troubleshooting

If you encounter issues with the key management system:

1. **Keys Not Loading**: Ensure the passphrase environment variable is correctly set
2. **Decryption Errors**: Verify you're using the correct passphrase
3. **File Not Found**: Check that the key file path in `voice_config.json` is correct
4. **Permission Denied**: Ensure the application has read access to the key file

For additional help, run the application with verbose logging enabled.