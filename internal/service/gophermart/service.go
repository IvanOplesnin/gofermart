package gophermart

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/IvanOplesnin/gofermart.git/internal/config"
	"github.com/IvanOplesnin/gofermart.git/internal/handler"
	mw "github.com/IvanOplesnin/gofermart.git/internal/handler/middleware"
	"github.com/IvanOplesnin/gofermart.git/internal/logger"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	hash   Hasher
	secret []byte

	userCRUD   UserCRUD
	addOrdered AddOrdered
}

var ErrNoRow = errors.New("no row")

func New(cfg *config.Config, hasher Hasher, userCRUD UserCRUD, addOrdered AddOrdered) *Service {
	if cfg == nil || hasher == nil || userCRUD == nil || addOrdered == nil {
		return nil
	}

	return &Service{
		hash:       hasher,
		secret:     []byte(cfg.Secret),
		userCRUD:   userCRUD,
		addOrdered: addOrdered,
	}
}

func (s *Service) Register(ctx context.Context, login string, password string) (string, error) {
	const msg = "service.Register"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	hashPass, err := s.hash.HashPassword(password)
	if err != nil {
		return "", wrapError(err)
	}

	logger.Log.Debugf("login: %s, hash: %s", login, hashPass)
	userId, err := s.userCRUD.AddUser(ctx, login, hashPass)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return "", handler.ErrUserAlreadyExists
		}
		return "", wrapError(err)
	}
	tokenString, err := JwtToken(userId, s.secret)
	if err != nil {
		return "", wrapError(err)
	}
	return tokenString, nil
}

func (s *Service) Auth(ctx context.Context, login string, password string) (string, error) {
	const msg = "service.Auth"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	user, err := s.userCRUD.GetUserByLogin(ctx, login)
	if errors.Is(err, ErrNoRow) {
		return "", handler.ErrUserNotFound
	}
	if err != nil {
		return "", wrapError(err)
	}
	ok, err := s.hash.ComparePasswordHash(user.HashPassword, password)
	if err != nil {
		return "", wrapError(err)
	}
	if !ok {
		return "", handler.ErrInvalidPassword
	}

	tokenString, err := JwtToken(user.ID, s.secret)
	if err != nil {
		return "", wrapError(err)
	}
	return tokenString, nil
}

func (s *Service) CheckToken(ctx context.Context, token string) (mw.Claims, error) {
	const msg = "service.CheckToken"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	claims, err := ParseJwtToken(token, s.secret)
	if err != nil {
		return mw.Claims{}, err
	}

	_, err = s.userCRUD.GetUserByID(ctx, claims.UserID)
	if errors.Is(err, ErrNoRow) {
		return mw.Claims{}, handler.ErrUserNotFound
	}
	if err != nil {
		return mw.Claims{}, wrapError(err)
	}

	return mw.Claims{UserID: claims.UserID}, nil
}

func ParseJwtToken(token string, secret []byte) (Claims, error) {
	var claims JwtClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	tok, err := parser.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return Claims{}, mw.ErrInvalidToken
	}

	if tok == nil || !tok.Valid {
		return Claims{}, mw.ErrInvalidToken
	}

	if claims.UserID == 0 {
		return Claims{}, mw.ErrNotUserFound
	}

	return Claims{UserID: claims.UserID}, nil
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

func JwtToken(userID uint64, secret []byte) (string, error) {
	const msg = "service.JwtToken"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	claims := JwtClaims{
		Claims:           Claims{UserID: userID},
		RegisteredClaims: jwt.RegisteredClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", wrapError(err)
	}
	return tokenString, nil
}
