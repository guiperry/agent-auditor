# How to Contribute to Agent Auditor

Thank you for your interest in contributing to Agent Auditor! This guide will help you get started with contributing to our open source project.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Contribution Workflow](#contribution-workflow)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

We are committed to providing a friendly, safe, and welcoming environment for all contributors. Please read and follow our [Code of Conduct](https://github.com/guiperry/agent-auditor/blob/main/CODE_OF_CONDUCT.md).

## Getting Started

1. **Fork the repository**: Start by forking the [Agent Auditor repository](https://github.com/guiperry/agent-auditor) to your GitHub account.

2. **Clone your fork**: Clone your fork to your local machine:
   ```bash
   git clone https://github.com/YOUR-USERNAME/agent-auditor.git
   cd agent-auditor
   ```

3. **Add upstream remote**: Add the original repository as an upstream remote:
   ```bash
   git remote add upstream https://github.com/guiperry/agent-auditor.git
   ```

## Development Environment

### Prerequisites
- Go 1.21 or later
- Node.js 16 or later (for documentation generation)
- Docker (for running tests and simulations)

### Setup
1. Install Go dependencies:
   ```bash
   go mod download
   ```

2. Install Node.js dependencies (for documentation):
   ```bash
   npm install
   ```

3. Build the project:
   ```bash
   go build -o agent-auditor
   ```

## Contribution Workflow

1. **Create a branch**: Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes**: Make your changes to the codebase.

3. **Run tests**: Ensure all tests pass:
   ```bash
   go test ./...
   ```

4. **Commit changes**: Commit your changes with a descriptive commit message:
   ```bash
   git commit -m "Add feature: your feature description"
   ```

5. **Push changes**: Push your changes to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**: Open a pull request from your fork to the main repository.

## Pull Request Guidelines

- Ensure your PR addresses a specific issue or feature
- Include a clear description of the changes
- Update documentation as needed
- Add tests for new features
- Ensure all tests pass
- Follow the coding standards

## Coding Standards

We follow standard Go coding conventions:

- Use `gofmt` to format your code
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Document all exported functions, types, and constants
- Write clear, concise comments
- Use meaningful variable and function names

## Testing

- Write unit tests for all new functionality
- Ensure existing tests pass with your changes
- Consider adding integration tests for complex features
- Test edge cases and error conditions

## Documentation

- Update documentation for any changed functionality
- Document new features thoroughly
- Use clear, concise language
- Include examples where appropriate
- Run the documentation generator to verify changes:
  ```bash
  node scripts/doc_generator.js
  ```

## Community

- Join our [Discord server](https://discord.gg/agent-auditor) for discussions
- Participate in issue discussions on GitHub
- Help review pull requests from other contributors
- Share your experiences using Agent Auditor

Thank you for contributing to Agent Auditor! Your efforts help make this project better for everyone.

---

<div class="footer-links">
<a href="#/legal/CODE_OF_CONDUCT.md" class="footer-link">Contributor Covenant Code of Conduct</a> | <a href="#/legal/PRIVACY_POLICY.md" class="footer-link">PRIVACY_POLICY.md</a> | <a href="#/legal/TERMS_AND_CONDITIONS.md" class="footer-link">TERMS AND CONDITIONS</a>

© 2025 Agent Auditor
</div>
