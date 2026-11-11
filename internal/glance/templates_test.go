package glance

import (
	"testing"
	"time"
)

func withWarsawDisplayLocation(t *testing.T) {
	t.Helper()

	originalLocation := appTimeLocation
	originalLocal := time.Local

	warsaw, err := time.LoadLocation(defaultTimeZone)
	if err != nil {
		t.Fatalf("could not load %s: %v", defaultTimeZone, err)
	}

	appTimeLocation = warsaw
	time.Local = time.UTC

	t.Cleanup(func() {
		appTimeLocation = originalLocation
		time.Local = originalLocal
	})
}

func TestFormatPolishDateUsesDisplayTimeZone(t *testing.T) {
	withWarsawDisplayLocation(t)

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "winter time crossing midnight",
			input:    time.Date(2026, time.January, 15, 23, 30, 0, 0, time.UTC),
			expected: "piątek, 16 stycznia 2026",
		},
		{
			name:     "summer time crossing midnight",
			input:    time.Date(2026, time.July, 15, 22, 30, 0, 0, time.UTC),
			expected: "czwartek, 16 lipca 2026",
		},
		{
			name:     "same day",
			input:    time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC),
			expected: "czwartek, 15 stycznia 2026",
		},
		{
			name:     "zero timestamp",
			input:    time.Unix(0, 0),
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := formatPolishDate(test.input); got != test.expected {
				t.Fatalf("formatPolishDate(%s) = %q, want %q", test.input, got, test.expected)
			}
		})
	}
}

func TestFormatPolishDateTimeUsesDisplayTimeZone(t *testing.T) {
	withWarsawDisplayLocation(t)

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "winter time crossing midnight",
			input:    time.Date(2026, time.January, 15, 23, 30, 0, 0, time.UTC),
			expected: "piątek, 16 stycznia 2026, 00:30",
		},
		{
			name:     "summer time crossing midnight",
			input:    time.Date(2026, time.July, 15, 22, 30, 0, 0, time.UTC),
			expected: "czwartek, 16 lipca 2026, 00:30",
		},
		{
			name:     "same day",
			input:    time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC),
			expected: "czwartek, 15 stycznia 2026, 13:00",
		},
		{
			name:     "zero timestamp",
			input:    time.Unix(0, 0),
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := formatPolishDateTime(test.input); got != test.expected {
				t.Fatalf("formatPolishDateTime(%s) = %q, want %q", test.input, got, test.expected)
			}
		})
	}
}

func TestResolveAppTimeLocation(t *testing.T) {
	originalLocal := time.Local
	t.Cleanup(func() {
		time.Local = originalLocal
	})

	t.Run("respects TZ from environment", func(t *testing.T) {
		t.Setenv("TZ", "Europe/London")
		time.Local = time.UTC

		if got := resolveAppTimeLocation().String(); got != "Europe/London" {
			t.Fatalf("resolveAppTimeLocation() = %q, want %q", got, "Europe/London")
		}
	})

	t.Run("falls back to default time zone when host is UTC", func(t *testing.T) {
		t.Setenv("TZ", "")
		time.Local = time.UTC

		if got := resolveAppTimeLocation().String(); got != defaultTimeZone {
			t.Fatalf("resolveAppTimeLocation() = %q, want %q", got, defaultTimeZone)
		}
	})
}
