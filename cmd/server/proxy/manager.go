package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Manager manages the proxy connections to piko
type Manager struct {
	pikoProxyURL string
	proxyPort    int
}

// NewManager creates a new proxy manager
func NewManager(proxyPort int) *Manager {
	return &Manager{
		proxyPort:    proxyPort,
		pikoProxyURL: fmt.Sprintf("http://127.0.0.1:%d", proxyPort),
	}
}

// ProxyRequest creates a handler that proxies requests to piko
func (m *Manager) ProxyRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract session ID from URL path
		// URL pattern: /:session/*path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 2 {
			http.Error(w, "Invalid session path", http.StatusBadRequest)
			return
		}

		sessionID := parts[1]
		if sessionID == "" {
			http.Error(w, "Session ID is required", http.StatusBadRequest)
			return
		}

		// Create proxy director
		targetURL, _ := url.Parse(m.pikoProxyURL)
		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				// Set the target URL
				pr.Out.URL = targetURL
				pr.Out.URL.Path = r.URL.Path
				pr.Out.URL.RawQuery = r.URL.RawQuery

				// Set piko endpoint header
				pr.Out.Header.Set("X-Piko-Endpoint", sessionID)

				// Copy other headers
				pr.Out.Header.Set("X-Forwarded-Host", r.Host)
				pr.Out.Header.Set("X-Forwarded-Proto", scheme(r))

				// Handle WebSocket upgrade
				if r.Header.Get("Upgrade") == "websocket" {
					pr.Out.Header.Set("Upgrade", "websocket")
					pr.Out.Header.Set("Connection", "upgrade")
				}
			},
			ModifyResponse: func(resp *http.Response) error {
				// Handle CORS if needed
				return nil
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("Proxy error: %v", err)
				http.Error(w, "Proxy error", http.StatusBadGateway)
			},
		}

		// Flush the response after writing to support SSE/WebSocket
		proxy.FlushInterval = 100 * time.Millisecond

		// Serve the proxy
		proxy.ServeHTTP(w, r)
	}
}


// scheme returns the scheme of the request (http or https)
func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}

