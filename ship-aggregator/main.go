package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
	aisStreamURL   = "wss://stream.aisstream.io/v0/stream"
	hfAPIBase      = "https://huggingface.co/api"
	uploadInterval = 1 * time.Minute
)

// AISStreamMessage represents a message received from aisstream.io.
type AISStreamMessage struct {
	MessageType string          `json:"MessageType"`
	MetaData    AISStreamMeta   `json:"MetaData"`
	Message     json.RawMessage `json:"Message"`
}

// AISStreamMeta contains metadata about the AIS message.
type AISStreamMeta struct {
	MMSI       int         `json:"MMSI"`
	MMSIString json.Number `json:"MMSI_String"`
	ShipName   string      `json:"ShipName"`
	Latitude   float64     `json:"latitude"`
	Longitude  float64     `json:"longitude"`
	TimeUtc    string      `json:"time_utc"`
}

// PositionReport represents a decoded AIS position report from aisstream.io.
type PositionReport struct {
	Cog                       float64 `json:"Cog"`
	CommunicationState        int     `json:"CommunicationState"`
	Latitude                  float64 `json:"Latitude"`
	Longitude                 float64 `json:"Longitude"`
	MessageID                 int     `json:"MessageID"`
	NavigationalStatus        int     `json:"NavigationalStatus"`
	PositionAccuracy          bool    `json:"PositionAccuracy"`
	Raim                      bool    `json:"Raim"`
	RateOfTurn                float64 `json:"RateOfTurn"`
	RepeatIndicator           int     `json:"RepeatIndicator"`
	Sog                       float64 `json:"Sog"`
	Spare                     int     `json:"Spare"`
	SpecialManoeuvreIndicator int     `json:"SpecialManoeuvreIndicator"`
	Timestamp                 int     `json:"Timestamp"`
	TrueHeading               int     `json:"TrueHeading"`
	UserID                    int     `json:"UserID"`
	Valid                     bool    `json:"Valid"`
}

// ShipStaticData represents decoded AIS static/voyage data from aisstream.io.
type ShipStaticData struct {
	AisVersion           int           `json:"AisVersion"`
	CallSign             string        `json:"CallSign"`
	Destination          string        `json:"Destination"`
	Dimension            ShipDimension `json:"Dimension"`
	Draught              float64       `json:"Draught"`
	Dte                  bool          `json:"Dte"`
	Eta                  EtaData       `json:"Eta"`
	ImoNumber            int           `json:"ImoNumber"`
	MaximumStaticDraught float64       `json:"MaximumStaticDraught"`
	MessageID            int           `json:"MessageID"`
	Name                 string        `json:"Name"`
	RepeatIndicator      int           `json:"RepeatIndicator"`
	Spare                bool          `json:"Spare"`
	Type                 int           `json:"Type"`
	UserID               int           `json:"UserID"`
	Valid                bool          `json:"Valid"`
}

// ShipDimension holds ship dimension data.
type ShipDimension struct {
	A int `json:"A"`
	B int `json:"B"`
	C int `json:"C"`
	D int `json:"D"`
}

// EtaData holds ETA fields.
type EtaData struct {
	Day    int `json:"Day"`
	Hour   int `json:"Hour"`
	Minute int `json:"Minute"`
	Month  int `json:"Month"`
}

// TrackingRecord is what we store: ship data plus our own receive timestamp.
type TrackingRecord struct {
	ReceiveTime string  `json:"receive_time"`
	MMSI        int     `json:"mmsi"`
	IMO         int     `json:"imo,omitzero"`
	Name        string  `json:"name"`
	CallSign    string  `json:"callsign,omitzero"`
	ShipType    int     `json:"ship_type,omitzero"`
	Heading     int     `json:"heading"`
	Course      float64 `json:"course"`
	Speed       float64 `json:"speed"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	NavStatus   int     `json:"nav_status"`
	AISTime     string  `json:"ais_time"`
	Draught     float64 `json:"draught,omitzero"`
	Dest        string  `json:"destination,omitzero"`
	ETA         string  `json:"eta,omitzero"`
}

func fetchEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}

func ensureHFRepo(token, repoID string) error {
	url := fmt.Sprintf("%s/repos/create", hfAPIBase)
	parts := strings.SplitN(repoID, "/", 2)
	payload := map[string]any{
		"type":    "dataset",
		"name":    parts[len(parts)-1],
		"private": false,
	}
	if len(parts) == 2 {
		payload["organization"] = parts[0]
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("creating HF repo: %w", err)
	}
	defer resp.Body.Close()

	// 409 means repo already exists, which is fine.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HF repo create returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func uploadToHF(token, repoID string, records []TrackingRecord, scrapeTime time.Time) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling records: %w", err)
	}

	filename := fmt.Sprintf("data/%s.json", scrapeTime.UTC().Format("2006-01-02/150405"))
	apiURL := fmt.Sprintf("%s/datasets/%s/commit/main", hfAPIBase, repoID)

	// HF commit API uses NDJSON format.
	headerLine, _ := json.Marshal(map[string]any{
		"key": "header",
		"value": map[string]any{
			"summary": fmt.Sprintf("Add data %s", scrapeTime.UTC().Format("2006-01-02 15:04:05")),
			"operations": []map[string]string{
				{"key": "file", "path": filename},
			},
		},
	})
	fileLine, _ := json.Marshal(map[string]any{
		"key": "file",
		"value": map[string]string{
			"content":  base64.StdEncoding.EncodeToString(data),
			"encoding": "base64",
			"path":     filename,
		},
	})

	var body bytes.Buffer
	body.Write(headerLine)
	body.WriteByte('\n')
	body.Write(fileLine)
	body.WriteByte('\n')

	req, err := http.NewRequest(http.MethodPost, apiURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("uploading to HF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HF upload returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func connectAISStream(apiKey string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(aisStreamURL, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to aisstream.io: %w", err)
	}

	// Subscribe to all position reports and static data worldwide.
	subscribeMsg := map[string]any{
		"APIKey": apiKey,
		"BoundingBoxes": [][]any{
			{[]float64{-90, -180}, []float64{90, 180}},
		},
		"FilterMessageTypes": []string{"PositionReport", "ShipStaticData"},
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("sending subscription: %w", err)
	}

	return conn, nil
}

func run() error {
	godotenv.Load()
	aisStreamKey := fetchEnv("AISSTREAM_API_KEY")
	hfToken := fetchEnv("HF_TOKEN")
	hfRepo := os.Getenv("HF_REPO")

	log.Printf("Starting ship aggregator (upload every %s)", uploadInterval)
	log.Printf("HF dataset: %s", hfRepo)

	if err := ensureHFRepo(hfToken, hfRepo); err != nil {
		log.Printf("Warning: could not ensure HF repo exists: %v", err)
	}

	var (
		mu      sync.Mutex
		records []TrackingRecord
	)

	// Upload buffered records periodically.
	go func() {
		ticker := time.NewTicker(uploadInterval)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			if len(records) == 0 {
				mu.Unlock()
				continue
			}
			batch := records
			records = nil
			mu.Unlock()

			now := time.Now()
			log.Printf("Uploading %d records to HF dataset %s", len(batch), hfRepo)
			if err := uploadToHF(hfToken, hfRepo, batch, now); err != nil {
				log.Printf("Error uploading to Hugging Face: %v", err)
				// Put records back so they aren't lost.
				mu.Lock()
				records = append(batch, records...)
				mu.Unlock()
			} else {
				log.Printf("Uploaded %d records to HF dataset %s", len(batch), hfRepo)
			}
		}
	}()

	// Connect and stream with automatic reconnection.
	for {
		log.Println("Connecting to aisstream.io...")
		conn, err := connectAISStream(aisStreamKey)
		if err != nil {
			log.Printf("Error connecting: %v, retrying in 10s...", err)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Println("Connected to aisstream.io, receiving AIS data...")

		for {
			var msg AISStreamMessage
			if err := conn.ReadJSON(&msg); err != nil {
				log.Printf("Error reading from stream: %v, reconnecting...", err)
				conn.Close()
				break
			}

			now := time.Now().UTC().Format(time.RFC3339)

			switch msg.MessageType {
			case "PositionReport":
				var pos PositionReport
				if err := json.Unmarshal(msg.Message, &pos); err != nil {
					log.Printf("Error parsing position report: %v", err)
					continue
				}
				rec := TrackingRecord{
					ReceiveTime: now,
					MMSI:        msg.MetaData.MMSI,
					Name:        msg.MetaData.ShipName,
					Heading:     pos.TrueHeading,
					Course:      pos.Cog,
					Speed:       pos.Sog,
					Longitude:   msg.MetaData.Longitude,
					Latitude:    msg.MetaData.Latitude,
					NavStatus:   pos.NavigationalStatus,
					AISTime:     msg.MetaData.TimeUtc,
				}
				mu.Lock()
				records = append(records, rec)
				mu.Unlock()

			case "ShipStaticData":
				var sd ShipStaticData
				if err := json.Unmarshal(msg.Message, &sd); err != nil {
					log.Printf("Error parsing static data: %v", err)
					continue
				}
				eta := ""
				if sd.Eta.Month > 0 {
					eta = fmt.Sprintf("%02d-%02dT%02d:%02d", sd.Eta.Month, sd.Eta.Day, sd.Eta.Hour, sd.Eta.Minute)
				}
				rec := TrackingRecord{
					ReceiveTime: now,
					MMSI:        msg.MetaData.MMSI,
					IMO:         sd.ImoNumber,
					Name:        sd.Name,
					CallSign:    sd.CallSign,
					ShipType:    sd.Type,
					Longitude:   msg.MetaData.Longitude,
					Latitude:    msg.MetaData.Latitude,
					AISTime:     msg.MetaData.TimeUtc,
					Draught:     sd.MaximumStaticDraught,
					Dest:        sd.Destination,
					ETA:         eta,
				}
				mu.Lock()
				records = append(records, rec)
				mu.Unlock()
			}
		}

	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
