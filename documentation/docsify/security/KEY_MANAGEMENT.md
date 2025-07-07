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
go build -o generate-keys ./cmd/generate_keys/main.go

# Run the utility to create a new key file
./generate-keys -output default.key

# Alternatively, use the Makefile target
make keys
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
- `ansible_vault_password` - Password for Ansible Vault encryption

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
go build -o test-keys ./cmd/test_keys/main.go

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
6. **GitHub Actions Secrets**: Store deployment credentials only as GitHub repository secrets
7. **Limit Secret Access**: Restrict which workflows and branches can access sensitive secrets
8. **Audit Secret Usage**: Regularly review which workflows are using secrets and for what purpose

## CI/CD Secrets Management

For automated deployments using GitHub Actions, several secrets need to be configured in your repository settings:

### GitHub Repository Secrets

Configure these secrets in your GitHub repository (Settings > Secrets and variables > Actions):

1. **SSH_PRIVATE_KEY**: The private SSH key used to connect to deployment servers
2. **SSH_KNOWN_HOSTS**: The SSH known hosts file content for secure connections
3. **ANSIBLE_VAULT_PASSWORD**: The password used to decrypt Ansible Vault encrypted files

### Setting Up GitHub Actions Secrets

1. Go to your GitHub repository
2. Navigate to Settings > Secrets and variables > Actions
3. Click "New repository secret"
4. Add each of the required secrets with their appropriate values

### Using Secrets in Workflows

The secrets are referenced in the GitHub workflow file (`.github/workflows/build-and-deploy.yml`) and are automatically made available to the workflow during execution:

```yaml
- name: Configure SSH
  env:
    SSH_PRIVATE_KEY: ${{ secrets.SSH_PRIVATE_KEY }}
    SSH_KNOWN_HOSTS: ${{ secrets.SSH_KNOWN_HOSTS }}
  run: |
    # SSH configuration commands

- name: Deploy with Ansible
  env:
    ANSIBLE_VAULT_PASSWORD: ${{ secrets.ANSIBLE_VAULT_PASSWORD }}
  run: |
    # Ansible deployment commands
```

### Ansible Vault Encryption

Ansible Vault provides an additional layer of security for sensitive deployment variables. Here's how it works in the project:

#### How and When the Ansible Vault File Gets Encrypted

The Ansible Vault file (typically located at `ansible/group_vars/all/vault.yml`) is encrypted once during the initial setup of your deployment configuration and remains encrypted on disk and in your Git repository.

##### Creating a New Encrypted Vault File

To create a new encrypted vault file:

```bash
ansible-vault create ansible/group_vars/all/vault.yml
```

When you run this command:
1. Ansible prompts you to create a new vault password
2. You enter and confirm this password
3. Ansible opens your default text editor
4. You add your secrets in plain YAML format (e.g., `vault_aegong_key_pass: "your-secret-passphrase"`)
5. When you save and close the editor, Ansible encrypts the file's contents with AES256 encryption

##### Encrypting an Existing Vault File

If you already have an unencrypted vault file, you can encrypt it in-place:

```bash
ansible-vault encrypt ansible/group_vars/all/vault.yml
```

When you run this command:
1. Ansible prompts you to create a new vault password
2. You enter and confirm this password
3. It then encrypts the existing file in-place

This is the simplest way to secure an existing plain text vault file without having to recreate it.

For your specific project, you would run:

```bash
ansible-vault encrypt /home/gperry/Documents/GitHub/agent-auditor/ansible/group_vars/all/vault.yml
```

This will encrypt your existing vault file that contains:
```yaml
vault_aegong_key_pass: "the-correct-secret-passphrase"
```

From this point on, the `vault.yml` file contains only encrypted ciphertext, not plain text.

#### How Ansible Vault is Used in the Project

The vault file is only decrypted in memory by Ansible at runtime:

- **Local Deployment**: When running `make deploy`, the Makefile uses the `--ask-vault-pass` flag, prompting you for the vault password in your terminal
- **Automated Deployment**: The GitHub Actions workflow provides the password from the `ANSIBLE_VAULT_PASSWORD` secret, avoiding interactive prompts and keeping the secret out of logs

To edit an existing encrypted vault file:

```bash
ansible-vault edit ansible/group_vars/all/vault.yml
```

This command handles the temporary decryption and re-encryption safely, without leaving unencrypted data on disk.

## Troubleshooting

If you encounter issues with the key management system:

1. **Keys Not Loading**: Ensure the passphrase environment variable is correctly set
2. **Decryption Errors**: Verify you're using the correct passphrase
3. **File Not Found**: Check that the key file path in `voice_config.json` is correct
4. **Permission Denied**: Ensure the application has read access to the key file
5. **CI/CD Deployment Failures**: Verify that all required GitHub secrets are properly configured
6. **Environment Variable Configuration**: 
   - Confirm that the `key_pass_env` in voice_config.json matches the environment variable name you're setting
   - For this project, ensure `AEGONG_KEY_PASS` is set with the value `AEGONGIsTheAgentGuardian`
7. **Ansible Vault Errors**: 
   - For "Decryption failed" errors, check that the correct vault password is being used
   - For local deployments, ensure you're entering the correct password when prompted
   - For CI/CD deployments, verify the ANSIBLE_VAULT_PASSWORD secret is correctly set in GitHub

For additional help, run the application with verbose logging enabled.