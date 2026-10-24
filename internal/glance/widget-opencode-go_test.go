package glance

import (
	"strings"
	"testing"
	"time"
)

func TestParseOpenCodeGoResetTime(t *testing.T) {
	parsed := parseOpenCodeGoResetTime("2026-08-13T16:27:38.287Z")
	if parsed.IsZero() {
		t.Fatal("expected a parsed reset time")
	}

	if got := parsed.UTC().Format(time.RFC3339Nano); got != "2026-08-13T16:27:38.287Z" {
		t.Errorf("unexpected reset time: %s", got)
	}

	if !parseOpenCodeGoResetTime("").IsZero() {
		t.Error("expected empty value to produce a zero time")
	}

	if !parseOpenCodeGoResetTime("not-a-time").IsZero() {
		t.Error("expected invalid value to produce a zero time")
	}
}

func TestOpenCodeGoWidgetInitializeDefaults(t *testing.T) {
	widget := &openCodeGoWidget{APIKey: "test-key"}

	if err := widget.initialize(); err != nil {
		t.Fatalf("initialize returned an error: %v", err)
	}

	if widget.Title != "OpenCode Go" {
		t.Errorf("unexpected default title: %q", widget.Title)
	}

	if widget.TitleURL != defaultOpenCodeGoTitleURL {
		t.Errorf("unexpected default title url: %q", widget.TitleURL)
	}

	if !widget.ShowForecast {
		t.Error("expected the forecast to be enabled by default")
	}
}

func TestCalculateOpenCodeGoForecast(t *testing.T) {
	// Cycle: 12 Oct -> 12 Nov (31 days), 24% used after 12 days.
	// The API floors the percentage, so the widget estimates 24.5%:
	// 24.5% / 12 days = 2.04%/day, so the cycle should end at about 63.3%.
	resetsAt := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)

	forecast, dailyAverage, ok := calculateOpenCodeGoForecast(24, resetsAt, now)
	if !ok {
		t.Fatal("expected a forecast to be calculated")
	}

	if dailyAverage < 2.04 || dailyAverage > 2.05 {
		t.Errorf("unexpected daily average: %f", dailyAverage)
	}

	if forecast < 63.2 || forecast > 63.4 {
		t.Errorf("unexpected forecast: %f", forecast)
	}
}

func TestCalculateOpenCodeGoForecastAtLimit(t *testing.T) {
	// At the limit the console reports exactly 100%, without any floor correction.
	resetsAt := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)

	forecast, dailyAverage, ok := calculateOpenCodeGoForecast(100, resetsAt, now)
	if !ok {
		t.Fatal("expected a forecast to be calculated")
	}

	if forecast < 99.99 || forecast > 100.01 {
		t.Errorf("expected the forecast to be capped at 100, got %f", forecast)
	}

	if dailyAverage < 8.33 || dailyAverage > 8.34 {
		t.Errorf("unexpected daily average: %f", dailyAverage)
	}
}

func TestCalculateOpenCodeGoForecastEdgeCases(t *testing.T) {
	resetsAt := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

	if _, _, ok := calculateOpenCodeGoForecast(0, resetsAt, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)); ok {
		t.Error("expected no forecast for zero usage")
	}

	if _, _, ok := calculateOpenCodeGoForecast(10, time.Time{}, time.Now()); ok {
		t.Error("expected no forecast without a reset time")
	}

	// Too early in the cycle to extrapolate reliably.
	if _, _, ok := calculateOpenCodeGoForecast(1, resetsAt, time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC)); ok {
		t.Error("expected no forecast right after the cycle start")
	}
}

func TestSubtractOneMonthClamped(t *testing.T) {
	tests := []struct {
		in   time.Time
		want time.Time
	}{
		{time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)},
		{time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)},
		{time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		if got := subtractOneMonthClamped(test.in); !got.Equal(test.want) {
			t.Errorf("subtractOneMonthClamped(%s) = %s, want %s", test.in, got, test.want)
		}
	}
}

func TestOpenCodeGoWindowRounding(t *testing.T) {
	window := openCodeGoWindow{Forecast: 60.5, DailyAverage: 2.0416667}

	if got := window.ForecastPercent(); got != 61 {
		t.Errorf("ForecastPercent() = %d, want 61", got)
	}

	if got := window.DailyAveragePercent(); got != 2.0 {
		t.Errorf("DailyAveragePercent() = %f, want 2.0", got)
	}
}

func TestOpenCodeGoWindowForecastWidth(t *testing.T) {
	tests := []struct {
		name   string
		window openCodeGoWindow
		want   float64
	}{
		{"regular", openCodeGoWindow{HasForecast: true, Percent: 24, Forecast: 61}, 37},
		{"capped at 100", openCodeGoWindow{HasForecast: true, Percent: 91, Forecast: 112}, 9},
		{"below current", openCodeGoWindow{HasForecast: true, Percent: 50, Forecast: 40}, 0},
		{"at limit", openCodeGoWindow{HasForecast: true, Percent: 100, Forecast: 100}, 0},
		{"no forecast", openCodeGoWindow{Percent: 24, Forecast: 61}, 0},
	}

	for _, test := range tests {
		if got := test.window.ForecastWidth(); got != test.want {
			t.Errorf("%s: ForecastWidth() = %f, want %f", test.name, got, test.want)
		}
	}
}

func TestOpenCodeGoWidgetRender(t *testing.T) {
	widget := &openCodeGoWidget{
		WarningThreshold: defaultOpenCodeGoWarningThreshold,
		ShowForecast:     true,
		Windows: []openCodeGoWindow{
			{Label: "5 godzin", Status: "ok", Percent: 0, ResetsAt: time.Date(2026, 8, 13, 16, 27, 0, 0, time.UTC)},
			{Label: "Tydzień", Status: "ok", Percent: 12.5, ResetsAt: time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)},
			{Label: "Miesiąc", Status: "rate-limited", Percent: 91, HasForecast: true, Forecast: 112, DailyAverage: 3.2},
		},
	}
	widget.Title = "OpenCode Go"
	widget.TitleURL = defaultOpenCodeGoTitleURL
	widget.ContentAvailable = true

	rendered := string(widget.Render())

	for _, expected := range []string{"5 godzin", "Tydzień", "Miesiąc", "13", "91", "opencode-go-forecast", "prognoza", "112%", "3.2%/dzień", "opencode-go-forecast-bar"} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("expected rendered output to contain %q", expected)
		}
	}

	if !strings.Contains(rendered, `href="`+defaultOpenCodeGoTitleURL+`"`) {
		t.Error("expected the title to link to the OpenCode console")
	}

	if !strings.Contains(rendered, "opencode-go-bar-danger") {
		t.Error("expected a rate-limited window to use the danger bar style")
	}

	if !strings.Contains(rendered, "opencode-go-forecast-danger") {
		t.Error("expected a forecast above 100% to use the danger style")
	}
}
