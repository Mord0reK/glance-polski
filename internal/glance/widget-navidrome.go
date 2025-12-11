package glance

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

var navidromeWidgetTemplate = mustParseTemplate("navidrome.html", "widget-base.html")

type navidromeWidget struct {
	widgetBase `yaml:",inline"`
	URL        string `yaml:"url"`
	User       string `yaml:"user"`
	Token      string `yaml:"token"` // Password
	Salt       string `yaml:"salt"`  // Optional

	// Runtime data
	Playlists  []navidromePlaylist `yaml:"-"`
	AuthParams string              `yaml:"-"`
}

type navidromePlaylist struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SongCount int    `json:"songCount"`
	Duration  int    `json:"duration"`
	CoverArt  string `json:"coverArt"`
}

type subsonicResponse struct {
	SubsonicResponse struct {
		Status    string `json:"status"`
		Version   string `json:"version"`
		Playlists struct {
			Playlist []navidromePlaylist `json:"playlist"`
		} `json:"playlists"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	} `json:"subsonic-response"`
}

func (w *navidromeWidget) initialize() error {
	w.withTitle("Navidrome").withCacheDuration(time.Hour)
	if w.URL == "" {
		return fmt.Errorf("navidrome widget: url is required")
	}
	if w.User == "" {
		return fmt.Errorf("navidrome widget: user is required")
	}
	if w.Token == "" {
		return fmt.Errorf("navidrome widget: token (password) is required")
	}
	return nil
}

func (w *navidromeWidget) update(ctx context.Context) {
	// Generate auth params
	salt := w.Salt
	if salt == "" {
		salt = generateRandomString(6)
	}
	token := md5Hash(w.Token + salt)

	query := url.Values{}
	query.Set("u", w.User)
	query.Set("t", token)
	query.Set("s", salt)
	query.Set("v", "1.16.1")
	query.Set("c", "glance")
	query.Set("f", "json")

	w.AuthParams = query.Encode()

	// Fetch playlists
	apiURL := fmt.Sprintf("%s/rest/getPlaylists?%s", w.URL, w.AuthParams)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		w.canContinueUpdateAfterHandlingErr(fmt.Errorf("failed to create request: %w", err))
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.canContinueUpdateAfterHandlingErr(fmt.Errorf("failed to fetch playlists: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.canContinueUpdateAfterHandlingErr(fmt.Errorf("failed to fetch playlists: status %d", resp.StatusCode))
		return
	}

	var subResp subsonicResponse
	if err := json.NewDecoder(resp.Body).Decode(&subResp); err != nil {
		w.canContinueUpdateAfterHandlingErr(fmt.Errorf("failed to decode response: %w", err))
		return
	}

	if subResp.SubsonicResponse.Status == "failed" {
		w.canContinueUpdateAfterHandlingErr(fmt.Errorf("api error: %s", subResp.SubsonicResponse.Error.Message))
		return
	}

	w.Playlists = subResp.SubsonicResponse.Playlists.Playlist
	w.canContinueUpdateAfterHandlingErr(nil)
}

func (w *navidromeWidget) Render() template.HTML {
	return w.renderTemplate(w, navidromeWidgetTemplate)
}

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			b[i] = letters[0] // Fallback
		} else {
			b[i] = letters[num.Int64()]
		}
	}
	return string(b)
}
