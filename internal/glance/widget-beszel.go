package glance

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"time"
)

var beszelWidgetTemplate = mustParseTemplate("beszel.html", "widget-base.html")

type beszelWidget struct {
	widgetBase      `yaml:",inline"`
	URL             string         `yaml:"url"`
	Token           string         `yaml:"token"`
	RedirectURL     string         `yaml:"redirect-url"`
	ShowChartsRaw   *bool          `yaml:"show-charts"`
	ShowCharts      bool           `yaml:"-"`
	Systems         []beszelSystem `yaml:"-"`
	OnlineSystems   []beszelSystem `yaml:"-"`
	OfflineSystems  []beszelSystem `yaml:"-"`
}

type beszelResponse struct {
	Items []beszelSystem `json:"items"`
}

type beszelSystem struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Host     string     `json:"host"`
	Status   string     `json:"status"`
	Info     beszelInfo `json:"info"`
	BootTime time.Time  `json:"-"`
}

type beszelInfo struct {
	Kernel       string  `json:"k"`
	Uptime       float64 `json:"u"`
	CPUModel     string  `json:"m"`
	CPU          float64 `json:"cpu"`
	Memory       float64 `json:"mp"`
	Disk         float64 `json:"dp"`
	Load1        float64 `json:"l1"`
	Load5        float64 `json:"l5"`
	Load15       float64 `json:"l15"`
	Cores        int     `json:"c"`
	Threads      int     `json:"t"`
	AgentVersion string  `json:"v"`
	Hostname     string  `json:"h"`
}

// Struktury dla danych wykresów
type beszelChartResponse struct {
	Items []beszelStatsRecord `json:"items"`
}

type beszelStatsRecord struct {
	Created string      `json:"created"`
	Stats   beszelStats `json:"stats"`
}

type beszelStats struct {
	CPU         float64   `json:"cpu"`
	Mem         float64   `json:"m"`
	MemUsed     float64   `json:"mu"`
	MemPct      float64   `json:"mp"`
	DiskTotal   float64   `json:"d"`
	DiskUsed    float64   `json:"du"`
	DiskPct     float64   `json:"dp"`
	NetworkSent float64   `json:"ns"`
	NetworkRecv float64   `json:"nr"`
	LoadAvg     []float64 `json:"la"`
}

type beszelChartData struct {
	Points     string                   `json:"points"`
	Series     []beszelChartSeriesPoint `json:"series"`
	AxisLabels []beszelAxisLabel        `json:"axisLabels"`
	MinValue   float64                  `json:"minValue"`
	MaxValue   float64                  `json:"maxValue"`
}

type beszelChartSeriesPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type beszelAxisLabel struct {
	Label     string  `json:"label"`
	Left      float64 `json:"left"`
	Transform string  `json:"transform"`
}

func (w *beszelWidget) initialize() error {
	w.withTitle("Beszel").withCacheDuration(10 * time.Second)
	if w.URL == "" {
		return errors.New("beszel widget: url is required")
	}
	// Domyślnie wykresy są włączone, chyba że użytkownik ustawił show-charts: false
	if w.ShowChartsRaw == nil {
		w.ShowCharts = true
	} else {
		w.ShowCharts = *w.ShowChartsRaw
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

	// Sortuj systemy - online najpierw, offline na końcu
	sort.Slice(w.Systems, func(i, j int) bool {
		if w.Systems[i].Status == "up" && w.Systems[j].Status != "up" {
			return true
		}
		if w.Systems[i].Status != "up" && w.Systems[j].Status == "up" {
			return false
		}
		return w.Systems[i].Name < w.Systems[j].Name
	})

	// Rozdziel systemy na online i offline
	w.OnlineSystems = nil
	w.OfflineSystems = nil
	for _, sys := range w.Systems {
		if sys.Status == "up" {
			w.OnlineSystems = append(w.OnlineSystems, sys)
		} else {
			w.OfflineSystems = append(w.OfflineSystems, sys)
		}
	}

	w.withError(nil).scheduleNextUpdate()
}

func (w *beszelWidget) Render() template.HTML {
	return w.renderTemplate(w, beszelWidgetTemplate)
}

// FetchChartData pobiera dane wykresu dla konkretnego systemu
func (w *beszelWidget) FetchChartData(systemID string, metricType string, timeRange string) (*beszelChartData, error) {
	// Mapowanie timeRange na typ rekordu w Beszel
	// Beszel przechowuje dane w typach: 1m, 10m, 20m, 120m, 480m
	recordType := "1m"
	var timeFilter string
	now := time.Now().UTC()

	// Format daty zgodny z Beszel API (bez milisekund, bez Z)
	const beszelTimeFormat = "2006-01-02 15:04:05"

	switch timeRange {
	case "1h":
		// Ostatnia godzina - dane co 1m, ~60 punktów
		recordType = "1m"
		timeFilter = now.Add(-1 * time.Hour).Format(beszelTimeFormat)
	case "12h":
		// Ostatnie 12h - dane co 10m, ~72 punkty
		recordType = "10m"
		timeFilter = now.Add(-12 * time.Hour).Format(beszelTimeFormat)
	case "24h":
		// Ostatnie 24h - dane co 20m, ~72 punkty
		recordType = "20m"
		timeFilter = now.Add(-24 * time.Hour).Format(beszelTimeFormat)
	case "7d":
		// Ostatnie 7 dni - dane co 2h (120m), ~84 punkty
		recordType = "120m"
		timeFilter = now.AddDate(0, 0, -7).Format(beszelTimeFormat)
	case "30d":
		// Ostatnie 30 dni - dane co 8h (480m), ~90 punktów
		recordType = "480m"
		timeFilter = now.AddDate(0, 0, -30).Format(beszelTimeFormat)
	default:
		recordType = "1m"
		timeFilter = now.Add(-1 * time.Hour).Format(beszelTimeFormat)
	}

	// Budowanie URL z filtrem - format zgodny z Beszel API
	filter := fmt.Sprintf("system='%s' && created > '%s' && type='%s'", systemID, timeFilter, recordType)
	apiURL := fmt.Sprintf("%s/api/collections/system_stats/records?page=1&perPage=500&skipTotal=1&filter=%s&fields=created,stats&sort=created",
		w.URL, url.QueryEscape(filter))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if w.Token != "" {
		req.Header.Set("Authorization", "Bearer "+w.Token)
	}

	resp, err := decodeJsonFromRequest[*beszelChartResponse](defaultHTTPClient, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, errors.New("no data available")
	}

	// Ekstrakcja wartości na podstawie typu metryki
	values := make([]float64, len(resp.Items))
	series := make([]beszelChartSeriesPoint, len(resp.Items))

	for i, item := range resp.Items {
		var value float64
		switch metricType {
		case "cpu":
			value = item.Stats.CPU
		case "ram":
			value = item.Stats.MemPct
		case "disk":
			value = item.Stats.DiskPct
		case "network":
			value = item.Stats.NetworkSent + item.Stats.NetworkRecv
		default:
			value = item.Stats.CPU
		}
		values[i] = value

		// Parsowanie czasu dla etykiety
		t, _ := time.Parse("2006-01-02 15:04:05.000Z", item.Created)
		if t.IsZero() {
			t, _ = time.Parse(time.RFC3339, item.Created)
		}

		var label string
		switch timeRange {
		case "1h", "12h":
			label = t.Format("15:04")
		case "24h":
			label = t.Format("15:04")
		case "7d", "30d":
			label = t.Format("02.01")
		default:
			label = t.Format("15:04")
		}

		series[i] = beszelChartSeriesPoint{
			Label: label,
			Value: value,
		}
	}

	// Generowanie punktów SVG
	chartPoints := svgPolylineCoordsFromYValues(100, 50, values)

	// Obliczanie min/max dla osi Y
	minValue := slices.Min(values)
	maxValue := slices.Max(values)

	// Generowanie etykiet osi
	axisLabels := generateBeszelAxisLabels(series, timeRange)

	return &beszelChartData{
		Points:     chartPoints,
		Series:     series,
		AxisLabels: axisLabels,
		MinValue:   minValue,
		MaxValue:   maxValue,
	}, nil
}

func generateBeszelAxisLabels(series []beszelChartSeriesPoint, timeRange string) []beszelAxisLabel {
	numPoints := len(series)
	if numPoints < 2 {
		return nil
	}

	var indices []int
	switch timeRange {
	case "1h":
		indices = []int{0, numPoints / 4, numPoints / 2, 3 * numPoints / 4, numPoints - 1}
	case "12h", "24h":
		indices = []int{0, numPoints / 3, 2 * numPoints / 3, numPoints - 1}
	case "7d":
		indices = []int{0, numPoints / 3, 2 * numPoints / 3, numPoints - 1}
	case "30d":
		indices = []int{0, numPoints / 4, numPoints / 2, 3 * numPoints / 4, numPoints - 1}
	default:
		indices = []int{0, numPoints / 2, numPoints - 1}
	}

	labels := make([]beszelAxisLabel, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < numPoints {
			left := (float64(idx) / float64(numPoints-1)) * 100
			transform := "translateX(-50%)"
			if idx == 0 {
				transform = "translateX(0)"
			} else if idx == numPoints-1 {
				transform = "translateX(-100%)"
			}

			labels = append(labels, beszelAxisLabel{
				Label:     series[idx].Label,
				Left:      left,
				Transform: transform,
			})
		}
	}

	return labels
}
