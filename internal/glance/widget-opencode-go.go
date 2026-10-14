package glance

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var openCodeGoWidgetTemplate = mustParseTemplate("opencode-go.html", "widget-base.html")

const (
	defaultOpenCodeGoURL              = "https://opencode.ai/zen/go/v1/usage"
	defaultOpenCodeGoWarningThreshold = 85.0
)

type openCodeGoWidget struct {
	widgetBase       `yaml:",inline"`
	URL              string             `yaml:"url"`
	APIKey           string             `yaml:"api-key"`
	WarningThreshold float64            `yaml:"warning-threshold"`
	Windows          []openCodeGoWindow `yaml:"-"`
}

// openCodeGoWindow is one usage window (rolling / weekly / monthly) prepared
// for rendering by the template.
type openCodeGoWindow struct {
	Label    string
	Status   string
	Percent  float64
	ResetsAt time.Time
}

type openCodeGoResponse struct {
	Usage openCodeGoUsage `json:"usage"`
}

type openCodeGoUsage struct {
	Rolling openCodeGoUsageWindow `json:"rolling"`
	Weekly  openCodeGoUsageWindow `json:"weekly"`
	Monthly openCodeGoUsageWindow `json:"monthly"`
}

type openCodeGoUsageWindow struct {
	Status   string  `json:"status"`
	Percent  float64 `json:"percent"`
	ResetsAt string  `json:"resetsAt"`
}

func (widget *openCodeGoWidget) initialize() error {
	widget.withTitle("OpenCode Go").withCacheDuration(5 * time.Minute)

	if widget.URL == "" {
		widget.URL = defaultOpenCodeGoURL
	}

	if widget.APIKey == "" {
		return errors.New("api-key is required")
	}

	if widget.WarningThreshold == 0 {
		widget.WarningThreshold = defaultOpenCodeGoWarningThreshold
	}

	return nil
}

func (widget *openCodeGoWidget) update(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, widget.URL, nil)
	if err != nil {
		widget.withError(fmt.Errorf("creating request: %w", err))
		return
	}

	req.Header.Set("Authorization", "Bearer "+widget.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := decodeJsonFromRequest[*openCodeGoResponse](defaultHTTPClient, req)
	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Windows = []openCodeGoWindow{
		{Label: "5 godzin", Status: resp.Usage.Rolling.Status, Percent: resp.Usage.Rolling.Percent, ResetsAt: parseOpenCodeGoResetTime(resp.Usage.Rolling.ResetsAt)},
		{Label: "Tydzień", Status: resp.Usage.Weekly.Status, Percent: resp.Usage.Weekly.Percent, ResetsAt: parseOpenCodeGoResetTime(resp.Usage.Weekly.ResetsAt)},
		{Label: "Miesiąc", Status: resp.Usage.Monthly.Status, Percent: resp.Usage.Monthly.Percent, ResetsAt: parseOpenCodeGoResetTime(resp.Usage.Monthly.ResetsAt)},
	}
}

func parseOpenCodeGoResetTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	resetTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}

	return resetTime.Local()
}

func (widget *openCodeGoWidget) Render() template.HTML {
	return widget.renderTemplate(widget, openCodeGoWidgetTemplate)
}
