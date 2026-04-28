package scanner

import (
	"net/http"
	"testing"
	"time"
)

func TestFetchPoolData(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	data, err := FetchPoolData(client)
	if err != nil {
		t.Fatalf("Failed to fetch pool data: %v", err)
	}

	if data.TotalBlocks <= 0 {
		t.Errorf("Expected TotalBlocks > 0, got %d", data.TotalBlocks)
	}

	if data.LastBlockTime.IsZero() {
		t.Error("Expected LastBlockTime to be non-zero")
	}

	if data.EpochBlocks < 0 {
		t.Errorf("Expected EpochBlocks >= 0, got %d", data.EpochBlocks)
	}

	t.Logf("Fetched data: TotalBlocks=%d, EpochBlocks=%d, LastBlockTime=%v", data.TotalBlocks, data.EpochBlocks, data.LastBlockTime)
}
