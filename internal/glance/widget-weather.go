package glance

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	_ "time/tzdata"
)

var weatherWidgetTemplate = mustParseTemplate("weather.html", "widget-base.html")

type weatherWidget struct {
	widgetBase   `yaml:",inline"`
	Location     string                      `yaml:"location"`
	ShowAreaName bool                        `yaml:"show-area-name"`
	HideLocation bool                        `yaml:"hide-location"`
	HourFormat   string                      `yaml:"hour-format"`
	Units        string                      `yaml:"units"`
	UseBrightSky bool                        `yaml:"use-brightsky"`
	Lat          float64                     `yaml:"lat"`
	Lon          float64                     `yaml:"lon"`
	Place        *openMeteoPlaceResponseJson `yaml:"-"`
	Weather      *weather                    `yaml:"-"`
	TimeLabels   [12]string                  `yaml:"-"`
}

var timeLabels12h = [12]string{"2am", "4am", "6am", "8am", "10am", "12pm", "2pm", "4pm", "6pm", "8pm", "10pm", "12am"}
var timeLabels24h = [12]string{"2:00", "4:00", "6:00", "8:00", "10:00", "12:00", "14:00", "16:00", "18:00", "20:00", "22:00", "00:00"}

func (widget *weatherWidget) initialize() error {
	widget.withTitle("Weather").withCacheOnTheHour()

	if !widget.UseBrightSky && widget.Location == "" {
		return fmt.Errorf("location is required")
	}

	if widget.HourFormat == "" || widget.HourFormat == "12h" {
		widget.TimeLabels = timeLabels12h
	} else if widget.HourFormat == "24h" {
		widget.TimeLabels = timeLabels24h
	} else {
		return errors.New("hour-format must be either 12h or 24h")
	}

	if widget.Units == "" {
		widget.Units = "metric"
	} else if widget.Units != "metric" && widget.Units != "imperial" {
		return errors.New("units must be either metric or imperial")
	}

	return nil
}

func (widget *weatherWidget) update(ctx context.Context) {
	if widget.UseBrightSky {
		weather, stationName, err := fetchWeatherFromBrightSky(widget.Lat, widget.Lon, widget.Units)
		if !widget.canContinueUpdateAfterHandlingErr(err) {
			return
		}

		if widget.Place == nil {
			widget.Place = &openMeteoPlaceResponseJson{
				Name:    widget.Location,
				Country: fmt.Sprintf("%.2f, %.2f", widget.Lat, widget.Lon),
			}
		}

		if stationName != "" && widget.Location == "" {
			widget.Place.Name = strings.Title(strings.ToLower(stationName))
			widget.Place.Country = ""
		}

		if widget.Place.Name == "" {
			widget.Place.Name = "Bright Sky"
		}

		widget.Weather = weather
		return
	}

	if widget.Place == nil {
		place, err := fetchOpenMeteoPlaceFromName(widget.Location)
		if err != nil {
			widget.withError(err).scheduleEarlyUpdate()
			return
		}

		widget.Place = place
	}

	weather, err := fetchWeatherForOpenMeteoPlace(widget.Place, widget.Units)

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Weather = weather
}

func (widget *weatherWidget) Render() template.HTML {
	return widget.renderTemplate(widget, weatherWidgetTemplate)
}

type weather struct {
	Temperature         int
	ApparentTemperature int
	WeatherCode         int
	CloudCover          int
	CurrentColumn       int
	SunriseColumn       int
	SunsetColumn        int
	Columns             []weatherColumn
}

func (w *weather) WeatherCodeAsString() string {
	if weatherCode, ok := weatherCodeTable[w.WeatherCode]; ok {
		return weatherCode
	}

	return ""
}

func (w *weather) WeatherIcon() string {
	isDay := true
	if w.SunriseColumn != -1 && w.SunsetColumn != -1 {
		isDay = w.CurrentColumn >= w.SunriseColumn && w.CurrentColumn <= w.SunsetColumn
	}

	// Dynamic cloudiness check (threshold 80% for "full overcast")
	isOvercast := w.CloudCover >= 80

	switch w.WeatherCode {
	case 0, 1: // Clear / Mainly clear
		if isOvercast {
			return "overcast.svg"
		}
		if isDay {
			return "clear-day.svg"
		}
		return "clear-night.svg"
	case 2: // Partly cloudy
		if isOvercast {
			return "overcast.svg"
		}
		if isDay {
			return "partly-cloudy-day.svg"
		}
		return "partly-cloudy-night.svg"
	case 3: // Overcast
		return "overcast.svg"
	case 45, 48: // Fog
		if isOvercast {
			return "fog.svg"
		}
		if isDay {
			return "fog-day.svg"
		}
		return "fog-night.svg"
	case 51, 53, 55: // Drizzle
		if isOvercast {
			return "drizzle.svg"
		}
		if isDay {
			return "partly-cloudy-day-drizzle.svg"
		}
		return "partly-cloudy-night-drizzle.svg"
	case 56, 57, 66, 67: // Freezing drizzle/rain (Sleet)
		if isOvercast {
			return "sleet.svg"
		}
		if isDay {
			return "partly-cloudy-day-sleet.svg"
		}
		return "partly-cloudy-night-sleet.svg"
	case 61, 63, 65, 80, 81, 82: // Rain
		if isOvercast {
			return "rain.svg"
		}
		if isDay {
			return "partly-cloudy-day-rain.svg"
		}
		return "partly-cloudy-night-rain.svg"
	case 71, 73, 75, 77, 85, 86: // Snow
		if isOvercast {
			return "snow.svg"
		}
		if isDay {
			return "partly-cloudy-day-snow.svg"
		}
		return "partly-cloudy-night-snow.svg"
	case 95, 96, 99: // Thunderstorm
		if isOvercast {
			return "thunderstorms-rain.svg"
		}
		if isDay {
			return "thunderstorms-day.svg"
		}
		return "thunderstorms-night.svg"
	default:
		return "not-available.svg"
	}
}

type openMeteoPlacesResponseJson struct {
	Results []openMeteoPlaceResponseJson
}

type openMeteoPlaceResponseJson struct {
	Name      string
	Area      string `json:"admin1"`
	Latitude  float64
	Longitude float64
	Timezone  string
	Country   string
	location  *time.Location
}

type openMeteoWeatherResponseJson struct {
	Timezone string `json:"timezone"`

	Daily struct {
		Sunrise []int64 `json:"sunrise"`
		Sunset  []int64 `json:"sunset"`
	} `json:"daily"`

	Hourly struct {
		Temperature              []float64 `json:"temperature_2m"`
		PrecipitationProbability []int     `json:"precipitation_probability"`
	} `json:"hourly"`

	Current struct {
		Temperature         float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		WeatherCode         int     `json:"weather_code"`
		CloudCover          int     `json:"cloud_cover"`
	} `json:"current"`
}

type weatherColumn struct {
	Temperature      int
	Scale            float64
	HasPrecipitation bool
}

var commonCountryAbbreviations = map[string]string{
	"US":  "United States",
	"USA": "United States",
	"UK":  "United Kingdom",
}

func expandCountryAbbreviations(name string) string {
	if expanded, ok := commonCountryAbbreviations[strings.TrimSpace(name)]; ok {
		return expanded
	}

	return name
}

// Separates the location that Open Meteo accepts from the administrative area
// which can then be used to filter to the correct place after the list of places
// has been retrieved. Also expands abbreviations since Open Meteo does not accept
// country names like "US", "USA" and "UK"
func parsePlaceName(name string) (string, string) {
	parts := strings.Split(name, ",")

	if len(parts) == 1 {
		return name, ""
	}

	if len(parts) == 2 {
		return parts[0] + ", " + expandCountryAbbreviations(parts[1]), ""
	}

	return parts[0] + ", " + expandCountryAbbreviations(parts[2]), strings.TrimSpace(parts[1])
}

func fetchOpenMeteoPlaceFromName(location string) (*openMeteoPlaceResponseJson, error) {
	location, area := parsePlaceName(location)
	requestUrl := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=20&language=en&format=json", url.QueryEscape(location))
	request, _ := http.NewRequest("GET", requestUrl, nil)
	responseJson, err := decodeJsonFromRequest[openMeteoPlacesResponseJson](defaultHTTPClient, request)
	if err != nil {
		return nil, fmt.Errorf("fetching places data: %v", err)
	}

	if len(responseJson.Results) == 0 {
		return nil, fmt.Errorf("no places found for %s", location)
	}

	var place *openMeteoPlaceResponseJson

	if area != "" {
		area = strings.ToLower(area)

		for i := range responseJson.Results {
			if strings.ToLower(responseJson.Results[i].Area) == area {
				place = &responseJson.Results[i]
				break
			}
		}

		if place == nil {
			return nil, fmt.Errorf("no place found for %s in %s", location, area)
		}
	} else {
		place = &responseJson.Results[0]
	}

	loc, err := time.LoadLocation(place.Timezone)
	if err != nil {
		return nil, fmt.Errorf("loading location: %v", err)
	}

	place.location = loc

	return place, nil
}

func fetchWeatherForOpenMeteoPlace(place *openMeteoPlaceResponseJson, units string) (*weather, error) {
	query := url.Values{}
	var temperatureUnit string

	if units == "imperial" {
		temperatureUnit = "fahrenheit"
	} else {
		temperatureUnit = "celsius"
	}

	query.Add("latitude", fmt.Sprintf("%f", place.Latitude))
	query.Add("longitude", fmt.Sprintf("%f", place.Longitude))
	query.Add("timeformat", "unixtime")
	query.Add("timezone", place.Timezone)
	query.Add("forecast_days", "1")
	query.Add("current", "temperature_2m,apparent_temperature,weather_code,cloud_cover")
	query.Add("hourly", "temperature_2m,precipitation_probability")
	query.Add("daily", "sunrise,sunset")
	query.Add("temperature_unit", temperatureUnit)

	requestUrl := "https://api.open-meteo.com/v1/forecast?" + query.Encode()
	request, _ := http.NewRequest("GET", requestUrl, nil)
	responseJson, err := decodeJsonFromRequest[openMeteoWeatherResponseJson](defaultHTTPClient, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errNoContent, err)
	}

	now := time.Now().In(place.location)
	bars := make([]weatherColumn, 0, 24)
	currentBar := now.Hour() / 2
	sunriseBar := (time.Unix(int64(responseJson.Daily.Sunrise[0]), 0).In(place.location).Hour()) / 2
	sunsetBar := (time.Unix(int64(responseJson.Daily.Sunset[0]), 0).In(place.location).Hour() - 1) / 2

	if sunsetBar < 0 {
		sunsetBar = 0
	}

	if len(responseJson.Hourly.Temperature) == 24 {
		temperatures := make([]int, 12)
		precipitations := make([]bool, 12)

		t := responseJson.Hourly.Temperature
		p := responseJson.Hourly.PrecipitationProbability

		for i := 0; i < 24; i += 2 {
			if i/2 == currentBar {
				temperatures[i/2] = int(responseJson.Current.Temperature)
			} else {
				temperatures[i/2] = int(math.Round((t[i] + t[i+1]) / 2))
			}

			precipitations[i/2] = (p[i]+p[i+1])/2 > 75
		}

		minT := slices.Min(temperatures)
		maxT := slices.Max(temperatures)

		temperaturesRange := float64(maxT - minT)

		for i := 0; i < 12; i++ {
			bars = append(bars, weatherColumn{
				Temperature:      temperatures[i],
				HasPrecipitation: precipitations[i],
			})

			if temperaturesRange > 0 {
				bars[i].Scale = float64(temperatures[i]-minT) / temperaturesRange
			} else {
				bars[i].Scale = 1
			}
		}
	}

	return &weather{
		Temperature:         int(responseJson.Current.Temperature),
		ApparentTemperature: int(responseJson.Current.ApparentTemperature),
		WeatherCode:         responseJson.Current.WeatherCode,
		CloudCover:          responseJson.Current.CloudCover,
		CurrentColumn:       currentBar,
		SunriseColumn:       sunriseBar,
		SunsetColumn:        sunsetBar,
		Columns:             bars,
	}, nil
}

type brightSkyWeatherRecord struct {
	Timestamp                time.Time `json:"timestamp"`
	Temperature              float64   `json:"temperature"`
	WindSpeed                float64   `json:"wind_speed"`
	DewPoint                 float64   `json:"dew_point"`
	CloudCover               int       `json:"cloud_cover"`
	Precipitation            float64   `json:"precipitation"`
	Condition                string    `json:"condition"`
	PrecipitationProbability *int      `json:"precipitation_probability"`
}

type brightSkyWeatherResponseJson struct {
	Weather []brightSkyWeatherRecord `json:"weather"`
	Sources []struct {
		StationName string `json:"station_name"`
	} `json:"sources"`
}

// brightSkyCache stores weather records per location to preserve historical data
type brightSkyCache struct {
	mu      sync.RWMutex
	records map[string][]brightSkyWeatherRecord // key: "lat,lon"
	date    map[string]string                    // key: "lat,lon", value: date (YYYY-MM-DD)
}

var globalBrightSkyCache = &brightSkyCache{
	records: make(map[string][]brightSkyWeatherRecord),
	date:    make(map[string]string),
}

// getCacheKey generates a cache key from latitude and longitude
func getCacheKey(lat, lon float64) string {
	return fmt.Sprintf("%.4f,%.4f", lat, lon)
}

// addRecords adds new weather records to cache, merging with existing data for the same day
func (c *brightSkyCache) addRecords(key string, records []brightSkyWeatherRecord, currentDate string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if it's a new day - if so, clear old data
	if c.date[key] != currentDate {
		c.records[key] = nil
		c.date[key] = currentDate
	}

	// Merge new records with existing ones, avoiding duplicates
	existing := c.records[key]
	recordMap := make(map[time.Time]brightSkyWeatherRecord)

	// Add existing records to map
	for _, r := range existing {
		recordMap[r.Timestamp] = r
	}

	// Add or update with new records
	for _, r := range records {
		recordMap[r.Timestamp] = r
	}

	// Convert back to slice and sort by timestamp
	merged := make([]brightSkyWeatherRecord, 0, len(recordMap))
	for _, r := range recordMap {
		merged = append(merged, r)
	}

	// Sort by timestamp
	slices.SortFunc(merged, func(a, b brightSkyWeatherRecord) int {
		if a.Timestamp.Before(b.Timestamp) {
			return -1
		}
		if a.Timestamp.After(b.Timestamp) {
			return 1
		}
		return 0
	})

	c.records[key] = merged
}

// getRecords retrieves cached weather records for a location
func (c *brightSkyCache) getRecords(key string, currentDate string) []brightSkyWeatherRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if cache is for the current day
	if c.date[key] != currentDate {
		return nil
	}

	return c.records[key]
}

func fetchWeatherFromBrightSky(lat, lon float64, units string) (*weather, string, error) {
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	cacheKey := getCacheKey(lat, lon)
	
	// Fetch slightly more data to ensure we cover the full 24h of the current local day
	date := now.Add(-24 * time.Hour).Format("2006-01-02")
	lastDate := now.Add(24 * time.Hour).Format("2006-01-02")
	requestUrl := fmt.Sprintf("https://api.brightsky.dev/weather?lat=%f&lon=%f&date=%s&last_date=%s", lat, lon, date, lastDate)
	request, _ := http.NewRequest("GET", requestUrl, nil)
	responseJson, err := decodeJsonFromRequest[brightSkyWeatherResponseJson](defaultHTTPClient, request)
	if err != nil {
		return nil, "", fmt.Errorf("fetching Bright Sky data: %v", err)
	}

	if len(responseJson.Weather) == 0 {
		return nil, "", fmt.Errorf("no weather data returned from Bright Sky")
	}

	stationName := ""
	if len(responseJson.Sources) > 0 {
		stationName = responseJson.Sources[0].StationName
	}

	// Add fresh records to cache
	globalBrightSkyCache.addRecords(cacheKey, responseJson.Weather, currentDate)

	// Get all cached records (which now includes the fresh data)
	allRecords := globalBrightSkyCache.getRecords(cacheKey, currentDate)
	if allRecords == nil {
		allRecords = responseJson.Weather
	}

	// Filter and group records for the current local day (00:00 - 23:59)
	localDayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	localDayEnd := localDayStart.Add(24 * time.Hour)

	// Fetch sunrise/sunset from Open-Meteo as Bright Sky doesn't provide them easily
	sunriseBar, sunsetBar := -1, -1
	omUrl := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&daily=sunrise,sunset&timeformat=unixtime&timezone=auto&forecast_days=1", lat, lon)
	omRequest, _ := http.NewRequest("GET", omUrl, nil)
	omResponse, err := decodeJsonFromRequest[openMeteoWeatherResponseJson](defaultHTTPClient, omRequest)
	if err == nil && len(omResponse.Daily.Sunrise) > 0 {
		// We use .Local() to ensure consistency with how bars are grouped (which also uses server local time)
		sunriseBar = time.Unix(omResponse.Daily.Sunrise[0], 0).Local().Hour() / 2
		sunsetBar = (time.Unix(omResponse.Daily.Sunset[0], 0).Local().Hour() - 1) / 2

		if sunsetBar < 0 {
			sunsetBar = 0
		}
	}

	// Find record closest to now for current weather
	var currentRecord brightSkyWeatherRecord
	minDiff := math.MaxFloat64
	currentFound := false
	for _, r := range allRecords {
		diff := math.Abs(now.Sub(r.Timestamp).Seconds())
		if diff < minDiff {
			minDiff = diff
			currentRecord = r
			currentFound = true
		}
	}

	if !currentFound {
		return nil, "", fmt.Errorf("could not find current weather record in Bright Sky response")
	}

	temp := currentRecord.Temperature
	apparentTemp := CalculateApparentTemperature(currentRecord.Temperature, currentRecord.WindSpeed, currentRecord.DewPoint)
	if units == "imperial" {
		temp = temp*1.8 + 32
		apparentTemp = apparentTemp*1.8 + 32
	}

	// Build columns
	temperatures := make([]float64, 12)
	counts := make([]int, 12)
	precipitations := make([]int, 12)
	precipCounts := make([]int, 12)

	for _, r := range allRecords {
		// Only use records from the current local day
		if r.Timestamp.Before(localDayStart) || r.Timestamp.After(localDayEnd) {
			continue
		}

		hour := r.Timestamp.Local().Hour()
		col := hour / 2
		if col >= 0 && col < 12 {
			t := r.Temperature
			if units == "imperial" {
				t = t*1.8 + 32
			}
			temperatures[col] += t
			counts[col]++

			if r.PrecipitationProbability != nil {
				precipitations[col] += *r.PrecipitationProbability
				precipCounts[col]++
			} else {
				// Fallback to precipitation amount
				if r.Precipitation > 0 {
					precipitations[col] += 100
				}
				precipCounts[col]++
			}
		}
	}

	bars := make([]weatherColumn, 12)
	finalTemps := make([]int, 12)
	for i := 0; i < 12; i++ {
		if counts[i] > 0 {
			finalTemps[i] = int(math.Round(temperatures[i] / float64(counts[i])))
		} else {
			finalTemps[i] = int(math.Round(temp)) // Fallback
		}

		hasPrecip := false
		if precipCounts[i] > 0 {
			hasPrecip = (precipitations[i] / precipCounts[i]) > 75
		}
		bars[i].Temperature = finalTemps[i]
		bars[i].HasPrecipitation = hasPrecip
	}

	minT := slices.Min(finalTemps)
	maxT := slices.Max(finalTemps)
	temperaturesRange := float64(maxT - minT)

	for i := 0; i < 12; i++ {
		if temperaturesRange > 0 {
			bars[i].Scale = float64(bars[i].Temperature-minT) / temperaturesRange
		} else {
			bars[i].Scale = 1
		}
	}

	return &weather{
		Temperature:         int(math.Round(temp)),
		ApparentTemperature: int(math.Round(apparentTemp)),
		WeatherCode:         brightSkyConditionToWMO(currentRecord.Condition),
		CloudCover:          currentRecord.CloudCover,
		CurrentColumn:       now.Hour() / 2,
		Columns:             bars,
		SunriseColumn:       sunriseBar,
		SunsetColumn:        sunsetBar,
	}, stationName, nil
}

func brightSkyConditionToWMO(condition string) int {
	switch condition {
	case "dry", "clear-day", "clear-night":
		return 0
	case "partly-cloudy-day", "partly-cloudy-night":
		return 2
	case "cloudy":
		return 3
	case "fog":
		return 45
	case "rain":
		return 61
	case "sleet":
		return 66
	case "snow":
		return 71
	case "hail":
		return 77
	case "thunderstorm":
		return 95
	default:
		return 0
	}
}

func CalculateApparentTemperature(temp, windSpeed, dewPoint float64) float64 {
	// 1. Calculate Relative Humidity (RH) using Magnus-Tetens formula
	rh := 100 * math.Exp((17.625*dewPoint)/(243.04+dewPoint)) / math.Exp((17.625*temp)/(243.04+temp))

	if temp <= 10 {
		// 2. Wind Chill (standard Jagua/NWS, windSpeed in km/h)
		if windSpeed < 4.8 {
			return temp
		}
		v016 := math.Pow(windSpeed, 0.16)
		return 13.12 + 0.6215*temp - 11.37*v016 + 0.3965*temp*v016
	}

	if temp >= 26.7 {
		// 3. Heat Index (Rothfusz formula, requires Fahrenheit)
		tf := temp*1.8 + 32
		hi := -42.379 + 2.04901523*tf + 10.14333127*rh - 0.22475541*tf*rh - 0.00683783*tf*tf - 0.05481717*rh*rh + 0.00122874*tf*tf*rh + 0.00085282*tf*rh*rh - 0.00000199*tf*tf*rh*rh
		return (hi - 32) / 1.8
	}

	// 4. For temperatures between 10°C and 26.7°C, return raw temperature
	return temp
}

var weatherCodeTable = map[int]string{
	0:  "Bezchmurnie",
	1:  "Głównie bezchmurnie",
	2:  "Częściowe zachmurzenie",
	3:  "Pochmurno",
	45: "Mgła",
	48: "Mgła szronowa",
	51: "Mżawka lekka",
	53: "Mżawka",
	55: "Mżawka gęsta",
	56: "Mżawka marznąca lekka",
	57: "Mżawka marznąca gęsta",
	61: "Deszcz",
	63: "Umiarkowany deszcz",
	65: "Intensywny deszcz",
	66: "Deszcz marznący lekki",
	67: "Deszcz marznący intensywny",
	71: "Śnieg",
	73: "Umiarkowany śnieg",
	75: "Intensywny śnieg",
	77: "Ziarna śniegu",
	80: "Deszcz",
	81: "Umiarkowany deszcz",
	82: "Intensywny deszcz",
	85: "Śnieg",
	86: "Śnieg",
	95: "Burza",
	96: "Burza",
	99: "Burza",
}
