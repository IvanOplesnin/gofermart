package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
)

type Balancer interface {
	Balance(ctx context.Context) (BalanceResponse, error)
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func BalanceHandler(b Balancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		resp, err := b.Balance(ctx)
		if err != nil {
			logger.Log.Errorf("BalanceHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Log.Errorf("BalanceHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
