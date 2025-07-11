#!/bin/bash

# sync_voice_config.sh
# Script to synchronize voice_config.json with Ansible template
# This ensures the Ansible template stays in sync with the local configuration

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# File paths
SOURCE_FILE="voice_config.json"
TEMPLATE_FILE="ansible/roles/agent_auditor/templates/voice_config.json.j2"
BACKUP_DIR="ansible/roles/agent_auditor/templates/backups"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to create backup
create_backup() {
    if [[ -f "$TEMPLATE_FILE" ]]; then
        mkdir -p "$BACKUP_DIR"
        local backup_file="$BACKUP_DIR/voice_config.json.j2.backup.$(date +%Y%m%d_%H%M%S)"
        cp "$TEMPLATE_FILE" "$backup_file"
        print_info "Created backup: $backup_file"
    fi
}

# Function to convert JSON to Jinja2 template
convert_to_template() {
    local source_file="$1"
    local template_file="$2"
    
    print_info "Converting $source_file to Jinja2 template format..."
    
    # Read the source JSON and convert to template format
    python3 -c "
import json
import sys

# Mapping of JSON keys to Ansible variables
key_mappings = {
    'enabled': 'voice_enabled | to_json',
    'provider': 'voice_provider',
    'key_file': 'voice_key_file_dest',
    'key_pass_env': 'voice_key_pass_env_var',
    'output_dir': 'voice_output_dir',
    'default_voice': 'voice_default_voice',
    'ws_url': 'voice_ws_url | default(\\'wss://your-livekit-instance.example.com\\')'
}

try:
    with open('$source_file', 'r') as f:
        config = json.load(f)
    
    # Generate Jinja2 template
    template_lines = ['{']
    
    for i, (key, value) in enumerate(config.items()):
        if key in key_mappings:
            ansible_var = key_mappings[key]
            if key == 'enabled':
                # Boolean values need special handling
                template_lines.append(f'    \"{key}\": {{{{ {ansible_var} }}}},')
            else:
                # String values
                template_lines.append(f'    \"{key}\": \"{{{{ {ansible_var} }}}}\",')
        else:
            # Fallback for unmapped keys
            if isinstance(value, bool):
                template_lines.append(f'    \"{key}\": {str(value).lower()},')
            elif isinstance(value, str):
                template_lines.append(f'    \"{key}\": \"{value}\",')
            else:
                template_lines.append(f'    \"{key}\": {json.dumps(value)},')
    
    # Remove trailing comma from last item
    if template_lines[-1].endswith(','):
        template_lines[-1] = template_lines[-1][:-1]
    
    template_lines.append('}')
    
    # Write template file
    with open('$template_file', 'w') as f:
        f.write('\n'.join(template_lines) + '\n')
    
    print('Template conversion completed successfully')
    
except Exception as e:
    print(f'Error: {e}', file=sys.stderr)
    sys.exit(1)
"
}

# Function to validate JSON
validate_json() {
    local file="$1"
    if ! python3 -m json.tool "$file" > /dev/null 2>&1; then
        print_error "Invalid JSON in $file"
        return 1
    fi
    return 0
}

# Function to show differences
show_differences() {
    if [[ -f "$TEMPLATE_FILE" ]]; then
        print_info "Showing differences between current template and new version:"
        echo "----------------------------------------"
        
        # Create temporary file with new template content
        local temp_file=$(mktemp)
        convert_to_template "$SOURCE_FILE" "$temp_file"
        
        # Show diff
        if command -v colordiff >/dev/null 2>&1; then
            colordiff -u "$TEMPLATE_FILE" "$temp_file" || true
        else
            diff -u "$TEMPLATE_FILE" "$temp_file" || true
        fi
        
        rm -f "$temp_file"
        echo "----------------------------------------"
    fi
}

# Main function
main() {
    print_info "Voice Config Synchronization Script"
    print_info "===================================="
    
    # Check if source file exists
    if [[ ! -f "$SOURCE_FILE" ]]; then
        print_error "Source file $SOURCE_FILE not found!"
        print_info "Please ensure you're running this script from the project root directory."
        exit 1
    fi
    
    # Validate source JSON
    print_info "Validating source JSON file..."
    if ! validate_json "$SOURCE_FILE"; then
        exit 1
    fi
    print_success "Source JSON is valid"
    
    # Check if template directory exists
    if [[ ! -d "$(dirname "$TEMPLATE_FILE")" ]]; then
        print_error "Template directory $(dirname "$TEMPLATE_FILE") not found!"
        exit 1
    fi
    
    # Show current configuration
    print_info "Current voice_config.json content:"
    echo "----------------------------------------"
    cat "$SOURCE_FILE"
    echo "----------------------------------------"
    
    # Show differences if template exists
    if [[ -f "$TEMPLATE_FILE" ]]; then
        show_differences
    fi
    
    # Ask for confirmation
    echo
    read -p "Do you want to update the Ansible template? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Operation cancelled by user"
        exit 0
    fi
    
    # Create backup
    create_backup
    
    # Convert and write template
    convert_to_template "$SOURCE_FILE" "$TEMPLATE_FILE"
    
    print_success "Template updated successfully!"
    print_info "Updated file: $TEMPLATE_FILE"
    
    # Show final template content
    print_info "Final template content:"
    echo "----------------------------------------"
    cat "$TEMPLATE_FILE"
    echo "----------------------------------------"
    
    print_success "Voice config synchronization completed!"
    print_info "You can now deploy with: make deploy"
}

# Run main function
main "$@"
