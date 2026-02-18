package gophermart

import (
	"context"
	"fmt"

	"github.com/IvanOplesnin/gofermart.git/internal/handler"
)

func (s *Service) AddOrder(ctx context.Context, orderID string) (bool, error) {
	const msg = "service.AddOrder"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	if !validateLuna(orderID) {
		return false, handler.ErrInvalidOrderID
	}
	userIDFromContext, err := handler.UserIDFromCtx(ctx)
	if err != nil {
		return false, wrapError(err)
	}
	created, owner, err := s.addOrdered.CreateOrder(ctx, userIDFromContext, orderID)
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
