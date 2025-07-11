#!/bin/bash
# update_ec2_ip.sh - Script to update EC2 IP address in Ansible configuration files and Cloudflare DNS
# 
# This script:
# 1. Takes an IP address as input (or uses the one from the inventory file)
# 2. Updates the IP in all relevant configuration files
# 3. Updates Cloudflare DNS records with the new IP address
#
# Usage: ./update_ec2_ip.sh [IP_ADDRESS]
#   If IP_ADDRESS is not provided, it will be extracted from the inventory file

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INVENTORY_FILE="${PROJECT_ROOT}/ansible/inventory/hosts.ini"
GROUP_VARS_FILE="${PROJECT_ROOT}/ansible/group_vars/all/main.yml"
ENV_FILE="${PROJECT_ROOT}/.env"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Load environment variables if .env file exists
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
fi

# Get the IP address from command line argument or from inventory file
if [ -n "$1" ]; then
    # Use the provided IP address
    PUBLIC_IP="$1"
    echo -e "${GREEN}Using provided IP address: ${PUBLIC_IP}${NC}"
elif [ -f "$INVENTORY_FILE" ]; then
    # Extract IP from inventory file
    PUBLIC_IP=$(grep -E "^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+" "$INVENTORY_FILE" | head -1 | awk '{print $1}')
    if [ -n "$PUBLIC_IP" ]; then
        echo -e "${GREEN}Using IP address from inventory file: ${PUBLIC_IP}${NC}"
    else
        echo -e "${RED}Error: Could not extract IP address from inventory file.${NC}"
        echo "Please provide an IP address as an argument: ./update_ec2_ip.sh <IP_ADDRESS>"
        exit 1
    fi
else
    echo -e "${RED}Error: No IP address provided and inventory file not found.${NC}"
    echo "Please provide an IP address as an argument: ./update_ec2_ip.sh <IP_ADDRESS>"
    exit 1
fi

# Validate IP address format
if ! [[ $PUBLIC_IP =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid IP address format: ${PUBLIC_IP}${NC}"
    exit 1
fi

# Update inventory file
if [ -f "$INVENTORY_FILE" ]; then
    echo -e "${GREEN}Updating inventory file: ${INVENTORY_FILE}${NC}"
    # Get the SSH key path from the current inventory file
    SSH_KEY_PATH=$(grep "ansible_ssh_private_key_file" "$INVENTORY_FILE" | sed -E 's/.*ansible_ssh_private_key_file=([^ ]*).*/\1/')
    
    # If SSH key path not found, use default
    if [ -z "$SSH_KEY_PATH" ]; then
        SSH_KEY_PATH="~/.ssh/AEGONG.pem"
        echo -e "${YELLOW}Warning: Could not determine SSH key path from inventory file. Using default: ${SSH_KEY_PATH}${NC}"
    fi
    
    # Create new inventory content
    NEW_INVENTORY="[aegong_servers]
$PUBLIC_IP ansible_user=ubuntu ansible_ssh_private_key_file=$SSH_KEY_PATH

# The group name '[aegong_servers]' is used by convention in this project.
# If you are using a specific SSH key for AWS, add 'ansible_ssh_private_key_file'
# and point it to your .pem file as shown above."
    
    # Write to inventory file
    echo "$NEW_INVENTORY" > "$INVENTORY_FILE"
    echo "  Updated inventory file with IP: $PUBLIC_IP"
else
    echo -e "${YELLOW}Warning: Inventory file not found: ${INVENTORY_FILE}${NC}"
    echo "Creating new inventory file..."
    
    # Create directory if it doesn't exist
    mkdir -p "$(dirname "$INVENTORY_FILE")"
    
    # Create new inventory content with default SSH key path
    NEW_INVENTORY="[aegong_servers]
$PUBLIC_IP ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/AEGONG.pem

# The group name '[aegong_servers]' is used by convention in this project.
# If you are using a specific SSH key for AWS, add 'ansible_ssh_private_key_file'
# and point it to your .pem file as shown above."
    
    # Write to inventory file
    echo "$NEW_INVENTORY" > "$INVENTORY_FILE"
    echo "  Created new inventory file with IP: $PUBLIC_IP"
fi

# Update group vars file
if [ -f "$GROUP_VARS_FILE" ]; then
    echo -e "${GREEN}Updating group vars file: ${GROUP_VARS_FILE}${NC}"
    # Replace the ssh_allowed_cidr line
    sed -i "s|ssh_allowed_cidr: \".*\"|ssh_allowed_cidr: \"$PUBLIC_IP/32\"|" "$GROUP_VARS_FILE"
    echo "  Updated ssh_allowed_cidr with IP: $PUBLIC_IP/32"
else
    echo -e "${YELLOW}Warning: Group vars file not found: ${GROUP_VARS_FILE}${NC}"
fi

# Update Cloudflare DNS records
if [ -n "$CLOUDFLARE_API_TOKEN" ] && [ -n "$CLOUDFLARE_ZONE_ID" ] && [ -n "$CLOUDFLARE_RECORD_NAME" ]; then
    echo -e "${GREEN}Updating Cloudflare DNS record for ${CLOUDFLARE_RECORD_NAME}${NC}"
    
    # First, get the DNS record ID
    RECORD_ID=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records?type=A&name=$CLOUDFLARE_RECORD_NAME" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ -n "$RECORD_ID" ]; then
        # Update existing record
        UPDATE_RESPONSE=$(curl -s -X PUT "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records/$RECORD_ID" \
            -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
            -H "Content-Type: application/json" \
            --data "{\"type\":\"A\",\"name\":\"$CLOUDFLARE_RECORD_NAME\",\"content\":\"$PUBLIC_IP\",\"ttl\":60,\"proxied\":false}")
        
        if echo "$UPDATE_RESPONSE" | grep -q '"success":true'; then
            echo -e "  ${GREEN}Successfully updated Cloudflare DNS record to point to ${PUBLIC_IP}${NC}"
        else
            echo -e "  ${RED}Failed to update Cloudflare DNS record: ${UPDATE_RESPONSE}${NC}"
        fi
    else
        # Create new record
        CREATE_RESPONSE=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
            -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
            -H "Content-Type: application/json" \
            --data "{\"type\":\"A\",\"name\":\"$CLOUDFLARE_RECORD_NAME\",\"content\":\"$PUBLIC_IP\",\"ttl\":60,\"proxied\":false}")
        
        if echo "$CREATE_RESPONSE" | grep -q '"success":true'; then
            echo -e "  ${GREEN}Successfully created Cloudflare DNS record pointing to ${PUBLIC_IP}${NC}"
        else
            echo -e "  ${RED}Failed to create Cloudflare DNS record: ${CREATE_RESPONSE}${NC}"
        fi
    fi
else
    echo -e "${YELLOW}Warning: Cloudflare API configuration not found in .env file.${NC}"
    echo -e "${YELLOW}To update Cloudflare DNS records, add the following to your .env file:${NC}"
    echo -e "${YELLOW}CLOUDFLARE_API_TOKEN=\"your-api-token\"${NC}"
    echo -e "${YELLOW}CLOUDFLARE_ZONE_ID=\"your-zone-id\"${NC}"
    echo -e "${YELLOW}CLOUDFLARE_RECORD_NAME=\"your-domain-name\"${NC}"
fi

echo -e "${GREEN}All files updated successfully with IP: ${PUBLIC_IP}${NC}"
exit 0