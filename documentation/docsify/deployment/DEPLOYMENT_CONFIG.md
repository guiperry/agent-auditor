# Deployment Config - Git-Based Version Management & Automation

## Overview

This document summarizes the configuration made to support automated git-based version management for production deployments, AWS port configuration, and voice configuration synchronization.

## Changes Made

### 1. Git-Based Version Management

#### **Makefile Deploy Target Enhancement**
- **Automatic Version Detection**: Deploy target now automatically detects current git tag or commit SHA
- **Dynamic app_version Update**: Updates `app_version` in Ansible defaults before deployment
- **Safe File Handling**: Creates backup and restores original defaults file after deployment
- **Production-Ready**: Prioritizes git tags for releases, falls back to commit SHA for development

#### **New Makefile Targets**:
- `make version` - Shows current git version (tag or commit SHA)
- `make test-deploy` - Tests version update process without running Ansible (dry run)
- `make deploy` - Enhanced to auto-update version from git and deploy

#### **Version Detection Logic**:
```bash
# Prioritizes git tags, falls back to commit SHA
GIT_VERSION=$(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
```

### 2. Port Configuration Updates

#### **main.go**
- **Changed default port**: From `8084` to `8080` (more standard for web applications)
- **Added environment variable support**: Application now reads `PORT` environment variable
- **Configurable port binding**: `http.ListenAndServe(":"+port, r)`

#### **Makefile**
- Updated help text and run target to reflect new default port `8080`
- Added `sync-voice-config` target for template synchronization

### 3. Ansible Configuration Updates

#### **ansible/roles/agent_auditor/defaults/main.yml**
- **Updated `app_version`**: Changed from `repo_branch` to `app_version` for git tag/commit SHA support
- **Removed hardcoded network settings**: Port configuration moved to environment variables
- **Streamlined configuration**: Focused on core application settings

#### **ansible/roles/agent_auditor/tasks/deploy_app.yml**
- **Updated git checkout**: Now uses `app_version` instead of `repo_branch`
- **Enhanced permissions**: Added proper `become_user` configuration
- **Version-based deployment**: Supports both git tags and commit SHAs

### 4. Legacy AWS Firewall Configuration (Previously Implemented)

#### **Note**: The following firewall configuration was previously implemented but may have been removed in recent updates:

#### **ansible/roles/agent_auditor/tasks/configure_firewall.yml** (Legacy)
- **UFW Configuration**: Automatically configures Ubuntu firewall
- **AWS Security Group Guidance**: Provides detailed instructions for AWS EC2
- **Instance Detection**: Automatically detects if running on AWS EC2

*Note: Port configuration is now handled via environment variables in the application itself.*

### 5. Voice Configuration Synchronization

#### **New File: sync_voice_config.sh**
- **Automated Sync**: Converts `voice_config.json` to Ansible template format
- **Backup Creation**: Creates timestamped backups before updates
- **Validation**: Validates JSON syntax before processing
- **Diff Display**: Shows differences between current and new template
- **Interactive Confirmation**: Asks user before making changes
- **Jinja2 Conversion**: Maps JSON keys to appropriate Ansible variables

#### **Key Mappings**:
```bash
enabled -> voice_enabled | to_json
provider -> voice_provider
key_file -> voice_key_file_dest
key_pass_env -> voice_key_pass_env_var
output_dir -> voice_output_dir
default_voice -> voice_default_voice
ws_url -> voice_ws_url | default('wss://your-livekit-instance.example.com')
```

## Usage Instructions

### 1. Git-Based Version Management

#### **Production Deployment with Git Tags**:
```bash
# Create and push a release tag
git tag v1.0.3
git push origin v1.0.3

# Deploy with automatic version detection
make deploy
# This will:
# 1. Detect git tag "v1.0.3"
# 2. Update app_version: "v1.0.3" in Ansible defaults
# 3. Deploy with Ansible
# 4. Restore original defaults file
```

#### **Development Deployment with Commit SHA**:
```bash
# Deploy current commit (no tag required)
make deploy
# This will:
# 1. Detect commit SHA (e.g., "cceaf1a")
# 2. Update app_version: "cceaf1a" in Ansible defaults
# 3. Deploy with Ansible
# 4. Restore original defaults file
```

#### **Test Version Update Process**:
```bash
# Dry run to test version detection and file updates
make test-deploy
```

#### **Check Current Version**:
```bash
# Show current git version that would be deployed
make version
```

### 2. Local Development
```bash
# Build and run with default port (8080)
make run

# Run with custom port
PORT=8084 ./aegong
```

### 3. Voice Configuration Sync
```bash
# Sync voice_config.json to Ansible template
make sync-voice-config
# OR
./sync_voice_config.sh
```

### 4. AWS Deployment

#### **Before Deployment**:
1. **Create a git tag for production releases**:
   ```bash
   git tag v1.0.3
   git push origin v1.0.3
   ```

2. **Sync voice configuration** (if needed):
   ```bash
   make sync-voice-config
   ```

3. **Ensure AWS Security Group allows**:
   - SSH (port 22)
   - Application port (8080 or configured port)

#### **Deploy**:
```bash
# Deploy with automatic version detection
make deploy

# Or test the process first
make test-deploy
```

#### **Post-Deployment**:
- Application will be accessible at: `http://YOUR_EC2_IP:8080`
- Version deployed matches your git tag/commit
- Service runs with proper environment configuration

### 4. AWS Security Group Setup

#### **Automated AWS Security Group Management**

To make your deployments fully automated, you can use Ansible to manage the AWS Security Group directly. This ensures that the correct ports are always open without manual intervention.

##### **Prerequisites**:

1. **Install the AWS Collection**:
   ```bash
   pip install ansible boto3 botocore
   ansible-galaxy collection install amazon.aws
   ```

2. **Configure AWS Credentials**:
   
   Your AWS Access Key ID and Secret Access Key are your permanent credentials. Never hardcode them in your code, commit them to Git, or share them publicly.

   **How to Create and Find Your Access Keys**:
   1. Sign in to the AWS Management Console
   2. Navigate to the IAM service (type "IAM" in the top search bar)
   3. Go to "Users" in the left-hand menu and click on your IAM user name (e.g., gperry)
      - *Security Note*: It's best practice to use an IAM user for this, not your root account
   4. Go to the "Security credentials" tab
   5. Scroll down to the "Access keys" section and click "Create access key"
   6. Choose "Command Line Interface (CLI)" as the use case and click "Next"
   7. (Optional) Set a description tag (e.g., "Ansible Control Node Key") and click "Create access key"
   8. **IMPORTANT**: Save both the Access key ID and the Secret access key (this is the only time you will see the Secret access key)
      - You can click the "Download .csv file" button to save them securely

   **Configure the Keys for Ansible**:
   ```bash
   # Install AWS CLI if needed
   sudo apt-get update && sudo apt-get install awscli -y
   
   # Configure AWS credentials
   aws configure
   
   # Enter your credentials when prompted:
   # AWS Access Key ID [None]: YOUR_ACCESS_KEY_ID_HERE
   # AWS Secret Access Key [None]: YOUR_SECRET_ACCESS_KEY_HERE
   # Default region name [None]: us-east-2
   # Default output format [None]: json
   ```

3. **Add the Security Group Task to Your Ansible Playbook**:

   Add the following task to your Ansible deployment playbook:

   ```yaml
   - name: Ensure security group exists
     amazon.aws.ec2_security_group:
       name: aegong-sg
       description: Agent Auditor Security Group
       vpc_id: "{{ vpc_id }}"
       region: "{{ aws_region | default(ansible_ec2_placement_region) }}"
       rules:
         - proto: tcp
           ports:
             - 22
           cidr_ip: "{{ ssh_allowed_cidr | default('0.0.0.0/0') }}"
           rule_desc: Allow SSH access
         - proto: tcp
           ports:
             - 8080
           cidr_ip: 0.0.0.0/0
           rule_desc: Allow application access
     register: security_group
   ```

4. **Required Variables**:
   - `vpc_id`: The ID of the VPC where the security group should be created
   - `aws_region`: The AWS region (defaults to the instance's region if running on EC2)
   - `ssh_allowed_cidr` (Optional): Restrict SSH access to a specific IP range for better security

#### **Manual Setup** (if not using Ansible automation):
1. Go to AWS Console > EC2 > Security Groups
2. Create new Security Group: `aegong-sg`
3. Add Inbound Rules:
   - SSH: TCP/22 from your IP
   - Custom TCP: TCP/8080 from 0.0.0.0/0 (or specific IPs)

#### **AWS CLI Setup** (alternative to Ansible):
```bash
aws ec2 create-security-group --group-name aegong-sg --description "Agent Auditor Security Group"
aws ec2 authorize-security-group-ingress --group-name aegong-sg --protocol tcp --port 22 --cidr 0.0.0.0/0
aws ec2 authorize-security-group-ingress --group-name aegong-sg --protocol tcp --port 8080 --cidr 0.0.0.0/0
```

## Benefits

1. **Git-Based Version Control**: Automatic version detection from git tags/commits
2. **Production-Ready Releases**: Tag-based deployments for proper version management
3. **Development Flexibility**: Commit SHA deployments for development/testing
4. **Safe Deployment Process**: Backup and restore of configuration files
5. **Traceability**: Clear mapping between deployed versions and git history
6. **Automated Workflow**: No manual version editing required
7. **Flexible Port Configuration**: Easy to adapt to different environments
8. **Automated Sync**: No manual template editing required
9. **Documentation**: Clear guidance for git-based deployment workflow

## Files Modified

### **Core Application**
- `main.go` - Port configuration via environment variables
- `Makefile` - **MAJOR UPDATE** - Enhanced deploy target with git version management

### **Ansible Configuration**
- `ansible/roles/agent_auditor/defaults/main.yml` - **UPDATED** - Changed to app_version-based deployment
- `ansible/roles/agent_auditor/tasks/deploy_app.yml` - **UPDATED** - Uses app_version instead of repo_branch
- `ansible/roles/agent_auditor/templates/voice_config.json.j2` - Enhanced voice configuration template

### **New Files**
- `sync_voice_config.sh` - **NEW** - Voice config synchronization script
- `docs/DEPLOYMENT_UPDATES.md` - **NEW** - This documentation file

### **Key Makefile Changes**
- **Enhanced `deploy` target**: Automatic git version detection and app_version updates
- **New `version` target**: Shows current git version (tag or commit SHA)
- **New `test-deploy` target**: Dry run testing of version update process
- **Updated help text**: Reflects new git-based deployment workflow

## Deployment Workflow Examples

### **Production Release Workflow**:
```bash
# 1. Prepare release
git add .
git commit -m "Release v1.0.3: Add new features"

# 2. Create and push tag
git tag v1.0.3
git push origin main
git push origin v1.0.3

# 3. Deploy to production
make deploy
# Output: 📋 Current version: v1.0.3
#         ✅ Updated app_version to: v1.0.3
#         🚀 Deploying application with Ansible...
```

### **Development/Testing Workflow**:
```bash
# 1. Make changes and commit
git add .
git commit -m "Fix bug in authentication"

# 2. Deploy development version
make deploy
# Output: 📋 Current version: a1b2c3d
#         ✅ Updated app_version to: a1b2c3d
#         🚀 Deploying application with Ansible...
```

### **Safe Testing Workflow**:
```bash
# Test version detection without deploying
make version
make test-deploy

# If everything looks good, deploy
make deploy
```

## Next Steps

1. **Test git tag-based deployment** on AWS EC2 instance
2. **Verify version tracking** in deployed application logs
3. **Confirm rollback capability** using previous git tags
4. **Test voice configuration** with actual TTS providers
5. **Document production release process** for team members
