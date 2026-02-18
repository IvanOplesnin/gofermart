package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
)

type Ordered interface {
	AddOrder(ctx context.Context, orderID string) (exist bool, err error)
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
