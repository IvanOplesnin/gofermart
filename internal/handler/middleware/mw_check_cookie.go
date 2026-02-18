package mw

import (
	"context"
	"errors"
	"net/http"
)

const tokenCookieName = "token"

type contextKey int

const ClaimsKey contextKey = iota

var ErrNotUserFound = errors.New("user not found")
var ErrInvalidPassword = errors.New("invalid password")
var ErrInvalidToken = errors.New("invalid token")

type TokenChecker interface {
	CheckToken(ctx context.Context, token string) (Claims, error)
}

type Claims struct {
	UserID uint64 `json:"user_id"`
}

func CheckCookie(cht TokenChecker) func(http.Handler) http.Handler {
	CheckToken := func(next http.Handler) http.Handler {
		checkCookieFunc := func(w http.ResponseWriter, r *http.Request) {
			var token string
			ctx := r.Context()
			c, err := r.Cookie(tokenCookieName)
			if err == nil && c != nil {
				token = c.Value
			} else if err != nil && !errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "failed to read cookie", http.StatusInternalServerError)
				return
			}
			if token != "" {
				claims, err := cht.CheckToken(ctx, token)
				if err == nil {
					ctx := context.WithValue(ctx, ClaimsKey, claims)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				} else if errors.Is(err, ErrNotUserFound) || errors.Is(err, ErrInvalidPassword) || errors.Is(err, ErrInvalidToken) {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				} else {
					http.Error(w, "failed to check token", http.StatusInternalServerError)
					return
				}
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}
		return http.HandlerFunc(checkCookieFunc)
	}
	return CheckToken
}
