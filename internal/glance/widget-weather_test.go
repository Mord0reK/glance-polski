package glance

import (
	"testing"
	"time"
)

func TestBrightSkyCacheAddAndGetRecords(t *testing.T) {
	cache := &brightSkyCache{
		records: make(map[string][]brightSkyWeatherRecord),
		date:    make(map[string]string),
	}

	key := getCacheKey(52.5200, 13.4050) // Berlin coordinates
	currentDate := "2026-01-09"

	// Create test records
	records := []brightSkyWeatherRecord{
		{
			Timestamp:   time.Date(2026, 1, 9, 10, 0, 0, 0, time.UTC),
			Temperature: 5.0,
			CloudCover:  50,
		},
		{
			Timestamp:   time.Date(2026, 1, 9, 12, 0, 0, 0, time.UTC),
			Temperature: 7.0,
			CloudCover:  30,
		},
	}

	// Add records to cache
	cache.addRecords(key, records, currentDate)

	// Get records from cache
	retrieved := cache.getRecords(key, currentDate)

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 records, got %d", len(retrieved))
	}

	// Verify records are sorted by timestamp
	if retrieved[0].Timestamp.After(retrieved[1].Timestamp) {
		t.Error("Records are not sorted by timestamp")
	}
}

func TestBrightSkyCacheMergesRecords(t *testing.T) {
	cache := &brightSkyCache{
		records: make(map[string][]brightSkyWeatherRecord),
		date:    make(map[string]string),
	}

	key := getCacheKey(52.5200, 13.4050)
	currentDate := "2026-01-09"

	// Add first batch of records
	records1 := []brightSkyWeatherRecord{
		{
			Timestamp:   time.Date(2026, 1, 9, 10, 0, 0, 0, time.UTC),
			Temperature: 5.0,
		},
	}
	cache.addRecords(key, records1, currentDate)

	// Add second batch with overlapping and new records
	records2 := []brightSkyWeatherRecord{
		{
			Timestamp:   time.Date(2026, 1, 9, 10, 0, 0, 0, time.UTC),
			Temperature: 5.5, // Updated temperature
		},
		{
			Timestamp:   time.Date(2026, 1, 9, 12, 0, 0, 0, time.UTC),
			Temperature: 7.0, // New record
		},
	}
	cache.addRecords(key, records2, currentDate)

	retrieved := cache.getRecords(key, currentDate)

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 unique records after merge, got %d", len(retrieved))
	}

	// Verify the updated temperature is used
	for _, r := range retrieved {
		if r.Timestamp.Hour() == 10 && r.Temperature != 5.5 {
			t.Errorf("Expected updated temperature 5.5, got %f", r.Temperature)
		}
	}
}

func TestBrightSkyCacheClearsOnNewDay(t *testing.T) {
	cache := &brightSkyCache{
		records: make(map[string][]brightSkyWeatherRecord),
		date:    make(map[string]string),
	}

	key := getCacheKey(52.5200, 13.4050)
	date1 := "2026-01-09"
	date2 := "2026-01-10"

	// Add records for day 1
	records1 := []brightSkyWeatherRecord{
		{
			Timestamp:   time.Date(2026, 1, 9, 10, 0, 0, 0, time.UTC),
			Temperature: 5.0,
		},
	}
	cache.addRecords(key, records1, date1)

	// Verify records exist for day 1
	if len(cache.getRecords(key, date1)) != 1 {
		t.Error("Expected 1 record for day 1")
	}

	// Add records for day 2
	records2 := []brightSkyWeatherRecord{
		{
			Timestamp:   time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC),
			Temperature: 6.0,
		},
	}
	cache.addRecords(key, records2, date2)

	// Verify day 1 records are cleared
	if cache.getRecords(key, date1) != nil {
		t.Error("Expected nil when requesting old date")
	}

	// Verify only day 2 records exist
	retrieved := cache.getRecords(key, date2)
	if len(retrieved) != 1 {
		t.Errorf("Expected 1 record for day 2, got %d", len(retrieved))
	}
	if retrieved[0].Temperature != 6.0 {
		t.Errorf("Expected temperature 6.0, got %f", retrieved[0].Temperature)
	}
}

func TestGetCacheKey(t *testing.T) {
	key1 := getCacheKey(52.5200, 13.4050)
	key2 := getCacheKey(52.5200, 13.4050)
	key3 := getCacheKey(52.5201, 13.4050)

	if key1 != key2 {
		t.Error("Same coordinates should produce same key")
	}

	if key1 == key3 {
		t.Error("Different coordinates should produce different keys")
	}
}
