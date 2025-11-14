package glance

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

var radyjkoWidgetTemplate = mustParseTemplate("radyjko.html", "widget-base.html")

type radyjkoWidget struct {
	widgetBase `yaml:",inline"`
	Stations   stationList `yaml:"-"`
}

type station struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	URL       string `json:"url"`
	Icon      string `json:"icon,omitempty"`
	IconURL   string `yaml:"-"`
	IsOpenFM  int    `json:"isOpenFM"`
	OpenFMID  *int   `json:"openFmId"`
}

type stationList []station

func (s *station) GetIconURL() string {
	if s.ShortName == "" {
		return ""
	}
	return "/static/images/radyjko/" + s.ShortName + ".png"
}

func (widget *radyjkoWidget) initialize() error {
	widget.withTitle("Radyjko").withCacheDuration(time.Hour * 24)
	return nil
}

func (widget *radyjkoWidget) update(ctx context.Context) {
	stations, err := fetchRadioStations(ctx)

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Stations = stations
}

func (widget *radyjkoWidget) Render() template.HTML {
	// Compute icon URLs using the asset resolver
	for i := range widget.Stations {
		widget.Stations[i].IconURL = widget.Providers.assetResolver("images/radyjko/" + widget.Stations[i].ShortName + ".png")

		// Handle OpenFM stations - construct streaming URL
		if widget.Stations[i].IsOpenFM == 1 && widget.Stations[i].OpenFMID != nil {
			widget.Stations[i].URL = fmt.Sprintf(
				"https://stream-cdn-1.open.fm/OFM%d/ngrp:standard/playlist.m3u8",
				*widget.Stations[i].OpenFMID,
			)
		}
	}
	return widget.renderTemplate(widget, radyjkoWidgetTemplate)
}

func fetchRadioStations(ctx context.Context) (stationList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://proxy.mordorek.dev/stations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch stations: status code %d", resp.StatusCode)
	}

	var stations stationList
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		slog.Error("failed to decode stations response", "error", err)
		return nil, fmt.Errorf("failed to decode stations: %w", err)
	}

	return stations, nil
}
