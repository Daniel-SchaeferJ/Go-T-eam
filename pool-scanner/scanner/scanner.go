package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	PoolIDHex    = "c473846af5dfa3d56d2bdb765d8a2835be6dcc5c70be5cc4ff136a45"
	PoolIDBech32 = "pool1c3ecg6h4m73a2mftmdm9mz3gxklxmnzuwzl9e38lzd4y26qcm8a"
	KoiosBaseURL = "https://api.koios.rest/api/v1"
)

type PoolInfo struct {
	BlockCount int64 `json:"block_count"`
}

type Tip struct {
	EpochNo int64 `json:"epoch_no"`
}

type Block struct {
	EpochNo     int64 `json:"epoch_no"`
	BlockHeight int64 `json:"block_height"`
	BlockTime   int64 `json:"block_time"`
}

type PoolData struct {
	TotalBlocks   int64
	LastBlockTime time.Time
	EpochBlocks   int64
}

func FetchPoolData(client *http.Client) (PoolData, error) {
	// 1. Get Tip (Current Epoch)
	tipResp, err := client.Get(KoiosBaseURL + "/tip")
	if err != nil {
		return PoolData{}, fmt.Errorf("tip fetch error: %w", err)
	}
	defer tipResp.Body.Close()
	var tips []Tip
	if err := json.NewDecoder(tipResp.Body).Decode(&tips); err != nil || len(tips) == 0 {
		return PoolData{}, fmt.Errorf("tip decode error: %w", err)
	}
	currentEpoch := tips[0].EpochNo

	// 2. Fetch Pool Info (Total Blocks)
	infoUrl := KoiosBaseURL + "/pool_info"
	infoReqBody, _ := json.Marshal(map[string][]string{"_pool_bech32_ids": {PoolIDBech32}})
	infoResp, err := client.Post(infoUrl, "application/json", bytes.NewBuffer(infoReqBody))
	if err != nil {
		return PoolData{}, err
	}
	defer infoResp.Body.Close()

	var infos []PoolInfo
	if err := json.NewDecoder(infoResp.Body).Decode(&infos); err != nil {
		return PoolData{}, err
	}
	if len(infos) == 0 {
		return PoolData{}, fmt.Errorf("pool not found")
	}

	// 3. Fetch Blocks in current epoch
	epochBlocksUrl := fmt.Sprintf("%s/pool_blocks?_pool_bech32=%s&epoch_no=eq.%d", KoiosBaseURL, PoolIDBech32, currentEpoch)
	epochBlocksResp, err := client.Get(epochBlocksUrl)
	if err != nil {
		return PoolData{}, err
	}
	defer epochBlocksResp.Body.Close()

	var epochBlocks []Block
	if err := json.NewDecoder(epochBlocksResp.Body).Decode(&epochBlocks); err != nil {
		return PoolData{}, err
	}

	// 4. Fetch Last Block (might be from previous epoch if none in current, but we'll use the most recent one overall)
	lastBlockUrl := fmt.Sprintf("%s/pool_blocks?_pool_bech32=%s&order=block_height.desc&limit=1", KoiosBaseURL, PoolIDBech32)
	lastBlockResp, err := client.Get(lastBlockUrl)
	if err != nil {
		return PoolData{}, err
	}
	defer lastBlockResp.Body.Close()

	var lastBlocks []Block
	if err := json.NewDecoder(lastBlockResp.Body).Decode(&lastBlocks); err != nil {
		return PoolData{}, err
	}

	lastBlockTime := time.Time{}
	if len(lastBlocks) > 0 {
		lastBlockTime = time.Unix(lastBlocks[0].BlockTime, 0)
	}

	return PoolData{
		TotalBlocks:   infos[0].BlockCount,
		LastBlockTime: lastBlockTime,
		EpochBlocks:   int64(len(epochBlocks)),
	}, nil
}
