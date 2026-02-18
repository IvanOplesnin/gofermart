package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
)

type Ordered interface {
	AddOrder(ctx context.Context, orderID string) (exist bool, err error)
	Orders(ctx context.Context) ([]Order, error)
}

type Order struct {
	Number     string      `json:"number"`
	Status     string      `json:"status"`
	Accrual    *float64    `json:"accrual"`
	UploadedAt RFC3339Time `json:"uploaded_at"`
}

var ErrInvalidOrderID = errors.New("invalid order id")
var ErrAnotherUserOrder = errors.New("another user order")

func AddOrderHandler(o Ordered) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(contentTypeKey) != textPlainValue {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		number, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		exist, err := o.AddOrder(ctx, string(number))
		if errors.Is(err, ErrInvalidOrderID) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, ErrAnotherUserOrder) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exist {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func OrdersHandler(o Ordered) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orders, err := o.Orders(ctx)
		if err != nil {
			logger.Log.Errorf("ordersHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(orders); err != nil {
			logger.Log.Errorf("ordersHandler error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
