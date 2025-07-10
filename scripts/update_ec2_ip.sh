#!/bin/bash
# update_ec2_ip.sh - Script to update EC2 IP address in Ansible configuration files
# 
# This script:
# 1. Takes an IP address as input (or uses the one from the inventory file)
# 2. Updates the IP in all relevant configuration files
#
# Usage: ./update_ec2_ip.sh [IP_ADDRESS]
#   If IP_ADDRESS is not provided, it will be extracted from the inventory file

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INVENTORY_FILE="${PROJECT_ROOT}/ansible/inventory/hosts.ini"
GROUP_VARS_FILE="${PROJECT_ROOT}/ansible/group_vars/all/main.yml"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

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

echo -e "${GREEN}All files updated successfully with IP: ${PUBLIC_IP}${NC}"
exit 0