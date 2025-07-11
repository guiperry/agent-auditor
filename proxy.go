package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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

// StartProxyIfNeeded is now a no-op since we're using NGINX as a reverse proxy
func StartProxyIfNeeded(appPort int) {
	log.Printf("Info: Built-in proxy is disabled. Using external NGINX as reverse proxy.")
	log.Printf("Info: Application is configured to listen on port %d", appPort)
	log.Printf("Info: NGINX should be configured to proxy requests to this port")
}
