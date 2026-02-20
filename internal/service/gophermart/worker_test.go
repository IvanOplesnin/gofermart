package gophermart

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestWorker_checkAndUpdate_ListPendingError_NoCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockGetAPIOrdered(ctrl)
	db := NewMockListUpdateApplyAccrual(ctrl)

	w := newWorker(client, db)

	db.EXPECT().
		ListPending(gomock.Any(), int32(limitRequest), []string{"NEW", "PROCESSING"}, gomock.Any()).
		Return(nil, errors.New("db down")).
		Times(1)

	client.EXPECT().GetOrder(gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().ApplyAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateSyncTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w.checkAndUpdate(context.Background())
}

func TestWorker_checkAndUpdate_EmptyList_NoCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockGetAPIOrdered(ctrl)
	db := NewMockListUpdateApplyAccrual(ctrl)

	w := newWorker(client, db)

	db.EXPECT().
		ListPending(gomock.Any(), int32(limitRequest), []string{"NEW", "PROCESSING"}, gomock.Any()).
		Return([]Order{}, nil).
		Times(1)

	client.EXPECT().GetOrder(gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().ApplyAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateSyncTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w.checkAndUpdate(context.Background())
}

func TestWorker_checkAndUpdate_StatusChanged_Processed_ApplyAccrual(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockGetAPIOrdered(ctrl)
	db := NewMockListUpdateApplyAccrual(ctrl)

	w := newWorker(client, db)

	orders := []Order{
		{
			Number:      "123",
			OrderStatus: "NEW",
			UserID:      7,
		},
	}

	db.EXPECT().
		ListPending(gomock.Any(), int32(limitRequest), []string{"NEW", "PROCESSING"}, gomock.Any()).
		Return(orders, nil).
		Times(1)

	client.EXPECT().
		GetOrder(gomock.Any(), "123").
		Return(&AccrualResponse{OrderNumber: "123", Status: "PROCESSED", Accrual: 12.34}, nil).
		Times(1)

	db.EXPECT().
		ApplyAccrual(gomock.Any(), "123", int64(1234), int32(7)).
		Return(nil).
		Times(1)

	db.EXPECT().UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateSyncTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w.checkAndUpdate(context.Background())
}

func TestWorker_checkAndUpdate_StatusChanged_NotProcessed_UpdateFromAccrual(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockGetAPIOrdered(ctrl)
	db := NewMockListUpdateApplyAccrual(ctrl)

	w := newWorker(client, db)

	orders := []Order{
		{
			Number:      "555",
			OrderStatus: "NEW",
			UserID:      1,
		},
	}

	db.EXPECT().
		ListPending(gomock.Any(), int32(limitRequest), []string{"NEW", "PROCESSING"}, gomock.Any()).
		Return(orders, nil).
		Times(1)

	client.EXPECT().
		GetOrder(gomock.Any(), "555").
		Return(&AccrualResponse{OrderNumber: "555", Status: "INVALID", Accrual: 0}, nil).
		Times(1)

	db.EXPECT().
		UpdateFromAccrual(gomock.Any(), "555", "INVALID", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, nextSync time.Time) error {
			if time.Until(nextSync) < 90*time.Second {
				t.Fatalf("expected nextSync about now+120s, got %v", nextSync)
			}
			return nil
		}).
		Times(1)

	db.EXPECT().ApplyAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateSyncTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w.checkAndUpdate(context.Background())
}

func TestWorker_checkAndUpdate_StatusSame_UpdateSyncTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockGetAPIOrdered(ctrl)
	db := NewMockListUpdateApplyAccrual(ctrl)

	w := newWorker(client, db)

	orders := []Order{
		{
			Number:      "777",
			OrderStatus: "PROCESSING",
			UserID:      2,
		},
	}

	db.EXPECT().
		ListPending(gomock.Any(), int32(limitRequest), []string{"NEW", "PROCESSING"}, gomock.Any()).
		Return(orders, nil).
		Times(1)

	client.EXPECT().
		GetOrder(gomock.Any(), "777").
		Return(&AccrualResponse{OrderNumber: "777", Status: "PROCESSING", Accrual: 0}, nil).
		Times(1)

	db.EXPECT().
		UpdateSyncTime(gomock.Any(), "777", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, nextSync time.Time) error {
			if time.Until(nextSync) < 90*time.Second {
				t.Fatalf("expected nextSync about now+120s, got %v", nextSync)
			}
			return nil
		}).
		Times(1)

	db.EXPECT().ApplyAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w.checkAndUpdate(context.Background())
}

func TestWorker_checkAndUpdate_TooManyRequests_SetsRateLimit_NoDbUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockGetAPIOrdered(ctrl)
	db := NewMockListUpdateApplyAccrual(ctrl)

	w := newWorker(client, db)

	orders := []Order{
		{
			Number:      "999",
			OrderStatus: "NEW",
			UserID:      3,
		},
	}

	db.EXPECT().
		ListPending(gomock.Any(), int32(limitRequest), []string{"NEW", "PROCESSING"}, gomock.Any()).
		Return(orders, nil).
		Times(1)

	client.EXPECT().
		GetOrder(gomock.Any(), "999").
		Return(nil, ErrToManyRequests).
		Times(1)

	db.EXPECT().ApplyAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateSyncTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w.checkAndUpdate(context.Background())

	if !w.rateLimitid.Load() {
		t.Fatalf("expected rateLimitid=true after ErrToManyRequests")
	}
}

func TestWorker_checkAndUpdate_GetOrderError_NoDbUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockGetAPIOrdered(ctrl)
	db := NewMockListUpdateApplyAccrual(ctrl)

	w := newWorker(client, db)

	orders := []Order{
		{
			Number:      "1000",
			OrderStatus: "NEW",
			UserID:      4,
		},
	}

	db.EXPECT().
		ListPending(gomock.Any(), int32(limitRequest), []string{"NEW", "PROCESSING"}, gomock.Any()).
		Return(orders, nil).
		Times(1)

	// ВАЖНО: response != nil, иначе в worker будет panic на responseAccrual.Status
	client.EXPECT().
		GetOrder(gomock.Any(), "1000").
		Return(&AccrualResponse{OrderNumber: "1000", Status: "NEW", Accrual: 0}, errors.New("network error")).
		Times(1)

	db.EXPECT().ApplyAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	db.EXPECT().UpdateSyncTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w.checkAndUpdate(context.Background())
}
