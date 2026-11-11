package glance

import (
	"os"
	"strings"
	"time"
	_ "time/tzdata"
)

// defaultTimeZone is used when the host timezone cannot be determined.
// Glance is often run in a container where time.Local defaults to UTC,
// which would otherwise shift timestamps returned by external APIs.
const defaultTimeZone = "Europe/Warsaw"

// appTimeLocation is the timezone used to display absolute timestamps
// returned by external APIs. It respects the host's TZ setting, but falls
// back to the app's default timezone inside containers without TZ.
var appTimeLocation = resolveAppTimeLocation()

func resolveAppTimeLocation() *time.Location {
	if tz := strings.TrimSpace(os.Getenv("TZ")); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}

	if time.Local.String() != "UTC" {
		return time.Local
	}

	if loc, err := time.LoadLocation(defaultTimeZone); err == nil {
		return loc
	}

	return time.UTC
}
