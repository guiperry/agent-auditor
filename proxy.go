package main

import (
	"fmt"
	"log"
	"net"
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
func NewProxyServer(targetHost string, targetPort int, proxyPort int) (*ProxyServer, error) {
	targetURL, err := url.Parse(fmt.Sprintf("http://%s:%d", targetHost, targetPort))
	if err != nil {
		return nil, err
	}

	// Create a custom director function to modify the request
	director := func(req *http.Request) {
		// Set the scheme and host to the target
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host

		// Preserve the original path and query
		// The path is already set in the request, so we don't need to modify it

		// Update the Host header to match the target
		req.Host = targetURL.Host

		// Log the modified request
		log.Printf("[PROXY] Director: %s %s -> %s://%s%s",
			req.Method, req.URL.Path, req.URL.Scheme, req.URL.Host, req.URL.Path)
	}

	// Create a custom reverse proxy with our director
	proxy := &httputil.ReverseProxy{
		Director: director,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[PROXY] Error: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("502 Bad Gateway - Proxy Error"))
		},
	}

	return &ProxyServer{
		targetURL: targetURL,
		proxy:     proxy,
		port:      proxyPort,
	}, nil
}

// Start begins listening and serving the proxy
func (p *ProxyServer) Start() error {
	// Create a custom handler that logs requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[PROXY] Received: %s %s", r.Method, r.URL.Path)

		// Add headers to help with debugging
		w.Header().Set("X-Proxied-By", "AEGONG-Proxy")

		// Serve the request through the proxy
		p.proxy.ServeHTTP(w, r)
	})

	// Start the server
	addr := fmt.Sprintf("0.0.0.0:%d", p.port)
	log.Printf("🔄 Proxy server starting on http://%s -> %s", addr, p.targetURL.String())
	return http.ListenAndServe(addr, handler)
}

// isPortInUse checks if a port is already in use
func isPortInUse(port int) bool {
	// Try to listen on the port to see if it's available
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)

	// If we can't listen, the port is in use
	if err != nil {
		return true
	}

	// Close the listener and return false (port is not in use)
	listener.Close()
	return false
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

	// Check if the port is already in use
	if isPortInUse(proxyPort) {
		log.Printf("Warning: Port %d is already in use. This could be another web server like Apache or Nginx.", proxyPort)
		log.Printf("Info: To use the built-in proxy, stop any other services using port %d", proxyPort)
		log.Printf("Info: You can also set PROXY_PORT environment variable to use a different port")
		return
	}

	// Get target host from environment or default to localhost
	targetHost := os.Getenv("TARGET_HOST")
	if targetHost == "" {
		targetHost = "localhost"
	}

	// Create and start the proxy in a goroutine
	proxy, err := NewProxyServer(targetHost, appPort, proxyPort)
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
