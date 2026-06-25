# Ship Aggregator

A Go program that scrapes AIS ship data from the [AISHub API](https://www.aishub.net/api) and stores it in a [Hugging Face](https://huggingface.co/) dataset, tracking ship locations over time.

## Features
- **AIS Data Scraping**: Fetches live vessel positions, headings, speeds, and metadata from AISHub.
- **Temporal Tracking**: Each scrape is timestamped so ship movements can be analyzed over time.
- **Hugging Face Storage**: Uploads JSON snapshots to a Hugging Face dataset repository, organized by date.
- **Auto-polling**: Scrapes every 10 minutes by default.

## Requirements
- Go 1.25 or later
- An [AISHub](https://www.aishub.net/) API key
- A [Hugging Face](https://huggingface.co/) write token

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `AISHUB_API_KEY` | Yes | Your AISHub API username/key |
| `HF_TOKEN` | Yes | Hugging Face write access token |
| `HF_REPO` | No | HF dataset repo name (default: `ship-tracking-data`) |

## How to Run

```bash
export AISHUB_API_KEY="your-aishub-key"
export HF_TOKEN="your-hf-token"
export HF_REPO="your-username/ship-tracking-data"
go run .
```

## Data Format

Each scrape produces a JSON file stored at `data/YYYY-MM-DD/HHMMSS.json` in the Hugging Face dataset. Each file contains an array of ship records:

```json
[
  {
    "scrape_time": "2026-06-25T04:30:00Z",
    "mmsi": 123456789,
    "imo": 9876543,
    "name": "EXAMPLE VESSEL",
    "callsign": "ABCD",
    "ship_type": 70,
    "heading": 180,
    "course": 179.5,
    "speed": 12.3,
    "longitude": -122.4194,
    "latitude": 37.7749,
    "nav_status": 0,
    "ais_time": "2026-06-25 04:00:00",
    "draught": 8.5,
    "destination": "SAN FRANCISCO",
    "eta": "06-25 12:00"
  }
]
```

## Testing

```bash
go test ./...
```
