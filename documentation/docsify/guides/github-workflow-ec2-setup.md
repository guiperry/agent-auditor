# GitHub Workflow-Based EC2 Instance Management

This document explains how to set up and use the GitHub workflow-based approach for starting your EC2 instance and updating Cloudflare DNS.

## Overview

Instead of directly using AWS credentials in your Netlify functions, this approach uses GitHub Actions workflows to:

1. Start your EC2 instance
2. Update Cloudflare DNS with the instance's IP address
3. Return the instance details to your loader application

This approach has several advantages:
- Avoids AWS credential issues in Netlify functions
- Uses GitHub's secure secrets management
- Leverages existing GitHub Actions workflows that are already working
- Simplifies permission management

## Setup Instructions

### 1. GitHub Personal Access Token

1. Go to your GitHub account settings
2. Navigate to Developer settings > Personal access tokens > Tokens (classic)
3. Generate a new token with the following permissions:
   - `repo` (Full control of private repositories)
   - `workflow` (Update GitHub Action workflows)
4. Copy the generated token

### 2. Netlify Environment Variables

Add the following environment variables in your Netlify dashboard:

1. Go to Site settings > Environment variables
2. Add the following variables:
   - `GITHUB_TOKEN`: Your personal access token
   - `GITHUB_OWNER`: Your GitHub username or organization name
   - `GITHUB_REPO`: The repository name (default: agent-auditor)
   - `GITHUB_WORKFLOW_ID`: The workflow file name (default: start-ec2-instance.yml)
   - `GITHUB_REF`: The branch or tag to use (default: main)

### 3. GitHub Secrets

Ensure the following secrets are set in your GitHub repository:

1. Go to your repository settings
2. Navigate to Secrets and variables > Actions
3. Add the following secrets:
   - `AWS_ACCESS_KEY_ID`: Your AWS access key
   - `AWS_SECRET_ACCESS_KEY`: Your AWS secret key
   - `AWS_REGION`: Your AWS region (e.g., us-east-2)
   - `EC2_INSTANCE_ID`: The ID of your EC2 instance (e.g., i-0123456789abcdef0)
   - `SSH_PRIVATE_KEY`: Your SSH private key for connecting to the EC2 instance
   - `SSH_KNOWN_HOSTS`: SSH known hosts file content
   - `ANSIBLE_VAULT_PASSWORD`: Password for Ansible vault
   - `CLOUDFLARE_API_TOKEN`: Your Cloudflare API token
   - `CLOUDFLARE_ZONE_ID`: Your Cloudflare zone ID
   - `CLOUDFLARE_RECORD_NAME`: Your domain name

## How It Works

1. **User Interaction**:
   - User visits the loader page and starts building their robot
   - The page calls the `github-start` Netlify function

2. **GitHub Workflow Trigger**:
   - The Netlify function triggers the GitHub workflow using the GitHub API
   - The workflow starts the EC2 instance and updates Cloudflare DNS

3. **Status Monitoring**:
   - The loader page periodically calls the `github-status` Netlify function
   - The function checks the workflow status using the GitHub API
   - When the workflow completes, the function returns the instance details

4. **Redirection**:
   - Once the instance is ready, the loader page redirects to the instance's IP address

## Troubleshooting

### GitHub Workflow Issues

If the GitHub workflow fails to start:

1. Check that your GitHub token has the correct permissions
2. Verify that the repository and workflow file exist
3. Ensure the token has permission to trigger workflows
4. Check the GitHub Actions logs for detailed error messages

### Netlify Function Issues

If the Netlify functions fail:

1. Check the Netlify function logs in the Netlify dashboard
2. Verify that all environment variables are set correctly
3. Ensure the functions have the necessary dependencies

### EC2 Instance Issues

If the EC2 instance fails to start:

1. Check the GitHub Actions logs for AWS errors
2. Verify that your AWS credentials have the necessary permissions
3. Ensure the EC2 instance exists and is in a stoppable state

## Additional Configuration

### Package Dependencies

Make sure your Netlify functions have the necessary dependencies by adding them to your `package.json`:

```json
{
  "dependencies": {
    "axios": "^0.24.0",
    "uuid": "^8.3.2"
  }
}
```

### Custom Domain Configuration

If you're using a custom domain with Cloudflare:

1. Update the `CLOUDFLARE_RECORD_NAME` in your GitHub secrets
2. Ensure your Cloudflare API token has the necessary permissions
3. Test the DNS update by manually triggering the workflow

---

<div class="footer-links">
<a href="#/legal/CODE_OF_CONDUCT.md" class="footer-link">Contributor Covenant Code of Conduct</a> | <a href="#/legal/PRIVACY_POLICY.md" class="footer-link">PRIVACY_POLICY.md</a> | <a href="#/legal/TERMS_AND_CONDITIONS.md" class="footer-link">TERMS AND CONDITIONS</a>

© 2025 Agent Auditor
</div>
