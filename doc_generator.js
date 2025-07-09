#!/usr/bin/env node

/**
 * Documentation Generator for Agent Auditor
 * 
 * This script processes all documentation from the docs folder and transforms it into
 * an organized documentation system in a new documentation folder, with support for
 * both Docsify and Go embedding.
 * 
 * Usage:
 *   node doc_generator.js
 * 
 * The script will:
 * 1. Read all markdown files from the docs directory
 * 2. Process and organize them into a structured documentation system
 * 3. Generate a sidebar and navigation structure
 * 4. Create Docsify-compatible output in the documentation folder
 * 5. Create a Go-embedding compatible structure
 */

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const { execSync } = require('child_process');

// Configuration
const CONFIG = {
  sourceDir: path.join(__dirname, 'docs'),
  outputDir: path.join(__dirname, 'documentation'),
  docsifyDir: path.join(__dirname, 'documentation', 'docsify'),
  hashFile: path.join(__dirname, 'documentation', '.doc_hashes.json'),
  projectName: 'Agent Auditor',
  projectRepo: 'https://github.com/guiperry/agent-auditor',
  categories: {
    'guides': 'User Guides',
    'deployment': 'Deployment',
    'development': 'Development',
    'api': 'API Reference',
    'security': 'Security',
    'voice': 'Voice Integration',
    'contribute': 'How to Contribute',
    'legal': 'Legal Documents'
  },
  // Special files that should be processed differently
  specialFiles: {
    'CODE_OF_CONDUCT.md': 'legal',
    'PRIVACY_POLICY.md': 'legal',
    'TERMS_AND_CONDITIONS.md': 'legal'
  }
};

// Ensure directories exist
function ensureDirectoryExists(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
    console.log(`Created directory: ${dir}`);
  }
}

// Clean output directory
function cleanDirectory(dir) {
  if (fs.existsSync(dir)) {
    // Instead of deleting the entire directory, we'll clear its contents
    // This approach is more idempotent as it doesn't change directory permissions or attributes
    const files = fs.readdirSync(dir);
    for (const file of files) {
      const filePath = path.join(dir, file);
      if (fs.lstatSync(filePath).isDirectory()) {
        cleanDirectory(filePath); // Recursively clean subdirectories
        fs.rmdirSync(filePath);   // Remove empty directory
      } else {
        fs.unlinkSync(filePath);  // Remove file
      }
    }
    console.log(`Cleaned contents of directory: ${dir}`);
  } else {
    ensureDirectoryExists(dir);
    console.log(`Created directory: ${dir}`);
  }
}

// Calculate hash of a file's content
function calculateFileHash(filePath) {
  const content = fs.readFileSync(filePath, 'utf8');
  return crypto.createHash('md5').update(content).digest('hex');
}

// Calculate hash of a string
function calculateStringHash(content) {
  return crypto.createHash('md5').update(content).digest('hex');
}

// Load existing hashes or create empty hash object
function loadHashes() {
  if (fs.existsSync(CONFIG.hashFile)) {
    try {
      return JSON.parse(fs.readFileSync(CONFIG.hashFile, 'utf8'));
    } catch (error) {
      console.log(`Error reading hash file: ${error.message}`);
      return {};
    }
  }
  return {};
}

// Save hashes to file
function saveHashes(hashes) {
  ensureDirectoryExists(path.dirname(CONFIG.hashFile));
  fs.writeFileSync(CONFIG.hashFile, JSON.stringify(hashes, null, 2));
}

// Check if file has changed by comparing hashes
function hasFileChanged(filePath, hashes) {
  const currentHash = calculateFileHash(filePath);
  const previousHash = hashes[filePath];
  return currentHash !== previousHash;
}

// Check if content has changed by comparing with stored hash
function hasContentChanged(content, key, hashes) {
  const currentHash = calculateStringHash(content);
  const previousHash = hashes[key];
  return currentHash !== previousHash;
}

// Update hash for a file
function updateFileHash(filePath, hashes) {
  hashes[filePath] = calculateFileHash(filePath);
  return hashes;
}

// Update hash for content
function updateContentHash(content, key, hashes) {
  hashes[key] = calculateStringHash(content);
  return hashes;
}

// Generate a consistent footer with legal links
function generateLegalFooter(categories) {
  let footer = `\n\n---\n\n<div class="footer-links">\n`;
  
  // Add legal documents if they exist
  if (categories['legal'] && categories['legal'].length > 0) {
    categories['legal'].forEach(doc => {
      footer += `<a href="#/legal/${doc.filename}" class="footer-link">${doc.title}</a> | `;
    });
    // Remove the last separator
    footer = footer.slice(0, -3);
  }
  
  footer += `\n\n© ${new Date().getFullYear()} ${CONFIG.projectName}\n</div>\n`;
  
  return footer;
}

// Get all markdown files from a directory
function getMarkdownFiles(dir) {
  const files = fs.readdirSync(dir);
  return files.filter(file => file.endsWith('.md'));
}

// Parse BibTeX file to extract citations
function parseBibTeX(bibPath) {
  if (!fs.existsSync(bibPath)) {
    return {};
  }
  
  const content = fs.readFileSync(bibPath, 'utf8');
  const citations = {};
  
  // Simple regex-based BibTeX parser
  // This is a basic implementation - a full parser would be more robust
  const entryRegex = /@(\w+)\s*{\s*([^,]+),\s*([\s\S]*?)\s*}\s*(?=@|\s*$)/g;
  const fieldRegex = /\s*(\w+)\s*=\s*{([^}]*)}/g;
  
  let match;
  while ((match = entryRegex.exec(content)) !== null) {
    const type = match[1];
    const key = match[2];
    const fieldsText = match[3];
    
    const fields = {};
    let fieldMatch;
    while ((fieldMatch = fieldRegex.exec(fieldsText)) !== null) {
      fields[fieldMatch[1].toLowerCase()] = fieldMatch[2];
    }
    
    citations[key] = {
      type,
      ...fields
    };
  }
  
  return citations;
}

// Format a citation in APA-like style
function formatCitation(citation) {
  if (!citation) {
    return '[Citation not found]';
  }
  
  try {
    const authors = citation.author ? citation.author.split(' and ').map(author => {
      const parts = author.split(',');
      if (parts.length > 1) {
        return `${parts[1].trim()} ${parts[0].trim()}`;
      }
      return author.trim();
    }).join(', ') : '';
    
    const year = citation.year || '';
    const title = citation.title || '';
    const journal = citation.journal || '';
    const volume = citation.volume || '';
    const number = citation.number ? `(${citation.number})` : '';
    const pages = citation.pages ? `pp. ${citation.pages}` : '';
    const publisher = citation.publisher || '';
    const url = citation.url || '';
    const link = citation.link || '';
    
    let formattedCitation = '';
    
    if (citation.type.toLowerCase() === 'article') {
      formattedCitation = `${authors} (${year}). ${title}. <em>${journal}</em>, ${volume}${number}, ${pages}.`;
    } else if (citation.type.toLowerCase() === 'book') {
      formattedCitation = `${authors} (${year}). <em>${title}</em>. ${publisher}.`;
    } else if (citation.type.toLowerCase() === 'inproceedings' || citation.type.toLowerCase() === 'conference') {
      formattedCitation = `${authors} (${year}). ${title}. In <em>${citation.booktitle || ''}</em>, ${pages}.`;
    } else {
      formattedCitation = `${authors} (${year}). ${title}.`;
    }
    
    // Add URL if available
    if (url) {
      formattedCitation += ` Retrieved from <a href="${url}" target="_blank">${url}</a>`;
    }
    
    // Add link to the paper/report if available
    if (link) {
      formattedCitation += ` <a href="${link}" target="_blank" class="citation-link">[View Paper]</a>`;
    }
    
    return formattedCitation;
  } catch (error) {
    console.error(`Error formatting citation: ${error.message}`);
    return '[Citation format error]';
  }
}

// Process citations in markdown content
function processCitations(content, citations) {
  if (!citations || Object.keys(citations).length === 0) {
    return content;
  }
  
  // Replace citation markers with formatted citations
  return content.replace(/\[cite:(\w+)\]/g, (match, citationKey) => {
    const citation = citations[citationKey];
    if (citation) {
      return formatCitation(citation);
    } else {
      console.warn(`Citation not found: ${citationKey}`);
      return `[Citation not found: ${citationKey}]`;
    }
  });
}

// Parse markdown file to extract metadata and content
function parseMarkdownFile(filePath, citations) {
  const content = fs.readFileSync(filePath, 'utf8');
  
  // Process citations in the content if citations are provided
  const processedContent = citations ? processCitations(content, citations) : content;
  
  const lines = processedContent.split('\n');
  
  // Extract title from first heading
  let title = '';
  const titleMatch = processedContent.match(/^# (.*)/m);
  if (titleMatch) {
    title = titleMatch[1];
  } else {
    // Use filename as fallback
    title = path.basename(filePath, '.md')
      .replace(/_/g, ' ')
      .replace(/\b\w/g, l => l.toUpperCase());
  }
  
  // Determine category based on content or filename
  let category = 'guides'; // Default category
  
  // Check for special files first
  const filename = path.basename(filePath);
  if (CONFIG.specialFiles[filename]) {
    category = CONFIG.specialFiles[filename];
  }
  // Check for other specific files
  else if (filename === 'KEY_MANAGEMENT.md') {
    category = 'security';
  } else if (filename === 'CONTRIBUTE.md') {
    category = 'contribute';
  } else if (filePath.toLowerCase().includes('contribute') || processedContent.toLowerCase().includes('how to contribute')) {
    category = 'contribute';
  } else if (filePath.toLowerCase().includes('deploy') || processedContent.toLowerCase().includes('deployment')) {
    category = 'deployment';
  } else if (filePath.toLowerCase().includes('voice') || processedContent.toLowerCase().includes('tts')) {
    category = 'voice';
  } else if (filePath.toLowerCase().includes('key') || processedContent.toLowerCase().includes('security')) {
    category = 'security';
  } else if (filePath.toLowerCase().includes('embed') || processedContent.toLowerCase().includes('development')) {
    category = 'development';
  } else if (filePath.toLowerCase().includes('api') || processedContent.toLowerCase().includes('api reference')) {
    category = 'api';
  }
  
  // Extract description from first paragraph after title
  let description = '';
  const paragraphs = processedContent.split('\n\n');
  for (let i = 0; i < paragraphs.length; i++) {
    if (!paragraphs[i].startsWith('#') && paragraphs[i].trim() !== '') {
      description = paragraphs[i].replace(/\n/g, ' ').trim();
      if (description.length > 160) {
        description = description.substring(0, 157) + '...';
      }
      break;
    }
  }
  
  return {
    title,
    category,
    description,
    content: processedContent,
    filename: path.basename(filePath),
    originalPath: filePath
  };
}

// Process all markdown files
function processMarkdownFiles() {
  const files = getMarkdownFiles(CONFIG.sourceDir);
  const docs = [];
  
  // Parse BibTeX file if it exists
  const bibPath = path.join(CONFIG.sourceDir, 'references.bib');
  const citations = parseBibTeX(bibPath);
  
  if (Object.keys(citations).length > 0) {
    console.log(`Loaded ${Object.keys(citations).length} citations from references.bib`);
  }
  
  files.forEach(file => {
    const filePath = path.join(CONFIG.sourceDir, file);
    const doc = parseMarkdownFile(filePath, citations);
    docs.push(doc);
    console.log(`Processed: ${file} (Category: ${doc.category})`);
  });
  
  return docs;
}

// Generate sidebar content for Docsify
function generateSidebar(docs) {
  let sidebar = `<a href="/" style="display:inline-block;margin-bottom:0.5rem;"><img src="/static/assets/icon.png" alt="Aegong Icon" style="width:5rem;vertical-align:middle;"></a>\n\n# ${CONFIG.projectName}\n\n`;
  
  // Group docs by category
  const categories = {};
  docs.forEach(doc => {
    if (!categories[doc.category]) {
      categories[doc.category] = [];
    }
    categories[doc.category].push(doc);
  });
  
  // Generate sidebar content for main categories (excluding legal)
  Object.keys(CONFIG.categories).forEach(category => {
    // Skip the legal category - we'll add it at the bottom
    if (category === 'legal') return;
    
    if (categories[category] && categories[category].length > 0) {
      sidebar += `## ${CONFIG.categories[category]}\n\n`;
      
      categories[category].forEach(doc => {
        const link = `${category}/${doc.filename}`;
        sidebar += `* [${doc.title}](${link})\n`;
      });
      
      sidebar += '\n';
    }
  });
  
  // Add footer with legal links
  sidebar += `<div class="sidebar-footer">\n\n---\n\n`;
  
  // Add legal documents if they exist
  if (categories['legal'] && categories['legal'].length > 0) {
    categories['legal'].forEach(doc => {
      const link = `legal/${doc.filename}`;
      sidebar += `* [${doc.title}](${link})\n`;
    });
  }
  
  sidebar += `\n© ${new Date().getFullYear()} ${CONFIG.projectName}\n</div>\n`;
  
  return sidebar;
}

// Generate index content for Docsify
function generateIndex(docs) {
  let index = `# ${CONFIG.projectName} Documentation\n\n`;
  
  // Add description
  index += `Welcome to the ${CONFIG.projectName} documentation. This guide provides comprehensive information about using, deploying, and developing with Agent Auditor.\n\n`;
  
  // Add developer community message
  index += `We're building an open source developer community around Agent Auditor. If you're interested in contributing, please check out our [How to Contribute](contribute/CONTRIBUTE.md) guide.\n\n`;
  
  // Add quick links
  index += `## Quick Links\n\n`;
  
  // Group docs by category
  const categories = {};
  docs.forEach(doc => {
    if (!categories[doc.category]) {
      categories[doc.category] = [];
    }
    categories[doc.category].push(doc);
  });
  
  // Generate quick links by category (excluding legal)
  Object.keys(CONFIG.categories).forEach(category => {
    // Skip the legal category - we'll add it at the bottom
    if (category === 'legal') return;
    
    if (categories[category] && categories[category].length > 0) {
      index += `### ${CONFIG.categories[category]}\n\n`;
      
      categories[category].forEach(doc => {
        index += `* [${doc.title}](${category}/${doc.filename}) - ${doc.description}\n`;
      });
      
      index += '\n';
    }
  });
  
  // Add footer with legal links
  index += generateLegalFooter(categories);
  
  return index;
}

// Generate Docsify configuration
function generateDocsifyConfig() {
  return `
<!DOCTYPE html>
<html>
<head>
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <meta charset="UTF-8">
  <title>${CONFIG.projectName} Documentation</title>
  
  <!-- Favicon -->
  <link rel="apple-touch-icon" sizes="180x180" href="/static/assets/favicon/apple-touch-icon.png">
  <link rel="icon" type="image/png" sizes="32x32" href="/static/assets/favicon/favicon-32x32.png">
  <link rel="icon" type="image/png" sizes="16x16" href="/static/assets/favicon/favicon-16x16.png">
  <link rel="manifest" href="/static/assets/favicon/site.webmanifest">
  <link rel="shortcut icon" href="/static/assets/favicon/favicon.ico">
  <meta name="theme-color" content="#6a0dad">
  <meta name="msapplication-TileColor" content="#6a0dad">
  <meta name="msapplication-config" content="/static/assets/favicon/browserconfig.xml">
  
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/docsify@4/themes/dark.css">
  <style>
    :root {
      --theme-color: #4a9eff;
      --sidebar-nav-link-color--active: #6fb5ff;
      --sidebar-nav-link-border-color--active: #6fb5ff;
    }
    body {
      background-color: #0a0a0a;
      color: #e0e0e0;
    }
      .sidebar:before {
      content: '';
      display: block;
      width: 2rem;
      height: 2rem;
      margin: 1rem auto 1rem 0.5rem;
    
    }
    .sidebar {
      background-color: #1a1a2e;
      border-right: 1px solid #333;
      padding-left: 5px;
    }
    .sidebar-toggle {
      background-color: rgba(26, 26, 46, 0.8);
    }
    .search input {
      background-color: #16213e;
      color: #e0e0e0;
      border: 1px solid #333;
    }
    .markdown-section code {
      background-color: #16213e;
    }
    .markdown-section pre {
      background-color: #16213e;
    }
    .markdown-section pre > code {
      background-color: #16213e;
    }
    .markdown-section blockquote {
      border-left: 4px solid #4a9eff;
      background-color: #16213e;
    }
    .markdown-section a {
      color: #4a9eff;
    }
    .markdown-section a:hover {
      color: #6fb5ff;
      text-decoration: underline;
    }
    .citation-link {
      display: inline-block;
      margin-left: 8px;
      padding: 2px 8px;
      background-color: #4a9eff;
      color: #fff;
      border-radius: 4px;
      font-size: 0.8em;
      text-decoration: none;
      transition: background-color 0.3s ease;
    }
    .citation-link:hover {
      background-color: #6fb5ff;
      color: #fff;
      text-decoration: none;
    }
    .sidebar-footer {
      margin-top: 2rem;
      padding-top: 1rem;
      border-top: 1px solid #333;
      font-size: 0.85rem;
      opacity: 0.8;
    }
    .footer-links {
      margin-top: 2rem;
      padding-top: 1rem;
      border-top: 1px solid #333;
      font-size: 0.85rem;
      opacity: 0.8;
      text-align: center;
    }
    .footer-link {
      color: #4a9eff;
      margin: 0 0.5rem;
      text-decoration: none;
    }
    .footer-link:hover {
      text-decoration: underline;
    }
  </style>
</head>
<body>
  <div id="app"></div>
  <script>
    window.$docsify = {
      name: '${CONFIG.projectName}',
      repo: '${CONFIG.projectRepo}',
      loadSidebar: true,
      subMaxLevel: 3,
      auto2top: true,
      themeColor: '#4a9eff',
      search: {
        maxAge: 86400000,
        paths: 'auto',
        placeholder: 'Search',
        noData: 'No results found',
        depth: 6
      }
    }
  </script>
  <script src="https://cdn.jsdelivr.net/npm/docsify@4"></script>
  <script src="https://cdn.jsdelivr.net/npm/docsify@4/lib/plugins/search.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/prismjs@1/components/prism-bash.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/prismjs@1/components/prism-go.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/prismjs@1/components/prism-yaml.min.js"></script>
</body>
</html>
  `;
}



// Write file if content has changed
function writeFileIfChanged(filePath, content, hashes) {
  const fileKey = `content:${filePath}`;
  
  if (!fs.existsSync(filePath) || hasContentChanged(content, fileKey, hashes)) {
    ensureDirectoryExists(path.dirname(filePath));
    fs.writeFileSync(filePath, content);
    console.log(`Updated: ${path.relative(__dirname, filePath)}`);
    updateContentHash(content, fileKey, hashes);
    return true;
  }
  
  return false;
}

// Create Docsify structure
function createDocsifyStructure(docs, hashes) {
  ensureDirectoryExists(CONFIG.docsifyDir);
  let filesChanged = 0;
  
  // Create index.html
  const indexHtmlPath = path.join(CONFIG.docsifyDir, 'index.html');
  const indexHtmlContent = generateDocsifyConfig();
  if (writeFileIfChanged(indexHtmlPath, indexHtmlContent, hashes)) {
    filesChanged++;
  }
  
  // Create README.md (homepage)
  const readmePath = path.join(CONFIG.docsifyDir, 'README.md');
  const readmeContent = generateIndex(docs);
  if (writeFileIfChanged(readmePath, readmeContent, hashes)) {
    filesChanged++;
  }
  
  // Create _sidebar.md
  const sidebarPath = path.join(CONFIG.docsifyDir, '_sidebar.md');
  const sidebarContent = generateSidebar(docs);
  if (writeFileIfChanged(sidebarPath, sidebarContent, hashes)) {
    filesChanged++;
  }
  
  // Create category directories and copy files
  const categories = {};
  docs.forEach(doc => {
    if (!categories[doc.category]) {
      categories[doc.category] = [];
      ensureDirectoryExists(path.join(CONFIG.docsifyDir, doc.category));
    }
    categories[doc.category].push(doc);
  });
  
  // Generate the legal footer once
  const legalFooter = generateLegalFooter(categories);
  
  // Process each document and add the footer if it's not a legal document
  docs.forEach(doc => {
    let content = doc.content;
    
    // Add the legal footer to all non-legal documents
    if (doc.category !== 'legal') {
      content += legalFooter;
    }
    
    // Write file to category directory
    const docPath = path.join(CONFIG.docsifyDir, doc.category, doc.filename);
    if (writeFileIfChanged(docPath, content, hashes)) {
      filesChanged++;
    }
  });
  
  console.log(`Docsify documentation structure: ${filesChanged} files updated`);
}



// Process references.bib if it exists
function processReferences(hashes) {
  const bibPath = path.join(CONFIG.sourceDir, 'references.bib');
  if (fs.existsSync(bibPath)) {
    const bibContent = fs.readFileSync(bibPath, 'utf8');
    
    // Copy to output directory if changed
    const docsifyBibPath = path.join(CONFIG.docsifyDir, 'references.bib');
    
    let filesChanged = 0;
    
    if (writeFileIfChanged(docsifyBibPath, bibContent, hashes)) {
      filesChanged++;
    }
    
    // Generate a references page with all citations
    const citations = parseBibTeX(bibPath);
    if (Object.keys(citations).length > 0) {
      let referencesContent = '# References\n\n';
      referencesContent += 'This page lists all references used in the documentation.\n\n';
      
      // Sort citations by key
      const sortedKeys = Object.keys(citations).sort();
      
      sortedKeys.forEach(key => {
        const citation = citations[key];
        referencesContent += `## ${key}\n\n`;
        referencesContent += `${formatCitation(citation)}\n\n`;
        referencesContent += `**Citation Key**: \`[cite:${key}]\`\n\n`;
        referencesContent += '---\n\n';
      });
      
      // Write references page to output directory
      const docsifyReferencesPath = path.join(CONFIG.docsifyDir, 'references.md');
      
      if (writeFileIfChanged(docsifyReferencesPath, referencesContent, hashes)) {
        filesChanged++;
      }
      
      console.log(`Generated references page with ${Object.keys(citations).length} citations`);
    }
    
    if (filesChanged > 0) {
      console.log(`References: ${filesChanged} files updated`);
    } else {
      console.log('References: No changes detected');
    }
  }
}

// Main function
function main() {
  console.log('Starting documentation generation...');
  
  // Load existing hashes
  let hashes = loadHashes();
  
  // Ensure output directory exists
  ensureDirectoryExists(CONFIG.outputDir);
  
  // Check if source files have changed
  const sourceFiles = getMarkdownFiles(CONFIG.sourceDir);
  let sourceChanged = false;
  
  // Check if any source files have changed
  for (const file of sourceFiles) {
    const filePath = path.join(CONFIG.sourceDir, file);
    if (hasFileChanged(filePath, hashes)) {
      sourceChanged = true;
      updateFileHash(filePath, hashes);
      console.log(`Source file changed: ${file}`);
    }
  }
  
  // Check if references.bib has changed
  const bibPath = path.join(CONFIG.sourceDir, 'references.bib');
  if (fs.existsSync(bibPath) && hasFileChanged(bibPath, hashes)) {
    sourceChanged = true;
    updateFileHash(bibPath, hashes);
    console.log('References file changed: references.bib');
  }
  
  // Process markdown files
  const docs = processMarkdownFiles();
  
  // Create Docsify structure
  createDocsifyStructure(docs, hashes);
  
  // Process references
  processReferences(hashes);
  
  // Save updated hashes
  saveHashes(hashes);
  
  console.log('Documentation generation complete!');
  console.log(`Documentation: ${CONFIG.docsifyDir}`);
}

// Run the main function
main();