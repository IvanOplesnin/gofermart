package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

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
		if r.Header.Get(contentTypeKey) != applicationJSONValue {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var withdrawData RequestWithdraw
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Errorf("WithdrawHandler error: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(raw, &withdrawData); err != nil {
			logger.Log.Errorf("WithdrawHandler error: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		err = wd.Withdraw(ctx, withdrawData.OrderNumber, withdrawData.Summa)
		if errors.Is(err, ErrInvalidOrderNumber) {
			logger.Log.Infof("WithdrawHandler invalid oreder_number: %s", withdrawData.OrderNumber)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, ErrEnoughMoney) {
			logger.Log.Infof("WithdrawHandler not enough money: %s", withdrawData.OrderNumber)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, ErrInvalidSumma) {
			logger.Log.Infof("WithdrawHandler not enough money: %s", withdrawData.OrderNumber)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err != nil {
			logger.Log.Errorf("WithdrawHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
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
