package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
)

type Withdrawer interface {
	Withdraw(ctx context.Context, orderNumber string, summa float64) error
	ListWithdraws(ctx context.Context) ([]Withdraw, error)
}

type RequestWithdraw struct {
	OrderNumber string  `json:"order"`
	Summa       float64 `json:"sum"`
}

type Withdraw struct {
	OrderNumber string      `json:"order"`
	Summa       float64     `json:"sum"`
	ProcessedAt RFC3339Time `json:"processed_at"`
}

var ErrInvalidOrderNumber = errors.New("invalid order number")
var ErrEnoughMoney = errors.New("not enough money")
var ErrInvalidSumma = errors.New("invalid summa")

func WithdrawHandler(wd Withdrawer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get(contentTypeKey)
		if ct == "" || !strings.HasPrefix(ct, applicationJSONValue) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var withdrawData RequestWithdraw
		if err := json.NewDecoder(r.Body).Decode(&withdrawData); err != nil {
			logger.Log.Errorf("WithdrawHandler decode error: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := wd.Withdraw(r.Context(), withdrawData.OrderNumber, withdrawData.Summa)

		switch {
		case err == nil:
			w.WriteHeader(http.StatusOK)
			return

		case errors.Is(err, ErrEnoughMoney):
			logger.Log.Infof("WithdrawHandler not enough money: order=%s", withdrawData.OrderNumber)
			w.WriteHeader(http.StatusPaymentRequired)
			return

		case errors.Is(err, ErrInvalidOrderNumber):
			logger.Log.Infof("WithdrawHandler invalid order_number: order=%s", withdrawData.OrderNumber)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return

		case errors.Is(err, ErrInvalidSumma):
			logger.Log.Infof("WithdrawHandler invalid summa: order=%s sum=%v", withdrawData.OrderNumber, withdrawData.Summa)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return

		default:
			logger.Log.Errorf("WithdrawHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func ListWithdrawHandler(wd Withdrawer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		withdraws, err := wd.ListWithdraws(ctx)
		if err != nil {
			logger.Log.Errorf("ListWithdrawHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(withdraws) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(withdraws); err != nil {
			logger.Log.Errorf("ListWithdrawHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
