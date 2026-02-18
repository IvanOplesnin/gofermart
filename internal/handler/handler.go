package handler

import (
	"context"
	"errors"

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

func InitHandler(reg Registrar, auther Auther, tokenChecker mw.TokenChecker, addOrderer Ordered) *chi.Mux {
	router := chi.NewRouter()
	router.Use(mw.WithLogging)

	router.Post("/api/user/register", Register(reg))
	router.Post("/api/user/login", Login(auther))

	router.Group(func(pr chi.Router) {
		pr.Use(mw.CheckCookie(tokenChecker))
		pr.Post("/api/user/orders", AddOrderHandler(addOrderer))
	})

	return router
}

func UserIDFromCtx(ctx context.Context) (uint64, error) {
	claims, ok := ctx.Value(mw.ClaimsKey).(mw.Claims)
	if !ok {
		return 0, errors.New("user id not found")
	}
	return claims.UserID, nil
}
