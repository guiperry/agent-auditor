# Go Embedding Scheme for Agent Auditor

## Overview

Go's `embed` package allows you to embed static files directly into your binary at compile time. This eliminates the need for external static file directories and makes deployment much simpler.

## Benefits

### 1. **Single Binary Deployment**
- No need to copy static files separately
- Reduces deployment complexity
- Eliminates "file not found" errors for static assets

### 2. **Improved Security**
- Static files can't be modified after compilation
- No risk of tampering with web assets
- Reduces attack surface

### 3. **Better Performance**
- Files are loaded from memory instead of disk
- No filesystem I/O for static assets
- Faster startup time

### 4. **Simplified Distribution**
- Single executable file contains everything
- No need to maintain directory structure
- Easier containerization

## Implementation Options

### Option 1: Embed Entire Directory
```go
//go:embed static/*
var staticFiles embed.FS

func setupRoutes() {
    staticFS, _ := fs.Sub(staticFiles, "static")
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
}
```

### Option 2: Embed Individual Files
```go
//go:embed static/index.html
var indexHTML []byte

//go:embed static/style.css
var styleCSS []byte

func serveIndex(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.Write(indexHTML)
}
```

### Option 3: Hybrid Approach (Recommended)
```go
//go:embed static/*
var staticFiles embed.FS

//go:embed static/index.html
var indexHTML []byte

// Use directory embedding for general static files
// Use individual embedding for frequently accessed files
```

## Current Project Structure Impact

### Before Embedding:
```
Agent_Auditor/
├── aegong                 # Binary
├── static/              # Required directory
│   ├── index.html
│   ├── style.css
│   ├── script.js
│   └── voice-integration.js
├── uploads/             # Runtime directory
├── reports/             # Runtime directory
└── voice_reports/       # Runtime directory
```

### After Embedding:
```
Agent_Auditor/
├── aegong                # Binary (contains static files)
├── uploads/             # Runtime directory
├── reports/             # Runtime directory
└── voice_reports/       # Runtime directory
```

## Migration Steps

### 1. Create embedded_assets.go
```go
package main

import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed static/*
var staticFiles embed.FS

func GetStaticFileSystem() http.FileSystem {
    staticFS, err := fs.Sub(staticFiles, "static")
    if err != nil {
        panic(err)
    }
    return http.FS(staticFS)
}
```

### 2. Update main.go
```go
// Replace this:
r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

// With this:
r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(GetStaticFileSystem())))
```

### 3. Update homeHandler
```go
// Replace this:
func homeHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "static/index.html")
}

// With this:
func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.Write(GetIndexHTML())
}
```

## Advanced Features

### Content Compression
```go
import "compress/gzip"

func serveCompressedStatic(w http.ResponseWriter, r *http.Request, content []byte) {
    if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
        w.Header().Set("Content-Encoding", "gzip")
        w.Header().Set("Content-Type", "text/css")
        
        gz := gzip.NewWriter(w)
        defer gz.Close()
        gz.Write(content)
    } else {
        w.Write(content)
    }
}
```

### Template Embedding
```go
//go:embed templates/*
var templateFiles embed.FS

func loadTemplates() *template.Template {
    return template.Must(template.ParseFS(templateFiles, "templates/*.html"))
}
```

### Build-time Asset Processing
```go
//go:generate go run build-tools/minify.go
//go:embed static/dist/*
var minifiedAssets embed.FS
```

## Considerations

### What to Embed
- ✅ Static HTML, CSS, JavaScript files (static/*)
- ✅ Python scripts (voice_inference.py)
- ✅ Requirements file (requirements.txt)
- ✅ Images, fonts, icons
- ✅ Configuration templates
- ❌ User uploads (uploads/)
- ❌ Generated reports (reports/)
- ❌ Runtime data (voice_reports/)
- ❌ Encrypted key files (default.key) - security sensitive
- ❌ Configuration files (voice_config.json) - user customizable

### File Size Impact
- Embedding adds files to binary size
- Consider compression for large assets
- Monitor binary size growth

### Development Workflow
- Files are embedded at compile time
- Need to rebuild for static file changes
- Consider development mode flag for filesystem serving

## Deployment Benefits for Agent Auditor

### Current Ansible Deployment
```yaml
- name: Copy static files
  copy:
    src: static/
    dest: "{{ app_dir }}/static/"
    
- name: Set permissions on static files
  file:
    path: "{{ app_dir }}/static"
    mode: '0644'
    recurse: yes
```

### With Embedding
```yaml
- name: Deploy binary
  copy:
    src: aegong
    dest: "{{ app_dir }}/aegong"
    mode: '0755'
# No need for static file deployment!
```

## Files Still Requiring Deployment

Even with embedding, some files still need to be handled at deployment:

### 1. Configuration Files
- `voice_config.json` - User-customizable TTS settings
- Should be templated by Ansible with environment-specific values

### 2. Security Files
- `default.key` - Encrypted API keys
- Must be deployed securely, never embedded in binary
- Contains sensitive API keys for TTS providers

### 3. Runtime Directories
- `uploads/` - User file uploads
- `reports/` - Generated audit reports
- `voice_reports/` - Generated audio files

### 4. Python Dependencies
- While `requirements.txt` is embedded for reference
- Python packages still need to be installed on target system
- `pip install -r requirements.txt` still required

## Recommended Implementation

For Agent Auditor, the implemented hybrid approach:
1. ✅ Embed all static web assets (HTML, CSS, JS)
2. ✅ Embed Python scripts (voice_inference.py) - written to filesystem when needed
3. ✅ Embed requirements.txt for reference
4. ✅ Keep runtime directories on filesystem (uploads, reports, voice_reports)
5. ✅ Keep configuration files deployable (voice_config.json, default.key)
6. ✅ Use individual file embedding for critical assets (index.html)
7. ✅ Use directory embedding for other static assets
