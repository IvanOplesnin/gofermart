package gophermart

import (
	"context"
	"errors"
	"fmt"

	"github.com/IvanOplesnin/gofermart.git/internal/handler"
)

func (s *Service) AddOrder(ctx context.Context, orderID string) (bool, error) {
	const msg = "service.AddOrder"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	if !validateLuna(orderID) {
		return false, handler.ErrInvalidOrderID
	}
<<<<<<< HEAD:internal/service/gophermart/order.go
	userIdFromContext, err := handler.UserIdFromCtx(ctx)
	created, owner, err := s.Ordered.CreateOrder(ctx, userIdFromContext, orderId)
=======
	userIDFromContext, err := handler.UserIDFromCtx(ctx)
	if err != nil {
		return false, wrapError(err)
	}
	created, owner, err := s.addOrdered.CreateOrder(ctx, userIDFromContext, orderID)
>>>>>>> master:internal/service/gophermart/add_order.go
	if err != nil {
		return false, wrapError(err)
	}
	if created {
		return false, nil
	}
	if owner != userIDFromContext {
		return false, handler.ErrAnotherUserOrder
	}
	return true, nil
}

// validateLuna checks a numeric string with the Luhn algorithm.
// Returns false for empty strings, non-digits, or if the check fails.
func validateLuna(number string) bool {
	// Optional: ignore spaces (sometimes numbers come formatted)
	// If you don't want that, remove this block.
	buf := make([]rune, 0, len(number))
	for _, r := range number {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		buf = append(buf, r)
	}
	if len(buf) == 0 {
		return false
	}

	sum := 0
	double := false // start from rightmost digit; double every second digit going left

	for i := len(buf) - 1; i >= 0; i-- {
		r := buf[i]
		if r < '0' || r > '9' {
			return false
		}
		d := int(r - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}
<<<<<<< HEAD:internal/service/gophermart/order.go

func (s *Service) Orders(ctx context.Context) ([]handler.Order, error) {
	const msg = "service.Orders"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	userId, err := handler.UserIdFromCtx(ctx)
	if err != nil {
		return nil, wrapError(err)
	}
	orders, err := s.Ordered.GetOrders(ctx, int32(userId))
	if errors.Is(err, ErrNoRow) {
		return []handler.Order{}, nil
	}
	if err != nil {
		return nil, wrapError(err)
	}
	respOrders := make([]handler.Order, 0, len(orders))
	for _, o := range orders {
		respOrders = append(respOrders, handler.Order{
			Number:     o.Number,
			Status:     o.OrderStatus,
			Accrual:    AccrualToFloatPtr(o.Accrual),
			UploadedAt: handler.RFC3339Time(o.UploadedAt),
		})
	}
	return respOrders, nil
}

func AccrualToFloatPtr(accrual int32) *float64 {
	if accrual == 0 {
		return nil
	}
	f := float64(accrual) / 100
	return &f
}
=======
>>>>>>> master:internal/service/gophermart/add_order.go
