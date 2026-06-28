package main

import (
	"testing"
)

func TestUnmarshalPositionReport(t *testing.T) {
	rawJSON := []byte(`{
		"PositionReport": {
			"Cog": 123.4,
			"Latitude": 32.7,
			"Longitude": -117.1,
			"NavigationalStatus": 1,
			"TrueHeading": 180
		}
	}`)

	var pos PositionReport
	if err := unmarshalMessage(rawJSON, "PositionReport", &pos); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if pos.Cog != 123.4 {
		t.Errorf("Cog = %v, want 123.4", pos.Cog)
	}
	if pos.TrueHeading != 180 {
		t.Errorf("TrueHeading = %v, want 180", pos.TrueHeading)
	}
}

func TestUnmarshalShipStaticData(t *testing.T) {
	rawJSON := []byte(`{
		"ShipStaticData": {
			"Name": "TEST SHIP",
			"ImoNumber": 1234567,
			"CallSign": "WDC1234",
			"Type": 70
		}
	}`)

	var sd ShipStaticData
	if err := unmarshalMessage(rawJSON, "ShipStaticData", &sd); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if sd.Name != "TEST SHIP" {
		t.Errorf("Name = %q, want TEST SHIP", sd.Name)
	}
	if sd.ImoNumber != 1234567 {
		t.Errorf("ImoNumber = %v, want 1234567", sd.ImoNumber)
	}
	if sd.Type != 70 {
		t.Errorf("Type = %v, want 70", sd.Type)
	}
}
