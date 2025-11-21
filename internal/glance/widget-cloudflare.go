package glance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var cloudflareWidgetTemplate = mustParseTemplate("cloudflare.html", "widget-base.html")

type cloudflareWidget struct {
	widgetBase `yaml:",inline"`
	ApiKey     string          `yaml:"api-key"`
	ZoneID     string          `yaml:"zone-id"`
	TimeRange  string          `yaml:"time-range"` // "24h", "7d", "30d"
	Data       *cloudflareData `yaml:"-"`
}

type cloudflareData struct {
	TotalRequests  int
	UniqueVisitors int
	ChartPoints    string
	Series         []cloudflareSeriesPoint
	AxisLabels     []cloudflareAxisLabel
}

type cloudflareAxisLabel struct {
	Label     string
	Left      float64
	Transform string
}

type cloudflareSeriesPoint struct {
	Label    string `json:"label"`
	Requests int    `json:"requests"`
	Uniques  int    `json:"uniques"`
}

func (widget *cloudflareWidget) initialize() error {
	widget.withTitle("Cloudflare").withCacheDuration(15 * time.Minute)

	if widget.ApiKey == "" {
		return fmt.Errorf("api-key is required")
	}
	if widget.ZoneID == "" {
		return fmt.Errorf("zone-id is required")
	}
	if widget.TimeRange == "" {
		widget.TimeRange = "24h"
	}

	return nil
}

func (widget *cloudflareWidget) update(ctx context.Context) {
	data, err := fetchCloudflareData(widget.ApiKey, widget.ZoneID, widget.TimeRange)
	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}
	widget.Data = data
}

func (widget *cloudflareWidget) Render() template.HTML {
	return widget.renderTemplate(widget, cloudflareWidgetTemplate)
}

// GraphQL structs
type cloudflareGraphQLResponse struct {
	Data struct {
		Viewer struct {
			Zones []struct {
				HttpRequests1dGroups []cloudflareGroup `json:"httpRequests1dGroups"`
				HttpRequests1hGroups []cloudflareGroup `json:"httpRequests1hGroups"`
			} `json:"zones"`
		} `json:"viewer"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type cloudflareGroup struct {
	Dimensions struct {
		Date     string `json:"date"`
		Datetime string `json:"datetime"`
	} `json:"dimensions"`
	Sum struct {
		Requests int `json:"requests"`
	} `json:"sum"`
	Uniq struct {
		Uniques int `json:"uniques"`
	} `json:"uniq"`
}

func fetchCloudflareData(apiKey, zoneID, timeRange string) (*cloudflareData, error) {
	// Construct GraphQL query
	var query string
	var limit int
	var dateFilter string

	now := time.Now()

	if timeRange == "24h" {
		limit = 24
		dateFilter = now.Add(-24 * time.Hour).Format(time.RFC3339)
		query = fmt.Sprintf(`
			query {
				viewer {
					zones(filter: {zoneTag: "%s"}) {
						httpRequests1hGroups(limit: %d, filter: {datetime_geq: "%s"}, orderBy: [datetime_ASC]) {
							dimensions {
								datetime
							}
							sum {
								requests
							}
							uniq {
								uniques
							}
						}
					}
				}
			}
		`, zoneID, limit, dateFilter)
	} else {
		if timeRange == "7d" {
			limit = 7
			dateFilter = now.AddDate(0, 0, -7).Format("2006-01-02")
		} else { // 30d
			limit = 30
			dateFilter = now.AddDate(0, 0, -30).Format("2006-01-02")
		}
		query = fmt.Sprintf(`
			query {
				viewer {
					zones(filter: {zoneTag: "%s"}) {
						httpRequests1dGroups(limit: %d, filter: {date_geq: "%s"}, orderBy: [date_ASC]) {
							dimensions {
								date
							}
							sum {
								requests
							}
							uniq {
								uniques
							}
						}
					}
				}
			}
		`, zoneID, limit, dateFilter)
	}

	reqBody, _ := json.Marshal(map[string]string{"query": query})
	req, _ := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/graphql", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := decodeJsonFromRequest[cloudflareGraphQLResponse](defaultHTTPClient, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("cloudflare api error: %s", resp.Errors[0].Message)
	}

	if len(resp.Data.Viewer.Zones) == 0 {
		return nil, fmt.Errorf("zone not found")
	}

	zone := resp.Data.Viewer.Zones[0]
	var groups []cloudflareGroup
	if timeRange == "24h" {
		groups = zone.HttpRequests1hGroups
	} else {
		groups = zone.HttpRequests1dGroups
	}

	totalRequests := 0
	uniqueVisitors := 0
	requestsSeries := make([]float64, len(groups))
	series := make([]cloudflareSeriesPoint, len(groups))

	for i, g := range groups {
		totalRequests += g.Sum.Requests
		uniqueVisitors += g.Uniq.Uniques
		requestsSeries[i] = float64(g.Sum.Requests)

		var label string
		if timeRange == "24h" {
			t, _ := time.Parse(time.RFC3339, g.Dimensions.Datetime)
			label = t.Format("15")
		} else {
			t, _ := time.Parse("2006-01-02", g.Dimensions.Date)
			label = t.Format("02.01")
		}

		series[i] = cloudflareSeriesPoint{
			Label:    label,
			Requests: g.Sum.Requests,
			Uniques:  g.Uniq.Uniques,
		}
	}

	chartPoints := svgPolylineCoordsFromYValues(100, 50, requestsSeries)

	var axisLabels []cloudflareAxisLabel
	numPoints := len(series)
	if numPoints > 1 {
		// Determine indices to show
		var indices []int
		if timeRange == "24h" {
			// 0, 6, 12, 18, last
			indices = []int{0, 6, 12, 18, numPoints - 1}
		} else if timeRange == "7d" {
			// 0, 2, 4, 6
			indices = []int{0, 2, 4, 6}
		} else { // 30d
			// 0, 7, 14, 21, last
			indices = []int{0, 7, 14, 21, numPoints - 1}
		}

		for _, idx := range indices {
			if idx >= 0 && idx < numPoints {
				left := (float64(idx) / float64(numPoints-1)) * 100
				transform := "translateX(-50%)"
				if idx == 0 {
					transform = "translateX(0)"
				} else if idx == numPoints-1 {
					transform = "translateX(-100%)"
				}

				axisLabels = append(axisLabels, cloudflareAxisLabel{
					Label:     series[idx].Label,
					Left:      left,
					Transform: transform,
				})
			}
		}
	}

	return &cloudflareData{
		TotalRequests:  totalRequests,
		UniqueVisitors: uniqueVisitors,
		ChartPoints:    chartPoints,
		Series:         series,
		AxisLabels:     axisLabels,
	}, nil
}
