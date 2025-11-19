package glance

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"time"
)

var beszelWidgetTemplate = mustParseTemplate("beszel.html", "widget-base.html")

type beszelWidget struct {
	widgetBase  `yaml:",inline"`
	URL         string         `yaml:"url"`
	Token       string         `yaml:"token"`
	RedirectURL string         `yaml:"redirect-url"`
	Systems     []beszelSystem `yaml:"-"`
}

type beszelResponse struct {
	Items []beszelSystem `json:"items"`
}

type beszelSystem struct {
	Name     string     `json:"name"`
	Host     string     `json:"host"`
	Status   string     `json:"status"`
	Info     beszelInfo `json:"info"`
	BootTime time.Time  `json:"-"`
}

type beszelInfo struct {
	Kernel   string  `json:"k"`
	Uptime   float64 `json:"u"`
	CPUModel string  `json:"m"`
	CPU      float64 `json:"cpu"`
	Memory   float64 `json:"mp"`
	Disk     float64 `json:"dp"`
	Load1    float64 `json:"l1"`
	Load5    float64 `json:"l5"`
	Load15   float64 `json:"l15"`
}

func (w *beszelWidget) initialize() error {
	w.withTitle("Beszel").withCacheDuration(10 * time.Second)
	if w.URL == "" {
		return errors.New("beszel widget: url is required")
	}
	return nil
}

func (w *beszelWidget) update(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, "GET", w.URL+"/api/collections/systems/records", nil)
	if err != nil {
		w.withError(err)
		return
	}

	if w.Token != "" {
		req.Header.Set("Authorization", "Bearer "+w.Token)
	}

	resp, err := decodeJsonFromRequest[*beszelResponse](defaultHTTPClient, req)
	if err != nil {
		w.withError(err)
		return
	}

	w.Systems = resp.Items
	now := time.Now()
	for i := range w.Systems {
		w.Systems[i].BootTime = now.Add(-time.Duration(w.Systems[i].Info.Uptime) * time.Second)
	}

	w.withError(nil).scheduleNextUpdate()
}

func (w *beszelWidget) Render() template.HTML {
	return w.renderTemplate(w, beszelWidgetTemplate)
}
