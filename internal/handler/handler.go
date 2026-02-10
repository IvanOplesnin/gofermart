package handler

import (
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

func InitHandler(reg Registrar, auther Auther) *chi.Mux {
	router := chi.NewRouter()
	router.Use(mw.WithLogging)

	router.Post("/api/user/register", Register(reg))
	router.Post("/api/user/login", Login(auther))

	return router
}
