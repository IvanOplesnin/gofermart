package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
)

type Registrar interface {
	Register(ctx context.Context, login string, password string) (string, error)
}

type RegiseterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var ErrUserAlreadyExists = errors.New("login is exists")
var ErrEmptyField = errors.New("empty field login or password")

func Register(reg Registrar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(contentTypeKey) != applicationJSONValue {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		raw, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var regReq RegiseterRequest
		if err := json.Unmarshal(raw, &regReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debugf("user Login: %s", regReq.Login)
		ctx := r.Context()

		if regReq.Login == "" || regReq.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := reg.Register(ctx, regReq.Login, regReq.Password)
		logger.Log.Debugf("token %s", token)
		if errors.Is(err, ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, ErrEmptyField) {
			w.WriteHeader(http.StatusBadRequest)
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
