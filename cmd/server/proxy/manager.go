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
	pikoProxyURL    string
	pikoUpstreamURL string
	proxyPort       int
	upstreamPort    int
}

// NewManager creates a new proxy manager
func NewManager(proxyPort, upstreamPort int) *Manager {
	return &Manager{
		proxyPort:       proxyPort,
		upstreamPort:    upstreamPort,
		pikoProxyURL:    fmt.Sprintf("http://127.0.0.1:%d", proxyPort),
		pikoUpstreamURL: fmt.Sprintf("http://127.0.0.1:%d", upstreamPort),
	}
}

// ProxyRequest creates a handler that proxies requests to piko
func (m *Manager) ProxyRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: ProxyRequest hit. URL: %s", r.URL.Path)
		
		// Extract session ID from URL path (the first segment)
		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		
		if parts[0] == "" {
			http.Error(w, "Session ID is required", http.StatusBadRequest)
			return
		}

		sessionID := parts[0]
		
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
				log.Printf("DEBUG: Setting X-Piko-Endpoint header: %s", sessionID)

				// Copy other headers
				pr.Out.Header.Set("X-Forwarded-Host", r.Host)
				pr.Out.Header.Set("X-Forwarded-Proto", scheme(r))

				// Handle WebSocket upgrade
				if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
					pr.Out.Header.Set("Upgrade", "websocket")
					pr.Out.Header.Set("Connection", "Upgrade")
				}
			},
			ModifyResponse: func(resp *http.Response) error {
				// If Piko returns 502, it means the upstream (client) is not connected.
				// We map this to 404 to indicate "Session Not Found".
				if resp.StatusCode == http.StatusBadGateway {
					resp.StatusCode = http.StatusNotFound
				}
				return nil
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("Proxy error for session %s: %v", sessionID, err)
				// If we can't connect to Piko proxy (localhost), that's a 502.
				// If Piko returns 502 (handled in ModifyResponse), it's a 404.
				http.Error(w, "Proxy error", http.StatusBadGateway)
			},
		}

		// Flush the response after writing to support SSE/WebSocket
		proxy.FlushInterval = 100 * time.Millisecond

		// Serve the proxy
		proxy.ServeHTTP(w, r)
	}
}

// ProxyRootRequest creates a handler that proxies requests to piko as root-service
// This is used for "/" and "/piko" paths
func (m *Manager) ProxyRootRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create proxy director
		targetURL, _ := url.Parse(m.pikoProxyURL)
		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				// Set the target URL
				pr.Out.URL = targetURL
				pr.Out.URL.Path = r.URL.Path
				pr.Out.URL.RawQuery = r.URL.RawQuery

				// Set piko endpoint header to root-service
				pr.Out.Header.Set("X-Piko-Endpoint", "root-service")

				// Copy other headers
				pr.Out.Header.Set("X-Forwarded-Host", r.Host)
				pr.Out.Header.Set("X-Forwarded-Proto", scheme(r))

				// Handle WebSocket upgrade
				if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
					pr.Out.Header.Set("Upgrade", "websocket")
					pr.Out.Header.Set("Connection", "Upgrade")
				}
			},
			ModifyResponse: func(resp *http.Response) error {
				// If Piko returns 502, it means the upstream (client) is not connected.
				// We map this to 404 to indicate "Session Not Found".
				if resp.StatusCode == http.StatusBadGateway {
					resp.StatusCode = http.StatusNotFound
				}
				return nil
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("Proxy error for root service: %v", err)
				// If we can't connect to Piko proxy (localhost), that's a 502.
				// If Piko returns 502 (handled in ModifyResponse), it's a 404.
				http.Error(w, "Proxy error", http.StatusBadGateway)
			},
		}

		// Flush the response after writing to support SSE/WebSocket
		proxy.FlushInterval = 100 * time.Millisecond

		// Serve the proxy
		proxy.ServeHTTP(w, r)
	}
}

// ProxyUpstreamRequest creates a handler that proxies requests to piko upstream
// This is used for "/piko" paths when acting as an agent connection endpoint
func (m *Manager) ProxyUpstreamRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: ProxyUpstreamRequest hit. URL: %s", r.URL.Path)
		// Create proxy director
		targetURL, _ := url.Parse(m.pikoUpstreamURL)
		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				// Set the target URL
				pr.Out.URL = targetURL
				
				// Ensure path starts with /piko
				newPath := r.URL.Path
				if !strings.HasPrefix(newPath, "/piko") {
					// Prepend /piko if missing (for /v1/upstream routes)
					if strings.HasPrefix(newPath, "/") {
						newPath = "/piko" + newPath
					} else {
						newPath = "/piko/" + newPath
					}
				}
				
				pr.Out.URL.Path = newPath
				pr.Out.URL.RawQuery = r.URL.RawQuery

				// Copy other headers
				pr.Out.Header.Set("X-Forwarded-Host", r.Host)
				pr.Out.Header.Set("X-Forwarded-Proto", scheme(r))

				// Handle WebSocket upgrade
				if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
					pr.Out.Header.Set("Upgrade", "websocket")
					pr.Out.Header.Set("Connection", "Upgrade")
				}
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("Proxy error for upstream service: %v", err)
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

