package glance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"
)

var cloudflareWidgetTemplate = mustParseTemplate("cloudflare.html", "widget-base.html")

// Use explicit Europe/Warsaw timezone instead of time.Local to ensure consistent behavior
var defaultLocation *time.Location

func init() {
	var err error
	defaultLocation, err = time.LoadLocation("Europe/Warsaw")
	if err != nil {
		defaultLocation = time.UTC
	}
}

type cloudflareWidget struct {
	widgetBase `yaml:",inline"`
	ApiKey     string          `yaml:"api-key"`
	ZoneID     string          `yaml:"zone-id"`
	TimeRange  string          `yaml:"time-range"` // "24h", "7d", "30d"
	Data       *cloudflareData `yaml:"-"`
}

type cloudflareData struct {
	TotalRequests    int
	CachedRequests   int
	UncachedRequests int
	Threats          int
	ChartPoints      string
	Series           []cloudflareSeriesPoint
	AxisLabels       []cloudflareAxisLabel
}

type cloudflareAxisLabel struct {
	Label     string
	Left      float64
	Transform string
}

type cloudflareSeriesPoint struct {
	Label            string `json:"label"`
	Timestamp        string `json:"timestamp"`
	Requests         int    `json:"requests"`
	CachedRequests   int    `json:"cachedRequests"`
	UncachedRequests int    `json:"uncachedRequests"`
	Threats          int    `json:"threats"`
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

type cloudflareSecurityResponse struct {
	Data struct {
		Viewer struct {
			Scope []struct {
				MitigatedByWAF     []cloudflareSecurityGroup `json:"mitigatedByWAF"`
				ServedByCloudflare []cloudflareSecurityGroup `json:"servedByCloudflare"`
				ServedByOrigin     []cloudflareSecurityGroup `json:"servedByOrigin"`
			} `json:"scope"`
		} `json:"viewer"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type cloudflareSecurityGroup struct {
	Count int `json:"count"`
	Avg   struct {
		SampleInterval float64 `json:"sampleInterval"`
	} `json:"avg"`
	Dimensions struct {
		Ts string `json:"ts"`
	} `json:"dimensions"`
}

func fetchCloudflareData(apiKey, zoneID, timeRange string) (*cloudflareData, error) {
	now := time.Now().In(defaultLocation)

	startTime := now.Add(-24 * time.Hour).Format(time.RFC3339)
	endTime := now.Format(time.RFC3339)
	filterBase := fmt.Sprintf(`datetime_geq: "%s", datetime_leq: "%s", requestSource: "eyeball"`, startTime, endTime)
	aggInterval := "datetimeFifteenMinutes"

	mitigatedFilter := fmt.Sprintf(`{%s, securityAction_in: ["block", "challenge", "jschallenge", "managed_challenge"]}`, filterBase)
	servedByCloudflareFilter := fmt.Sprintf(`{%s, securityAction_notin: ["block", "challenge", "jschallenge", "managed_challenge"], cacheStatus_notin: ["miss", "expired", "bypass", "dynamic"]}`, filterBase)
	servedByOriginFilter := fmt.Sprintf(`{%s, cacheStatus_in: ["miss", "expired", "bypass", "dynamic"]}`, filterBase)

	query := fmt.Sprintf(`
		query SecurityAnalyticsTimeseries($zoneTag: string) {
			viewer {
				scope: zones(filter: {zoneTag: $zoneTag}) {
					mitigatedByWAF: httpRequestsAdaptiveGroups(limit: 5000, filter: %s, orderBy: [%s_DESC]) {
						count
						avg {
							sampleInterval
						}
						dimensions {
							ts: %s
						}
					}
					servedByCloudflare: httpRequestsAdaptiveGroups(limit: 5000, filter: %s, orderBy: [%s_DESC]) {
						count
						avg {
							sampleInterval
						}
						dimensions {
							ts: %s
						}
					}
					servedByOrigin: httpRequestsAdaptiveGroups(limit: 5000, filter: %s, orderBy: [%s_DESC]) {
						count
						avg {
							sampleInterval
						}
						dimensions {
							ts: %s
						}
					}
				}
			}
		}
	`, mitigatedFilter, aggInterval, aggInterval, servedByCloudflareFilter, aggInterval, aggInterval, servedByOriginFilter, aggInterval, aggInterval)

	variables := map[string]string{
		"zoneTag": zoneID,
	}

	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/graphql", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := decodeJsonFromRequest[cloudflareSecurityResponse](defaultHTTPClient, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("cloudflare api error: %s", resp.Errors[0].Message)
	}

	if len(resp.Data.Viewer.Scope) == 0 {
		return nil, fmt.Errorf("zone not found")
	}

	zone := resp.Data.Viewer.Scope[0]

	mitigated := zone.MitigatedByWAF
	servedByCF := zone.ServedByCloudflare
	servedByOrigin := zone.ServedByOrigin

	var totalMitigated, totalServedByCF, totalServedByOrigin int

	for _, g := range mitigated {
		count := g.Count
		totalMitigated += count
	}

	for _, g := range servedByCF {
		count := g.Count
		totalServedByCF += count
	}

	for _, g := range servedByOrigin {
		count := g.Count
		totalServedByOrigin += count
	}

	totalRequests := totalMitigated + totalServedByCF + totalServedByOrigin

	series := buildCloudflareSeries(mitigated, servedByCF, servedByOrigin, timeRange)

	requestsSeries := make([]float64, len(series))
	for i, s := range series {
		requestsSeries[i] = float64(s.Requests)
	}

	chartPoints := svgPolylineCoordsFromYValues(100, 50, requestsSeries)
	axisLabels := buildCloudflareAxisLabels(series, timeRange)

	return &cloudflareData{
		TotalRequests:    totalRequests,
		CachedRequests:   totalServedByCF,
		UncachedRequests: totalServedByOrigin,
		Threats:          totalMitigated,
		ChartPoints:      chartPoints,
		Series:           series,
		AxisLabels:       axisLabels,
	}, nil
}

func buildCloudflareSeries(mitigated, servedByCF, servedByOrigin []cloudflareSecurityGroup, timeRange string) []cloudflareSeriesPoint {
	type timeData struct {
		mitigated      int
		servedByCF     int
		servedByOrigin int
	}

	timeMap := make(map[string]timeData)

	// Pre-populate timeMap with 15-minute intervals to prevent time jumping
	now := time.Now().UTC()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), (now.Minute()/15)*15, 0, 0, time.UTC)
	for i := 0; i <= 96; i++ {
		ts := endTime.Add(-time.Duration(i*15) * time.Minute).Format("2006-01-02T15:04:05Z")
		timeMap[ts] = timeData{}
	}

	for _, g := range mitigated {
		ts := g.Dimensions.Ts
		count := g.Count
		if d, ok := timeMap[ts]; ok {
			d.mitigated += count
			timeMap[ts] = d
		} else {
			timeMap[ts] = timeData{mitigated: count}
		}
	}

	for _, g := range servedByCF {
		ts := g.Dimensions.Ts
		count := g.Count
		if d, ok := timeMap[ts]; ok {
			d.servedByCF += count
			timeMap[ts] = d
		} else {
			timeMap[ts] = timeData{servedByCF: count}
		}
	}

	for _, g := range servedByOrigin {
		ts := g.Dimensions.Ts
		count := g.Count
		if d, ok := timeMap[ts]; ok {
			d.servedByOrigin += count
			timeMap[ts] = d
		} else {
			timeMap[ts] = timeData{servedByOrigin: count}
		}
	}

	var times []string
	for ts := range timeMap {
		times = append(times, ts)
	}

	sortByTimeDesc(times)

	series := make([]cloudflareSeriesPoint, len(times))
	for i, ts := range times {
		d := timeMap[ts]
		label := formatTimeLabel(ts, timeRange)

		series[i] = cloudflareSeriesPoint{
			Label:            label,
			Timestamp:        ts,
			Requests:         d.mitigated + d.servedByCF + d.servedByOrigin,
			CachedRequests:   d.servedByCF,
			UncachedRequests: d.servedByOrigin,
			Threats:          d.mitigated,
		}
	}

	return series
}

func sortByTimeDesc(times []string) {
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
}

func sortByDateDesc(times []string) {
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
}

func formatTimeLabel(ts, timeRange string) string {
	if timeRange == "24h" {
		if len(ts) >= 19 {
			t, err := time.Parse("2006-01-02T15:04:05Z", ts)
			if err == nil {
				hour := t.Hour()
				return fmt.Sprintf("%02d", hour)
			}
		}
		if len(ts) >= 13 {
			return ts[11:13]
		}
		return ts
	} else {
		if len(ts) >= 10 {
			t, err := time.Parse("2006-01-02", ts[:10])
			if err == nil {
				return t.Format("02.01")
			}
		}
		return ts
	}
}

func buildCloudflareAxisLabels(series []cloudflareSeriesPoint, timeRange string) []cloudflareAxisLabel {
	var axisLabels []cloudflareAxisLabel
	numPoints := len(series)

	if numPoints > 1 {
		step := numPoints / 5
		if step < 1 {
			step = 1
		}
		indices := []int{0}
		for i := step; i < numPoints-1; i += step {
			indices = append(indices, i)
		}
		indices = append(indices, numPoints-1)

		labelSeen := make(map[string]bool)
		for _, idx := range indices {
			if idx >= 0 && idx < numPoints {
				label := series[idx].Label
				left := (float64(idx) / float64(numPoints-1)) * 100
				transform := "translateX(-50%)"
				if idx == 0 {
					transform = "translateX(0)"
				} else if idx == numPoints-1 {
					transform = "translateX(-100%)"
				}

				if !labelSeen[label] {
					labelSeen[label] = true
					axisLabels = append(axisLabels, cloudflareAxisLabel{
						Label:     label,
						Left:      left,
						Transform: transform,
					})
				}
			}
		}
	}

	return axisLabels
}
