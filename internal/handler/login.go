package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Auther interface {
	Auth(ctx context.Context, login string, password string) (string, error)
}

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidPassword = errors.New("invalid password")

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func Login(auther Auther) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(contentTypeKey) != applicationJSONValue {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var authReq AuthRequest
		if err := json.Unmarshal(raw, &authReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		token, err := auther.Auth(ctx, authReq.Login, authReq.Password)
		if errors.Is(err, ErrInvalidPassword) || errors.Is(err, ErrUserNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if token != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     tokenCookieName,
				Value:    token,
				Path:     "/api",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}
		w.WriteHeader(http.StatusOK)
	}
}
