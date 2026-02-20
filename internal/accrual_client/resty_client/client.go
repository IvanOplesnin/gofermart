package accrualclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
	"github.com/IvanOplesnin/gofermart.git/internal/service/gophermart"
)

type Client struct {
	r *resty.Client
}

func New(baseURL string, timeOut *time.Duration) *Client {
	t := 20 * time.Second
	if timeOut == nil {
		timeOut = &t
	}

	r := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(*timeOut)

	return &Client{r: r}
}

func (c *Client) GetOrder(ctx context.Context, number string) (*gophermart.AccrualResponse, error) {
	const op = "accrualClient.GetOrder"

	mapStatus := map[string]string{
		"REGISTERED": "NEW",
		"INVALID":    "INVALID",
		"PROCESSING": "PROCESSING",
		"PROCESSED":  "PROCESSED",
	}

	var dto gophermart.AccrualResponse

	resp, err := c.r.R().
		SetContext(ctx).
		SetPathParam("number", number).
		SetResult(&dto).
		Get("/api/orders/{number}")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	logger.Log.Debugf("%s: resp.StatusCode: %d", op, resp.StatusCode())

	switch resp.StatusCode() {
	case http.StatusTooManyRequests:
		if ra := resp.Header().Get("Retry-After"); ra != "" {
			logger.Log.Warnf("%s: 429 Retry-After=%s", op, ra)
		}
		return nil, gophermart.ErrToManyRequests

	case http.StatusNoContent:
		return &gophermart.AccrualResponse{
			Status:      "NEW",
			OrderNumber: number,
		}, nil

	case http.StatusOK:
		status, ok := mapStatus[dto.Status]
		if !ok {
			return nil, fmt.Errorf("%s: status not found: %s", op, dto.Status)
		}
		dto.Status = status
		logger.Log.Debugf("%s: raw string: %s", op, resp.String())
		logger.Log.Debugf("%s: accrualResponse: %v", op, dto)
		return &dto, nil

	default:
		return nil, fmt.Errorf("%s: status code: %d", op, resp.StatusCode())
	}
}
