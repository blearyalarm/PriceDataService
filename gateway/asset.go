package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/erich/pricetracking/config"
	"github.com/erich/pricetracking/model"
)

// for json marshalling
type TempResult struct {
	Entries []TempEntry `json:"result"`
}

type TempEntry struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

const DATE_FORMAT = "2006-01-02T15:04:05"

type assetClient struct {
	cfg    *config.Config
	client *http.Client
}

type AssetClient interface {
	Load(ctx context.Context, startTime time.Time, endTime time.Time) ([]model.Entry, error)
}

func NewAssetClient(cfg *config.Config) (AssetClient, error) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	return &assetClient{
		cfg:    cfg,
		client: client,
	}, nil
}

func (ac *assetClient) Load(ctx context.Context, startTime time.Time, endTime time.Time) ([]model.Entry, error) {
	result := TempResult{}

	params := url.Values{
		"start": {startTime.Format(DATE_FORMAT)},
		"end":   {endTime.Format(DATE_FORMAT)},
	}
	reqUrl := ac.cfg.AssetClient.ServerAddr + "?" + params.Encode()
	resp, err := ac.client.Get(reqUrl)
	if err != nil {
		return []model.Entry{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []model.Entry{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []model.Entry{}, err
	}

	if err = json.Unmarshal(data, &result); err != nil {
		return []model.Entry{}, err
	}

	modelResult := make([]model.Entry, len(result.Entries))
	for i, v := range result.Entries {
		modelResult[i] = model.Entry{
			Time:  time.Unix(int64(v.Time), 0).UTC(),
			Value: v.Value,
		}
	}

	return modelResult, nil
}
