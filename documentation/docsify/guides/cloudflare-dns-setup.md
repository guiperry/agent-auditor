# Cloudflare DNS Update for EC2 Instances

This document explains how to set up automatic Cloudflare DNS updates for your EC2 instance IP address.

## Overview

When your EC2 instance starts, its public IP address may change. This solution automatically updates your Cloudflare DNS records to point to the new IP address, ensuring your domain always resolves to the correct server.

## Setup Instructions

### 1. Cloudflare API Configuration

1. Log in to your Cloudflare account
2. Navigate to "My Profile" > "API Tokens"
3. Create a new API token with the following permissions:
   - Zone - DNS - Edit
   - Zone - Zone - Read
4. Select the specific zone (domain) you want to manage
5. Copy the generated API token

### 2. Update Environment Variables

Add the following variables to your `.env` file:

```
CLOUDFLARE_API_TOKEN="your-api-token"
CLOUDFLARE_ZONE_ID="your-zone-id"
CLOUDFLARE_RECORD_NAME="your-domain-name.com"
```

Where:
- `your-api-token` is the token you generated in step 1
- `your-zone-id` is your Cloudflare zone ID (found in the Cloudflare dashboard URL when viewing your domain)
- `your-domain-name.com` is the domain or subdomain you want to point to your EC2 instance

### 3. Add GitHub Secrets (for GitHub Actions)

If you're using GitHub Actions for automation, add these secrets to your repository:

1. Go to your GitHub repository
2. Navigate to Settings > Secrets and variables > Actions
3. Add the following repository secrets:
   - `CLOUDFLARE_API_TOKEN`
   - `CLOUDFLARE_ZONE_ID`
   - `CLOUDFLARE_RECORD_NAME`
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_REGION`

### 4. Manual Update

To manually update your Cloudflare DNS record with your EC2 instance IP:

```bash
# If you know the IP address
./scripts/update_ec2_ip.sh 123.456.789.012

# To retrieve the IP from EC2 and update automatically
ansible-playbook ansible/playbook.yml --tags "ec2_management,update_dns"
```

### 5. Manual Updates via GitHub Actions

The system includes a GitHub workflow that can be manually triggered to update your Cloudflare DNS records:

- Trigger the workflow from the GitHub Actions UI when needed
- Particularly useful after starting your EC2 instance or when you know the IP has changed

## How It Works

1. The `update_ec2_ip.sh` script retrieves the EC2 instance IP address
2. It uses the Cloudflare API to update the DNS record
3. The GitHub workflow automates this process on a schedule

## Troubleshooting

If DNS updates are not working:

1. Check that your Cloudflare API token has the correct permissions
2. Verify your zone ID is correct
3. Ensure your domain name is specified correctly
4. Check the GitHub Actions logs for any error messages
5. Run the update script manually to see detailed output

## Integration with EC2 Startup

For the most efficient solution, consider integrating the DNS update with your EC2 instance startup process:

1. The current implementation updates DNS when you manually run the GitHub workflow
2. The Ansible playbook already includes this functionality when starting the EC2 instance
3. For a fully automated solution, you could add the DNS update script to your EC2 instance's user data or startup scripts

---

<div class="footer-links">
<a href="#/legal/CODE_OF_CONDUCT.md" class="footer-link">Contributor Covenant Code of Conduct</a> | <a href="#/legal/PRIVACY_POLICY.md" class="footer-link">PRIVACY_POLICY.md</a> | <a href="#/legal/TERMS_AND_CONDITIONS.md" class="footer-link">TERMS AND CONDITIONS</a>

© 2025 Agent Auditor
</div>
