package gophermart

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/IvanOplesnin/gofermart.git/internal/config"
	"github.com/IvanOplesnin/gofermart.git/internal/handler"
	"github.com/IvanOplesnin/gofermart.git/internal/logger"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	hash   Hasher
	secret []byte

	userAdd UserAdd
}

func New(cfg *config.Config, hasher Hasher, userAdd UserAdd) *Service {
	if cfg == nil || hasher == nil || userAdd == nil {
		return nil
	}

	return &Service{
		hash:    hasher,
		secret:  []byte(cfg.Secret),
		userAdd: userAdd,
	}
}

type Claims struct {
	UserID uint64
}

func (c *Claims) String() string {
	return strconv.Itoa(int(c.UserID))
}

type JwtClaims struct {
	Claims
	jwt.RegisteredClaims
}

func (s *Service) Register(ctx context.Context, login string, password string) (string, error) {
	const msg = "service.Register"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	hashPass, err := s.hash.HashPassword(password)
	if err != nil {
		return "", wrapError(err)
	}
	
	logger.Log.Debugf("login: %s, hash: %s", login, hashPass)
	userId, err := s.userAdd.AddUser(ctx, login, hashPass)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return "", handler.ErrUserAlreadyExists
		}
		return "", wrapError(err)
	}
	claims := JwtClaims{
		Claims:           Claims{UserID: userId},
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", wrapError(err)
	}
	return tokenString, nil
}
