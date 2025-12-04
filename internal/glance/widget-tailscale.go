package glance

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"
)

var tailscaleWidgetTemplate = mustParseTemplate("tailscale.html", "widget-base.html")

type tailscaleWidget struct {
	widgetBase          `yaml:",inline"`
	URL                 string `yaml:"url"`
	Token               string `yaml:"token"`
	Tailnet             string `yaml:"tailnet"`
	CollapseAfter       int    `yaml:"collapse-after"`
	ShowOnlineIndicator bool   `yaml:"show-online-indicator"`
	ShowExpiryDisabled  bool   `yaml:"show-expiry-disabled"`
	ShowDisconnected    bool   `yaml:"show-disconnected"`
	ShowBlocksIncoming  bool   `yaml:"show-blocks-incoming"`
	ShowJoinedDate      bool   `yaml:"show-joined-date"`
	Devices             []tailscaleDevice
	OnlineDevices       []tailscaleDevice
	OfflineDevices      []tailscaleDevice
}

type tailscaleDevice struct {
	ID              string
	Name            string
	ShortName       string
	OS              string
	User            string
	Addresses       []string
	PrimaryAddress  string
	LastSeen        time.Time
	LastSeenStr     string
	UpdateAvailable bool
	IsOnline        bool
	// Fields actually available from API
	KeyExpiryDisabled         bool
	BlocksIncomingConnections bool
	Expires                   time.Time
	ExpiresStr                string
	Created                   time.Time
	CreatedStr                string
	ConnectedToControl        bool
	// Feature flags
	IsExitNode                bool
	IsSubnetRouter            bool
	TailscaleSSHEnabled       bool
	AdvertisedRoutes          []string
}

type tailscaleAPIResponse struct {
	Devices []tailscaleAPIDevice `json:"devices"`
}

type tailscaleAPIDevice struct {
	ID                        string   `json:"id"`
	Name                      string   `json:"name"`
	Hostname                  string   `json:"hostname"`
	OS                        string   `json:"os"`
	User                      string   `json:"user"`
	Addresses                 []string `json:"addresses"`
	LastSeen                  string   `json:"lastSeen"`
	UpdateAvailable           bool     `json:"updateAvailable"`
	KeyExpiryDisabled         bool     `json:"keyExpiryDisabled"`
	Expires                   string   `json:"expires"`
	Created                   string   `json:"created"`
	BlocksIncomingConnections bool     `json:"blocksIncomingConnections"`
	ConnectedToControl        bool     `json:"connectedToControl"`
	ClientVersion             string   `json:"clientVersion"`
	IsExitNode                bool     `json:"isExitNode"`
	AdvertisedRoutes          []string `json:"advertisedRoutes"`
	EnabledRoutes             []string `json:"enabledRoutes"`
	TailscaleSSHEnabled       bool     `json:"tailscaleSSHEnabled"`
}

// Struktura odpowiedzi z /device/{id}/routes
type tailscaleRoutesResponse struct {
	AdvertisedRoutes []tailscaleRoute `json:"advertisedRoutes"`
	EnabledRoutes    []tailscaleRoute `json:"enabledRoutes"`
}

type tailscaleRoute struct {
	Route string `json:"route"`
}

// Struktura odpowiedzi z /device/{id}
type tailscaleDeviceDetailsResponse struct {
	TailscaleSSHEnabled bool `json:"tailscaleSSHEnabled"`
}

func (widget *tailscaleWidget) initialize() error {
	widget.withTitle("Tailscale").withCacheDuration(10 * time.Minute)

	if widget.Token == "" {
		return fmt.Errorf("token is required")
	}

	if widget.Tailnet == "" {
		widget.Tailnet = "-" // Default to current tailnet
	}

	if widget.URL == "" {
		widget.URL = fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/devices", widget.Tailnet)
	}

	if widget.CollapseAfter <= 0 {
		widget.CollapseAfter = 4
	}

	// Default badge visibility - all enabled by default
	// Users can disable specific badges in config
	// Note: these are set to true by default only if not explicitly configured

	return nil
}

func (widget *tailscaleWidget) update(ctx context.Context) {
	devices, err := widget.fetchDevices()

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Devices = devices

	// Grupowanie urządzeń na online i offline
	widget.OnlineDevices = make([]tailscaleDevice, 0)
	widget.OfflineDevices = make([]tailscaleDevice, 0)

	for _, device := range devices {
		if device.IsOnline {
			widget.OnlineDevices = append(widget.OnlineDevices, device)
		} else {
			widget.OfflineDevices = append(widget.OfflineDevices, device)
		}
	}
}

func (widget *tailscaleWidget) fetchDevices() ([]tailscaleDevice, error) {
	request, err := http.NewRequest("GET", widget.URL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)

	apiResponse, err := decodeJsonFromRequest[tailscaleAPIResponse](defaultHTTPClient, request)
	if err != nil {
		return nil, err
	}

	devices := make([]tailscaleDevice, len(apiResponse.Devices))
	now := time.Now()

	// Najpierw tworzymy podstawowe dane urządzeń
	for i, apiDevice := range apiResponse.Devices {
		device := tailscaleDevice{
			ID:                        apiDevice.ID,
			Name:                      apiDevice.Name,
			ShortName:                 extractShortName(apiDevice.Name),
			OS:                        apiDevice.OS,
			User:                      apiDevice.User,
			Addresses:                 apiDevice.Addresses,
			UpdateAvailable:           apiDevice.UpdateAvailable,
			KeyExpiryDisabled:         apiDevice.KeyExpiryDisabled,
			BlocksIncomingConnections: apiDevice.BlocksIncomingConnections,
			ConnectedToControl:        apiDevice.ConnectedToControl,
		}

		// Get primary address
		if len(apiDevice.Addresses) > 0 {
			device.PrimaryAddress = apiDevice.Addresses[0]
		}

		// Parse created time
		if apiDevice.Created != "" {
			created, err := time.Parse(time.RFC3339, apiDevice.Created)
			if err == nil {
				device.Created = created
				device.CreatedStr = created.Format("Jan 2006")
			}
		}

		// Parse last seen time
		if apiDevice.LastSeen != "" {
			lastSeen, err := time.Parse(time.RFC3339, apiDevice.LastSeen)
			if err == nil {
				device.LastSeen = lastSeen
				device.LastSeenStr = lastSeen.Format("Jan 2 3:04pm")

				// Device is considered online if last seen within 10 seconds
				device.IsOnline = lastSeen.After(now.Add(-10 * time.Second))
			}
		}

		// Parse expiry time
		if apiDevice.Expires != "" {
			expires, err := time.Parse(time.RFC3339, apiDevice.Expires)
			if err == nil {
				device.Expires = expires
				if !apiDevice.KeyExpiryDisabled {
					device.ExpiresStr = expires.Format("Jan 2 2006")
				}
			}
		}

		devices[i] = device
	}

	// Pobierz dodatkowe dane (routes i SSH) równolegle dla wszystkich urządzeń
	var wg sync.WaitGroup
	for i := range devices {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			widget.fetchDeviceDetails(&devices[idx])
		}(i)
	}
	wg.Wait()

	return devices, nil
}

// fetchDeviceDetails pobiera szczegóły urządzenia (routes i SSH)
func (widget *tailscaleWidget) fetchDeviceDetails(device *tailscaleDevice) {
	// Pobierz routes
	routesURL := fmt.Sprintf("https://api.tailscale.com/api/v2/device/%s/routes", device.ID)
	routesReq, err := http.NewRequest("GET", routesURL, nil)
	if err == nil {
		routesReq.Header.Set("Authorization", "Bearer "+widget.Token)
		routesResp, err := decodeJsonFromRequest[tailscaleRoutesResponse](defaultHTTPClient, routesReq)
		if err == nil {
			// Sprawdź czy jest Exit Node (route 0.0.0.0/0 lub ::/0)
			for _, route := range routesResp.EnabledRoutes {
				if route.Route == "0.0.0.0/0" || route.Route == "::/0" {
					device.IsExitNode = true
				} else {
					// Każdy inny route oznacza subnet router
					device.IsSubnetRouter = true
				}
				device.AdvertisedRoutes = append(device.AdvertisedRoutes, route.Route)
			}
		}
	}

	// Pobierz szczegóły urządzenia (dla SSH)
	detailsURL := fmt.Sprintf("https://api.tailscale.com/api/v2/device/%s", device.ID)
	detailsReq, err := http.NewRequest("GET", detailsURL, nil)
	if err == nil {
		detailsReq.Header.Set("Authorization", "Bearer "+widget.Token)
		detailsResp, err := decodeJsonFromRequest[tailscaleDeviceDetailsResponse](defaultHTTPClient, detailsReq)
		if err == nil {
			device.TailscaleSSHEnabled = detailsResp.TailscaleSSHEnabled
		}
	}
}

// extractShortName extracts the hostname before the first dot
func extractShortName(fullName string) string {
	if idx := strings.Index(fullName, "."); idx > 0 {
		return fullName[:idx]
	}
	return fullName
}

func (widget *tailscaleWidget) Render() template.HTML {
	return widget.renderTemplate(widget, tailscaleWidgetTemplate)
}
