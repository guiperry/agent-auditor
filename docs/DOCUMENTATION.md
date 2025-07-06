# Documentation System

## Overview

The Agent Auditor documentation system automatically processes all markdown files in the `docs` directory and transforms them into an organized documentation system in the `documentation` folder. This system supports both Docsify for web-based viewing and Go embedding for including documentation in the compiled binary.

## Documentation Structure

The documentation is organized into the following categories:

1. **User Guides** - General usage instructions and guides
2. **Deployment** - Deployment and configuration guides
3. **Development** - Development guides and API documentation
4. **API Reference** - API specifications and references
5. **Security** - Security-related documentation
6. **Voice Integration** - Voice and TTS integration guides

## Accessing Documentation

### Docsify Web Interface

The Docsify documentation can be accessed by opening `documentation/docsify/index.html` in a web browser. This provides a fully interactive documentation experience with:

- Sidebar navigation
- Full-text search
- Syntax highlighting for code blocks
- Mobile-responsive design

### Go Embedded Documentation

The documentation is also embedded in the compiled binary, making it accessible directly from the application. This ensures that documentation is always available with the application, even in offline environments.

## Documentation Generation

Documentation is automatically generated during the build process. You can also manually generate the documentation using:

```bash
make generate-docs
```

This will:

1. Process all markdown files in the `docs` directory
2. Organize them into categories based on content analysis
3. Generate a navigation structure
4. Create both Docsify and Go-embeddable versions

## Adding New Documentation

To add new documentation:

1. Create a new markdown file in the `docs` directory
2. Use a clear, descriptive filename (e.g., `API_REFERENCE.md`)
3. Start the file with a top-level heading (e.g., `# API Reference`)
4. Include a brief description in the first paragraph
5. Run `make generate-docs` to update the documentation system

The documentation generator will automatically:
- Determine the appropriate category based on content
- Add the document to the navigation structure
- Include it in both Docsify and Go-embedded versions

## Documentation Best Practices

1. **Use Descriptive Titles**: Start each document with a clear `# Title`
2. **Include a Description**: Add a brief description in the first paragraph
3. **Use Proper Heading Hierarchy**: Start with `#`, then `##`, `###`, etc.
4. **Include Code Examples**: Use fenced code blocks with language specifiers
5. **Add Cross-References**: Link to other documentation pages where relevant
6. **Keep Content Focused**: Each document should cover a specific topic
7. **Update Regularly**: Keep documentation in sync with code changes
8. **Use Citations**: Reference academic papers and other sources using the citation syntax

## Using Citations

The documentation system supports citations using BibTeX. To add citations to your documentation:

1. Add your references to the `docs/references.bib` file in BibTeX format:

```bibtex
@article{example2023paper,
  title={Example Paper Title},
  author={Example, Author and Another, Author},
  journal={Journal of Examples},
  year={2023},
  url={https://example.com/paper}
}
```

2. Reference the citation in your markdown files using the `[cite:key]` syntax:

```markdown
This is a statement that needs a citation [cite:example2023paper].
```

3. When the documentation is generated, the citation key will be replaced with a properly formatted citation.

4. A references page will be automatically generated with all citations listed.

## Technical Implementation

The documentation system consists of:

1. **doc_generator.js**: JavaScript script that processes markdown files
2. **documentation/docsify/**: Docsify-compatible documentation
3. **documentation/go-embed/**: Go-embeddable documentation
4. **documentation_embed.go**: Go helper for embedding documentation

The system is integrated with the build process through the Makefile, ensuring documentation is always up-to-date with each build.