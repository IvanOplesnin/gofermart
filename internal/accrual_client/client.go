package accrualclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
	"github.com/IvanOplesnin/gofermart.git/internal/service/gophermart"
)

type Client struct {
	baseUrl string
}

func New(baseUrl string) *Client {
	return &Client{
		baseUrl: baseUrl,
	}
}

func (c *Client) GetOrder(ctx context.Context, number string) (response *gophermart.AccrualResponse, err error) {
	uri, err := url.JoinPath(c.baseUrl, "/api/orders", number)
	if err != nil {
		return nil, fmt.Errorf("acrualClient.GetOrder: %s", err)
	}
	mapStatus := map[string]string{
		"REGISTERED": "NEW",
		"INVALID":    "INVALID",
		"PROCESSING": "PROCESSING",
		"PROCESSED":  "PROCESSED",
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("acrualClient.GetOrder: %w", err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("acrualClient.GetOrder: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, gophermart.ErrToManyRequests
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		logger.Log.Warnf("accrualClient.GetOrder: %s", resp.Request.URL)
		return nil, fmt.Errorf("status code: %v", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNoContent {
		return &gophermart.AccrualResponse{
			Status: "NEW",
			OrderNumber: number,
		}, nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("accrualClient.GetOrder: %w", err)
	}
	var accrualResponse gophermart.AccrualResponse
	if err := json.Unmarshal(raw, &accrualResponse); err != nil {
		return nil, fmt.Errorf("accrualClient.GetOrder: %w", err)
	}
	if status, ok := mapStatus[accrualResponse.Status]; ok {
		accrualResponse.Status = status
		return &accrualResponse, nil
	} else {
		return nil, fmt.Errorf("accrualClient.GetOrder: status not found: %s", accrualResponse.Status)
	}
}
