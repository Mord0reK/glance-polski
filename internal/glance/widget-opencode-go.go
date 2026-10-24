package glance

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"time"
)

var openCodeGoWidgetTemplate = mustParseTemplate("opencode-go.html", "widget-base.html")

const (
	defaultOpenCodeGoURL              = "https://opencode.ai/zen/go/v1/usage"
	defaultOpenCodeGoTitleURL         = "https://opencode.ai/console"
	defaultOpenCodeGoWarningThreshold = 85.0

	// The public API floors the percentage to a whole number, while the console
	// uses the exact ratio. Using the midpoint of the flooring interval removes
	// the systematic downward bias from the forecast.
	openCodeGoPercentEstimateOffset = 0.5

	// The console only shows a forecast once at least a day of the cycle has
	// elapsed; a forecast based on a few minutes of data would be meaningless.
	openCodeGoMinElapsedForForecast = 24 * time.Hour
)

type openCodeGoWidget struct {
	widgetBase       `yaml:",inline"`
	URL              string             `yaml:"url"`
	APIKey           string             `yaml:"api-key"`
	WarningThreshold float64            `yaml:"warning-threshold"`
	ShowForecastRaw  *bool              `yaml:"show-forecast"`
	ShowForecast     bool               `yaml:"-"`
	Windows          []openCodeGoWindow `yaml:"-"`
}

// openCodeGoWindow is one usage window (rolling / weekly / monthly) prepared
// for rendering by the template.
type openCodeGoWindow struct {
	Label    string
	Status   string
	Percent  float64
	ResetsAt time.Time

	// Forecast is only set for the monthly window and only when ShowForecast is
	// enabled and there is enough data to extrapolate from.
	HasForecast  bool
	Forecast     float64
	DailyAverage float64
}

// ForecastPercent mirrors the console, which rounds the forecast to a whole
// percentage point with Math.round.
func (w openCodeGoWindow) ForecastPercent() int {
	return int(math.Round(w.Forecast))
}

// DailyAveragePercent rounds the average daily usage to one decimal place.
func (w openCodeGoWindow) DailyAveragePercent() float64 {
	return math.Round(w.DailyAverage*10) / 10
}

// ForecastWidth returns the width, in percentage points, of the hatched
// forecast extension drawn on top of the current usage bar.
func (w openCodeGoWindow) ForecastWidth() float64 {
	if !w.HasForecast || w.Percent >= 100 {
		return 0
	}

	width := w.Forecast - w.Percent
	if width <= 0 {
		return 0
	}

	if w.Percent+width > 100 {
		width = 100 - w.Percent
	}

	return width
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
	widget.
		withTitle("OpenCode Go").
		withTitleURL(defaultOpenCodeGoTitleURL).
		withCacheDuration(5 * time.Minute)

	if widget.URL == "" {
		widget.URL = defaultOpenCodeGoURL
	}

	if widget.APIKey == "" {
		return errors.New("api-key is required")
	}

	if widget.WarningThreshold == 0 {
		widget.WarningThreshold = defaultOpenCodeGoWarningThreshold
	}

	if widget.ShowForecastRaw == nil {
		widget.ShowForecast = true
	} else {
		widget.ShowForecast = *widget.ShowForecastRaw
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

	if widget.ShowForecast {
		monthly := &widget.Windows[2]
		monthly.Forecast, monthly.DailyAverage, monthly.HasForecast = calculateOpenCodeGoForecast(
			monthly.Percent,
			monthly.ResetsAt,
			time.Now(),
		)
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

// calculateOpenCodeGoForecast estimates where the monthly usage will end up at
// the end of the current billing cycle.
//
// The OpenCode console computes this as
//
//	usedPercent * (resetsAt - startsAt) / (now - startsAt)
//
// where usedPercent comes from the exact micro-cent usage and startsAt is the
// exact start of the billing period. The public API only exposes the percentage
// floored to a whole number and the reset time, so the cycle start is derived by
// subtracting one month from resetsAt and the floored percentage is corrected by
// half a point to remove the systematic downward bias.
func calculateOpenCodeGoForecast(percent float64, resetsAt, now time.Time) (forecast, dailyAverage float64, ok bool) {
	if percent <= 0 || resetsAt.IsZero() {
		return 0, 0, false
	}

	cycleStart := subtractOneMonthClamped(resetsAt)
	total := resetsAt.Sub(cycleStart)
	elapsed := now.Sub(cycleStart)

	if total <= 0 || elapsed < openCodeGoMinElapsedForForecast {
		return 0, 0, false
	}

	if elapsed > total {
		elapsed = total
	}

	elapsedDays := elapsed.Hours() / 24

	// The console pins the forecast to 100% once the limit is reached.
	if percent >= 100 {
		return 100, 100 / elapsedDays, true
	}

	estimatedPercent := percent + openCodeGoPercentEstimateOffset
	totalDays := total.Hours() / 24
	dailyAverage = estimatedPercent / elapsedDays
	forecast = dailyAverage * totalDays

	return forecast, dailyAverage, true
}

// subtractOneMonthClamped subtracts one month from t, clamping the day to the
// last day of the target month so that e.g. March 31st becomes February 28th
// instead of spilling over into March.
func subtractOneMonthClamped(t time.Time) time.Time {
	year, month, day := t.Date()

	targetYear, targetMonth := year, month-1
	if targetMonth < time.January {
		targetMonth = time.December
		targetYear--
	}

	lastDay := time.Date(targetYear, targetMonth+1, 0, 0, 0, 0, 0, t.Location()).Day()
	if day > lastDay {
		day = lastDay
	}

	return time.Date(targetYear, targetMonth, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func (widget *openCodeGoWidget) Render() template.HTML {
	return widget.renderTemplate(widget, openCodeGoWidgetTemplate)
}
