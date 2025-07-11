package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

// ProxyServer represents a simple HTTP proxy server
type ProxyServer struct {
	targetURL *url.URL
	proxy     *httputil.ReverseProxy
	port      int
}

// NewProxyServer creates a new proxy server instance
func NewProxyServer(targetPort int, proxyPort int) (*ProxyServer, error) {
	targetURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", targetPort))
	if err != nil {
		return nil, err
	}

	return &ProxyServer{
		targetURL: targetURL,
		proxy:     httputil.NewSingleHostReverseProxy(targetURL),
		port:      proxyPort,
	}, nil
}

// Start begins listening and serving the proxy
func (p *ProxyServer) Start() error {
	// Create a custom handler that logs requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[PROXY] %s %s -> %s%s", r.Method, r.URL.Path, p.targetURL.String(), r.URL.Path)
		p.proxy.ServeHTTP(w, r)
	})

	// Start the server
	addr := fmt.Sprintf(":%d", p.port)
	log.Printf("🔄 Proxy server starting on http://localhost%s -> %s", addr, p.targetURL.String())
	return http.ListenAndServe(addr, handler)
}

// StartProxyIfNeeded starts the proxy server if needed
func StartProxyIfNeeded(appPort int) {
	// Check if we should start the proxy (only if we're not already running as root)
	if os.Getuid() != 0 {
		log.Printf("Info: Not running as root, proxy to port 80 will not be started")
		log.Printf("Info: To enable proxy, run with sudo or as root")
		return
	}

	// Get proxy port from environment or default to 80
	proxyPortStr := os.Getenv("PROXY_PORT")
	proxyPort := 80 // Default to standard HTTP port
	if proxyPortStr != "" {
		if port, err := strconv.Atoi(proxyPortStr); err == nil {
			proxyPort = port
		}
	}

	// Create and start the proxy in a goroutine
	proxy, err := NewProxyServer(appPort, proxyPort)
	if err != nil {
		log.Printf("Error creating proxy server: %v", err)
		return
	}

	go func() {
		if err := proxy.Start(); err != nil {
			log.Printf("Proxy server error: %v", err)
		}
	}()
}
