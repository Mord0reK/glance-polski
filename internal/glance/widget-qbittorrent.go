package glance

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

var qbittorrentWidgetTemplate = mustParseTemplate("qbittorrent.html", "widget-base.html")

type qbittorrentWidget struct {
	widgetBase     `yaml:",inline"`
	URL            string `yaml:"url"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	HideSeeding    bool   `yaml:"hide-seeding"`
	HideCompleted  bool   `yaml:"hide-completed"`
	ShowOnlyActive bool   `yaml:"show-only-active"`
	Limit          int    `yaml:"limit"`
	SortBy         string `yaml:"sort-by"` // name, progress, speed, eta

	// Internal state
	client            *http.Client
	clientMutex       sync.Mutex
	loginAttempts     int
	loginBlockedUntil time.Time

	// Data
	Summary qbittorrentSummary `yaml:"-"`
}

type qbittorrentTorrent struct {
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	Progress      float64 `json:"progress"`
	State         string  `json:"state"`
	Size          int64   `json:"size"`
	Downloaded    int64   `json:"downloaded"`
	DownloadSpeed float64 `json:"dlspeed"`
	UploadSpeed   float64 `json:"upspeed"`
	ETA           int64   `json:"eta"`
	NumSeeds      int     `json:"num_seeds"`
	NumLeeches    int     `json:"num_leechs"`

	// Computed fields for template
	ProgressPercent      int
	StateIcon            string
	StateText            string
	SizeFormatted        string
	DownloadedFormatted  string
	SpeedFormatted       string
	UploadSpeedFormatted string
	ETAFormatted         string
}

type qbittorrentSummary struct {
	TotalDownloadSpeed          float64
	TotalUploadSpeed            float64
	TotalDownloadSpeedFormatted string
	TotalUploadSpeedFormatted   string
	SeedingCount                int
	DownloadingCount            int
	PausedCount                 int
	TotalCount                  int
	Torrents                    []qbittorrentTorrent
}

func (widget *qbittorrentWidget) initialize() error {
	widget.withTitle("qBittorrent").withCacheDuration(1 * time.Minute)

	if widget.URL == "" {
		return fmt.Errorf("url is required")
	}
	widget.URL = strings.TrimSuffix(widget.URL, "/")

	if widget.Username == "" {
		return fmt.Errorf("username is required")
	}

	if widget.Password == "" {
		return fmt.Errorf("password is required")
	}

	if widget.Limit <= 0 {
		widget.Limit = 10
	}

	if widget.SortBy == "" {
		widget.SortBy = "progress"
	}

	return nil
}

func (widget *qbittorrentWidget) update(ctx context.Context) {
	summary, err := widget.fetchTorrents()

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Summary = summary
}

func (widget *qbittorrentWidget) Render() template.HTML {
	return widget.renderTemplate(widget, qbittorrentWidgetTemplate)
}

// Login to qBittorrent API
func (widget *qbittorrentWidget) login() error {
	widget.clientMutex.Lock()
	defer widget.clientMutex.Unlock()

	// Check if login is blocked due to repeated failures
	if time.Now().Before(widget.loginBlockedUntil) {
		remaining := time.Until(widget.loginBlockedUntil)
		return fmt.Errorf("login blocked for %v due to repeated failures", remaining.Round(time.Minute))
	}

	jar, _ := cookiejar.New(nil)
	widget.client = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	form := url.Values{}
	form.Set("username", widget.Username)
	form.Set("password", widget.Password)

	loginURL := widget.URL + "/api/v2/auth/login"

	slog.Debug("qBittorrent login attempt", "url", loginURL, "username", widget.Username)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", widget.URL)

	resp, err := widget.client.Do(req)
	if err != nil {
		widget.loginAttempts++
		if widget.loginAttempts >= 3 {
			widget.loginBlockedUntil = time.Now().Add(30 * time.Minute)
			slog.Warn("qBittorrent login failed 3 times, blocking for 30 minutes")
			widget.loginAttempts = 0
			return fmt.Errorf("login blocked for 30 minutes due to repeated failures")
		}
		return fmt.Errorf("login failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if strings.TrimSpace(string(body)) != "Ok." {
		widget.loginAttempts++
		if widget.loginAttempts >= 3 {
			widget.loginBlockedUntil = time.Now().Add(30 * time.Minute)
			slog.Warn("qBittorrent login failed 3 times, blocking for 30 minutes")
			widget.loginAttempts = 0
			return fmt.Errorf("login blocked for 30 minutes due to repeated failures")
		}
		return fmt.Errorf("login failed, response: %s", string(body))
	}

	widget.loginAttempts = 0
	slog.Debug("qBittorrent login successful")
	return nil
}

func (widget *qbittorrentWidget) fetchTorrents() (qbittorrentSummary, error) {
	summary, err := widget.fetchTorrentsOnce()
	if err != nil && strings.Contains(err.Error(), "unauthorized") {
		slog.Debug("qBittorrent session expired, re-logging in...")
		if loginErr := widget.login(); loginErr != nil {
			return qbittorrentSummary{}, loginErr
		}
		summary, err = widget.fetchTorrentsOnce()
	}
	return summary, err
}

func (widget *qbittorrentWidget) fetchTorrentsOnce() (qbittorrentSummary, error) {
	if widget.client == nil {
		if err := widget.login(); err != nil {
			return qbittorrentSummary{}, err
		}
	}

	req, err := http.NewRequest("GET", widget.URL+"/api/v2/torrents/info", nil)
	if err != nil {
		return qbittorrentSummary{}, err
	}

	resp, err := widget.client.Do(req)
	if err != nil {
		return qbittorrentSummary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return qbittorrentSummary{}, fmt.Errorf("unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return qbittorrentSummary{}, fmt.Errorf("failed to fetch torrents, status: %s, body: %s", resp.Status, string(body))
	}

	var rawTorrents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawTorrents); err != nil {
		return qbittorrentSummary{}, fmt.Errorf("failed decoding torrents JSON: %w", err)
	}

	return widget.processTorrents(rawTorrents), nil
}

func (widget *qbittorrentWidget) processTorrents(rawTorrents []map[string]interface{}) qbittorrentSummary {
	summary := qbittorrentSummary{
		Torrents: make([]qbittorrentTorrent, 0),
	}

	for _, raw := range rawTorrents {
		torrent := qbittorrentTorrent{
			Name:          toString(raw["name"]),
			Category:      toString(raw["category"]),
			Progress:      toFloat64(raw["progress"]),
			State:         toString(raw["state"]),
			Size:          toInt64(raw["size"]),
			Downloaded:    toInt64(raw["downloaded"]),
			DownloadSpeed: toFloat64(raw["dlspeed"]),
			UploadSpeed:   toFloat64(raw["upspeed"]),
			ETA:           toInt64(raw["eta"]),
			NumSeeds:      toInt(raw["num_seeds"]),
			NumLeeches:    toInt(raw["num_leechs"]),
		}

		// Compute display fields
		torrent.ProgressPercent = int(torrent.Progress * 100)
		torrent.StateIcon = qbittorrentStateToIcon(torrent.State)
		torrent.StateText = qbittorrentStateToText(torrent.State)
		torrent.SizeFormatted = formatBytes(torrent.Size)
		torrent.DownloadedFormatted = formatBytes(torrent.Downloaded)

		// Only show download speed for downloading torrents
		if torrent.DownloadSpeed > 0 && torrent.Progress < 1 {
			torrent.SpeedFormatted = formatBytesPerSecond(torrent.DownloadSpeed)
		}

		// Only show upload speed if uploading
		if torrent.UploadSpeed > 0 {
			torrent.UploadSpeedFormatted = formatBytesPerSecond(torrent.UploadSpeed)
		}

		// Only show ETA for downloading torrents with valid ETA
		if torrent.ETA > 0 && torrent.ETA < 8640000 && torrent.Progress < 1 {
			torrent.ETAFormatted = formatETA(torrent.ETA)
		}

		// Update summary counts
		summary.TotalCount++
		summary.TotalDownloadSpeed += torrent.DownloadSpeed
		summary.TotalUploadSpeed += torrent.UploadSpeed

		switch torrent.State {
		case "uploading", "forcedUP", "stalledUP", "queuedUP", "checkingUP":
			summary.SeedingCount++
		case "downloading", "forcedDL", "stalledDL", "queuedDL", "checkingDL", "metaDL":
			summary.DownloadingCount++
		case "pausedDL", "pausedUP", "stoppedDL", "stoppedUP":
			summary.PausedCount++
		}

		// Apply filters
		if widget.HideSeeding && (torrent.State == "uploading" || torrent.State == "forcedUP" || torrent.State == "stalledUP" || torrent.State == "queuedUP") {
			continue
		}
		if widget.HideCompleted && torrent.Progress >= 1.0 {
			continue
		}
		if widget.ShowOnlyActive && torrent.DownloadSpeed == 0 && torrent.UploadSpeed == 0 {
			continue
		}

		summary.Torrents = append(summary.Torrents, torrent)
	}

	// Sort torrents
	widget.sortTorrents(summary.Torrents)

	// Apply limit
	if len(summary.Torrents) > widget.Limit {
		summary.Torrents = summary.Torrents[:widget.Limit]
	}

	// Format summary speeds (in MB/s as number only)
	summary.TotalDownloadSpeedFormatted = fmt.Sprintf("%.2f", summary.TotalDownloadSpeed/(1024*1024))
	summary.TotalUploadSpeedFormatted = fmt.Sprintf("%.2f", summary.TotalUploadSpeed/(1024*1024))

	return summary
}

func (widget *qbittorrentWidget) sortTorrents(torrents []qbittorrentTorrent) {
	switch widget.SortBy {
	case "name":
		sort.Slice(torrents, func(i, j int) bool {
			return strings.ToLower(torrents[i].Name) < strings.ToLower(torrents[j].Name)
		})
	case "progress":
		sort.Slice(torrents, func(i, j int) bool {
			// Downloading torrents first (by progress ascending), then seeding
			if torrents[i].Progress < 1 && torrents[j].Progress >= 1 {
				return true
			}
			if torrents[i].Progress >= 1 && torrents[j].Progress < 1 {
				return false
			}
			return torrents[i].Progress < torrents[j].Progress
		})
	case "speed":
		sort.Slice(torrents, func(i, j int) bool {
			return torrents[i].DownloadSpeed > torrents[j].DownloadSpeed
		})
	case "eta":
		sort.Slice(torrents, func(i, j int) bool {
			// Torrents with valid ETA first
			if torrents[i].ETA > 0 && torrents[j].ETA <= 0 {
				return true
			}
			if torrents[i].ETA <= 0 && torrents[j].ETA > 0 {
				return false
			}
			return torrents[i].ETA < torrents[j].ETA
		})
	}
}

func qbittorrentStateToIcon(state string) string {
	switch state {
	case "uploading", "forcedUP":
		return "seeding"
	case "downloading", "forcedDL":
		return "downloading"
	case "stalledUP", "stalledDL":
		return "stalled"
	case "pausedDL", "pausedUP", "stoppedDL", "stoppedUP":
		return "paused"
	case "queuedDL", "queuedUP":
		return "queued"
	case "checkingDL", "checkingUP", "checkingResumeData":
		return "checking"
	case "metaDL":
		return "metadata"
	case "error", "missingFiles":
		return "error"
	default:
		return "unknown"
	}
}

func qbittorrentStateToText(state string) string {
	switch state {
	case "uploading", "forcedUP":
		return "Seeding"
	case "downloading", "forcedDL":
		return "Downloading"
	case "stalledUP":
		return "Seeding (stalled)"
	case "stalledDL":
		return "Downloading (stalled)"
	case "pausedDL", "stoppedDL":
		return "Paused"
	case "pausedUP", "stoppedUP":
		return "Completed"
	case "queuedDL":
		return "Queued (DL)"
	case "queuedUP":
		return "Queued (UP)"
	case "checkingDL", "checkingUP", "checkingResumeData":
		return "Checking"
	case "metaDL":
		return "Fetching metadata"
	case "error":
		return "Error"
	case "missingFiles":
		return "Missing files"
	default:
		return state
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatBytesPerSecond(bytesPerSec float64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	}
	if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	}
	return fmt.Sprintf("%.1f MB/s", bytesPerSec/(1024*1024))
}

func formatETA(seconds int64) string {
	if seconds <= 0 || seconds == 8640000 { // 8640000 is qBittorrent's "infinite" value
		return "∞"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
	}
	return fmt.Sprintf("%dd %dh", seconds/86400, (seconds%86400)/3600)
}

// Helper functions for type conversion
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	default:
		return 0
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}
