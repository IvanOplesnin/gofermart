package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	mw "github.com/IvanOplesnin/gofermart.git/internal/handler/middleware"
	"github.com/go-chi/chi/v5"
)

const (
	contentTypeKey       = "Content-Type"
	acceptEncodingKey    = "Accept-Encoding"
	contentEncodingKey   = "Content-Encoding"
	applicationJSONValue = "application/json"
	textPlainValue       = "text/plain"
	tokenCookieName      = "token"
)

type HandlerDeps struct {
	Reqistrar    Registrar
	Auther       Auther
	TokenChecker mw.TokenChecker
	Ordered      Ordered
	Balancer     Balancer
	Withdrawer   Withdrawer
}

func InitHandler(deps HandlerDeps) *chi.Mux {
	router := chi.NewRouter()
	router.Use(mw.WithLogging)

	router.Post("/api/user/register", Register(deps.Reqistrar))
	router.Post("/api/user/login", Login(deps.Auther))

	router.Group(func(pr chi.Router) {
		pr.Use(mw.CheckCookie(deps.TokenChecker))
		pr.Post("/api/user/orders", AddOrderHandler(deps.Ordered))
		pr.Get("/api/user/orders", OrdersHandler(deps.Ordered))
		pr.Get("/api/user/balance", BalanceHandler(deps.Balancer))
		pr.Post("/api/user/balance/withdraw", WithdrawHandler(deps.Withdrawer))
		pr.Get("/api/user/withdrawals", ListWithdrawHandler(deps.Withdrawer))
	})

	return router
}

func UserIDFromCtx(ctx context.Context) (int32, error) {
	claims, ok := ctx.Value(mw.ClaimsKey).(mw.Claims)
	if !ok {
		return 0, errors.New("user id not found")
	}
	return claims.UserID, nil
}

type RFC3339Time time.Time

func (t RFC3339Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(time.RFC3339))
}

func (t *RFC3339Time) UnmarshalJSON(data []byte) error {
	// null
	if bytes.Equal(data, []byte("null")) {
		*t = RFC3339Time(time.Time{})
		return nil
	}

	// строка
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*t = RFC3339Time(time.Time{})
		return nil
	}

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = RFC3339Time(tt)
	return nil
}
