# Use bash for all shell commands to ensure consistency.
SHELL := /bin/bash

# Project variables
BINARY_NAME=aegong
ANSIBLE_DIR=ansible
ANSIBLE_INVENTORY=$(ANSIBLE_DIR)/inventory/hosts.ini
ANSIBLE_PLAYBOOK=$(ANSIBLE_DIR)/playbook.yml

# Suppress "Entering directory" messages and hide commands by default.
.SILENT:

# Phony targets don't represent files.
.PHONY: help all build run test keys deploy clean sync-voice-config version test-deploy generate-docs

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Development Targets:"
	@echo "  build              Build the Go application binary."
	@echo "  run                Build and run the Go application locally on port 8080."
	@echo "  test               Run all Go tests."
	@echo "  keys               Generate a new encrypted API key file (default.key)."
	@echo "  sync-voice-config  Sync voice_config.json to Ansible template."
	@echo "  version            Show current git version (tag or commit SHA)."
	@echo "  clean              Remove the built binary and other generated files."
	@echo "  generate-docs      Generate documentation from docs folder."
	@echo ""
	@echo "Deployment Targets:"
	@echo "  deploy             Deploy with auto-updated version from git tag/commit."
	@echo "  test-deploy        Test the deploy version update process (dry run)."

all: build

generate-docs:
	@echo "🔄 Generating documentation from docs folder..."
	@node doc_generator.js
	@echo "✅ Documentation generated in documentation/ folder"

build: generate-docs
	@echo "Building Aegong Agent Auditor with embedded assets..."
	@echo "📦 Embedding: static/*, documentation/docsify/*, voice_inference.py, requirements.txt"
	go build -o $(BINARY_NAME) .
	@echo "✅ Build complete: ./$(BINARY_NAME) (single binary with embedded assets and documentation)"

run: build
	@echo "Starting Aegong Agent Auditor locally on http://localhost:8080"
	./$(BINARY_NAME)

test:
	@echo "🧪 Running Go tests..."
	GO_TEST=1 go test -v ./...
	@echo "✅ Tests passed."

keys:
	@echo "Generating new encrypted key file..."
	@go build -o generate-keys key_manager.go generate_key_file.go
	@./generate-keys -output default.key
	@echo "✅ New 'default.key' created. Place it in '$(ANSIBLE_DIR)/roles/agent_auditor/files/' before deploying."
	@rm generate-keys

sync-voice-config:
	@echo "Synchronizing voice_config.json with Ansible template..."
	./sync_voice_config.sh

version:
	@echo "Current git version:"
	@GIT_VERSION=$$(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD); \
	echo "📋 Version: $$GIT_VERSION"; \
	if git describe --tags --exact-match >/dev/null 2>&1; then \
		echo "🏷️  Type: Git tag"; \
	else \
		echo "🔗 Type: Commit SHA"; \
	fi

test-deploy:
	@echo "🧪 Testing deployment version update process..."
	@# Get current git tag or commit SHA
	@GIT_VERSION=$$(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD); \
	echo "📋 Current version: $$GIT_VERSION"; \
	echo "📝 Testing app_version update in Ansible defaults..."; \
	echo "📄 Original defaults file:"; \
	grep "^app_version:" $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml; \
	sed -i.bak "s/^app_version:.*/app_version: \"$$GIT_VERSION\"/" $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml; \
	echo "📄 Updated defaults file:"; \
	grep "^app_version:" $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml; \
	echo "✅ Version update test successful!"; \
	echo "🔄 Restoring original defaults file..."; \
	mv $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml.bak $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml; \
	echo "📄 Restored defaults file:"; \
	grep "^app_version:" $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml

deploy: generate-docs
	@echo "🚀 Preparing deployment with git version..."
	@# Get current git tag or commit SHA
	@GIT_VERSION=$$(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD); \
	echo "📋 Current version: $$GIT_VERSION"; \
	echo "📝 Updating app_version in Ansible defaults..."; \
	sed -i.bak "s/^app_version:.*/app_version: \"$$GIT_VERSION\"/" $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml; \
	echo "✅ Updated app_version to: $$GIT_VERSION"; \
	echo "🚀 Deploying application with Ansible..."; \
	ansible-playbook -i $(ANSIBLE_INVENTORY) $(ANSIBLE_PLAYBOOK) --ask-vault-pass; \
	echo "🔄 Restoring original defaults file..."; \
	mv $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml.bak $(ANSIBLE_DIR)/roles/agent_auditor/defaults/main.yml

clean:
	@echo "Cleaning up build artifacts..."
	@rm -f $(BINARY_NAME) generate-keys documentation_embed.go
	@echo "Cleaning up generated documentation..."
	@rm -rf documentation