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

	userCRUD UserCRUD
	Ordered  Ordered

	worker        *worker
	workerDb      ListUpdateApplyAccrual
	clientAccrual GetAPIOrdered

	withdrawDb WithdrawerDb

	balanceDb BalanceDb
}

var ErrNoRow = errors.New("no row")

type ServiceDeps struct {
	Hasher   Hasher
	UserCRUD UserCRUD
	Ordered  Ordered

	WorkerDB      ListUpdateApplyAccrual
	AccrualClient GetAPIOrdered

	WithdrawerDb WithdrawerDb
	BalanceDb    BalanceDb
}

func New(cfg *config.Config, deps ServiceDeps) (*Service, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gophermart.New: cfg is nil")
	}
	if deps.Hasher == nil {
		return nil, fmt.Errorf("gophermart.New: Hasher is nil")
	}
	if deps.UserCRUD == nil {
		return nil, fmt.Errorf("gophermart.New: UserCRUD is nil")
	}
	if deps.Ordered == nil {
		return nil, fmt.Errorf("gophermart.New: AddOrdered is nil")
	}
	if deps.WorkerDB == nil {
		return nil, fmt.Errorf("gophermart.New: WorkerDB is nil")
	}
	if deps.AccrualClient == nil {
		return nil, fmt.Errorf("gophermart.New: AccrualClient is nil")
	}
	if deps.WithdrawerDb == nil {
		return nil, fmt.Errorf("gophermart.New: WithdrawerDb is nil")
	}
	if deps.BalanceDb == nil {
		return nil, fmt.Errorf("gophermart.New: balanceDb is nil")
	}

	svc := &Service{
		hash:       deps.Hasher,
		secret:     []byte(cfg.Secret),
		userCRUD:   deps.UserCRUD,
		Ordered:    deps.Ordered,
		withdrawDb: deps.WithdrawerDb,
		balanceDb:  deps.BalanceDb,
	}

	svc.worker = newWorker(deps.AccrualClient, deps.WorkerDB)

	return svc, nil
}

func (s *Service) Start() {
	s.worker.Run()
}

func (s *Service) Stop() {
	s.worker.Stop()
}

func (s *Service) Register(ctx context.Context, login string, password string) (string, error) {
	const msg = "service.Register"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	hashPass, err := s.hash.HashPassword(password)
	if err != nil {
		return "", wrapError(err)
	}

	logger.Log.Debugf("login: %s, hash: %s", login, hashPass)
	userID, err := s.userCRUD.AddUser(ctx, login, hashPass)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return "", handler.ErrUserAlreadyExists
		}
		return "", wrapError(err)
	}
	tokenString, err := JwtToken(userID, s.secret)
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

func (s *Service) CheckToken(ctx context.Context, token string) (cl mw.Claims, err error) {
	const msg = "service.CheckToken"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	defer func() {
		if err != nil {
			logger.Log.Debugf("%s: %s", msg, err)
		}
	}()

	claims, err := ParseJwtToken(token, s.secret)
	if err != nil {
		return mw.Claims{}, err
	}

	logger.Log.Debugf("claims.UserID: %v", claims.UserID)

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
	if err != nil {
		return Claims{}, err
	}

	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	tok, err := parser.ParseWithClaims(token, &claims, keyFunc)
	if err != nil {
		logger.Log.Errorf("parseJwtError: %s", err.Error())
		return Claims{}, mw.ErrInvalidToken
	}

	if tok == nil || !tok.Valid {
		return Claims{}, mw.ErrInvalidToken
	}

	logger.Log.Debugf("claims.UserID: %v", claims.UserID)
	if claims.UserID == 0 {
		return Claims{}, mw.ErrNotUserFound
	}

	return Claims{UserID: claims.UserID}, nil
}

type Claims struct {
	UserID int32
}

func (c *Claims) String() string {
	return strconv.Itoa(int(c.UserID))
}

type JwtClaims struct {
	Claims
	jwt.RegisteredClaims
}

func JwtToken(userID int32, secret []byte) (string, error) {
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
