package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestToRecords(t *testing.T) {
	ships := []Ship{
		{
			MMSI:      123456789,
			IMO:       9876543,
			Name:      "TEST VESSEL",
			CallSign:  "ABCD",
			Type:      70,
			Heading:   180,
			Course:    179.5,
			Speed:     12.3,
			Longitude: -122.4194,
			Latitude:  37.7749,
			Status:    0,
			Timestamp: "2026-06-25 04:00:00",
			Draught:   8.5,
			Dest:      "SAN FRANCISCO",
			ETA:       "06-25 12:00",
		},
	}

	now := time.Date(2026, 6, 25, 4, 30, 0, 0, time.UTC)
	records := toRecords(ships, now)

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	r := records[0]
	if r.MMSI != 123456789 {
		t.Errorf("MMSI = %d, want 123456789", r.MMSI)
	}
	if r.Latitude != 37.7749 {
		t.Errorf("Latitude = %f, want 37.7749", r.Latitude)
	}
	if r.Longitude != -122.4194 {
		t.Errorf("Longitude = %f, want -122.4194", r.Longitude)
	}
	if r.ScrapeTime != "2026-06-25T04:30:00Z" {
		t.Errorf("ScrapeTime = %s, want 2026-06-25T04:30:00Z", r.ScrapeTime)
	}
	if r.Name != "TEST VESSEL" {
		t.Errorf("Name = %s, want TEST VESSEL", r.Name)
	}
}

func TestFetchShips(t *testing.T) {
	metadata := map[string]any{"ERROR": false, "USERNAME": "test", "FORMAT": 1}
	ships := []Ship{
		{MMSI: 111111111, Name: "SHIP ONE", Longitude: 10.0, Latitude: 20.0},
		{MMSI: 222222222, Name: "SHIP TWO", Longitude: 30.0, Latitude: 40.0},
	}

	metaJSON, _ := json.Marshal(metadata)
	shipsJSON, _ := json.Marshal(ships)
	response := "[" + string(metaJSON) + "," + string(shipsJSON) + "]"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer srv.Close()

	// Override the fetch to use our test server.
	result, err := fetchShipsFromURL(srv.URL)
	if err != nil {
		t.Fatalf("fetchShipsFromURL failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 ships, got %d", len(result))
	}
	if result[0].Name != "SHIP ONE" {
		t.Errorf("first ship name = %s, want SHIP ONE", result[0].Name)
	}
}
