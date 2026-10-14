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

func TestOpenCodeGoWidgetRender(t *testing.T) {
	widget := &openCodeGoWidget{
		WarningThreshold: defaultOpenCodeGoWarningThreshold,
		Windows: []openCodeGoWindow{
			{Label: "5 godzin", Status: "ok", Percent: 0, ResetsAt: time.Date(2026, 8, 13, 16, 27, 0, 0, time.UTC)},
			{Label: "Tydzień", Status: "ok", Percent: 12.5, ResetsAt: time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)},
			{Label: "Miesiąc", Status: "rate-limited", Percent: 91},
		},
	}
	widget.Title = "OpenCode Go"
	widget.ContentAvailable = true

	rendered := string(widget.Render())

	for _, expected := range []string{"5 godzin", "Tydzień", "Miesiąc", "13", "91"} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("expected rendered output to contain %q", expected)
		}
	}

	if !strings.Contains(rendered, "var(--color-negative)") {
		t.Error("expected a rate-limited window to use the negative color")
	}
}
