package glance

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (a *application) handleAudioProxyRequest(w http.ResponseWriter, r *http.Request) {
	streamURL := r.URL.Query().Get("url")
	if streamURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Only allow specific domains for security
	allowedDomains := []string{
		"streaming.radio.lublin.pl",
		"stream.open.fm",
		"stream-cdn-1.open.fm",
		"radioparty.pl",
		"s1.slotex.pl",
	}

	domainAllowed := false
	for _, domain := range allowedDomains {
		if strings.Contains(parsedURL.Host, domain) {
			domainAllowed = true
			break
		}
	}

	if !domainAllowed {
		http.Error(w, "Domain not allowed", http.StatusForbidden)
		return
	}

	// Create request to upstream
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	proxyReq, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		slog.Error("Failed to create proxy request", "error", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy relevant headers
	if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
		proxyReq.Header.Set("User-Agent", userAgent)
	}

	// Make request
	resp, err := client.Do(proxyReq)
	if err != nil {
		slog.Error("Failed to fetch stream", "error", err, "url", streamURL)
		http.Error(w, "Failed to fetch stream", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")

	// Copy content type
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Copy other relevant headers
	if icy := resp.Header.Get("icy-name"); icy != "" {
		w.Header().Set("icy-name", icy)
	}
	if icy := resp.Header.Get("icy-metaint"); icy != "" {
		w.Header().Set("icy-metaint", icy)
	}

	// Set cache control for streaming
	w.Header().Set("Cache-Control", "no-cache, no-store")

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Stream the content
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		slog.Error("Error streaming audio", "error", err)
	}
}
