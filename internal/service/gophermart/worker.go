package gophermart

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
	"github.com/sirupsen/logrus"
)

const (
	pollingInterval = 10 * time.Second
	limitRequest    = 10
)

var ErrToManyRequests = errors.New("too many requests")

type worker struct {
	accrualClient GetApiOrdered
	checkerDb     ListUpdateApplyAccrual

	rateLimitid atomic.Bool
	cancelLoop  func()

	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup
	chLimit   chan struct{}
}

func newWorker(client GetApiOrdered, checker ListUpdateApplyAccrual) *worker {
	return &worker{
		accrualClient: client,
		checkerDb:     checker,
		chLimit:       make(chan struct{}, limitRequest),
		rateLimitid:   atomic.Bool{},

		startOnce: sync.Once{},
		stopOnce:  sync.Once{},
		wg:        sync.WaitGroup{},
	}
}

func (w *worker) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancelLoop = cancel
	w.startOnce.Do(func() {
		w.wg.Add(1)
		go w.loop(ctx)
	})
}

func (w *worker) Stop() {
	if w.cancelLoop != nil {
		w.stopOnce.Do(func() {
			w.cancelLoop()
		})
	}
	w.wg.Wait()
}

func (w *worker) loop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		if w.rateLimitid.Load() {
			sleepCh := time.NewTimer(60 * time.Second)
			select {
			case <-ctx.Done():
				sleepCh.Stop()
				return
			case <-sleepCh.C:
				w.rateLimitid.Store(false)
				continue
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkAndUpdate(ctx)
		}
	}
}

func (w *worker) checkAndUpdate(ctx context.Context) {
	now := time.Now()
	orders, err := w.checkerDb.ListPending(ctx, limitRequest, []string{"NEW", "PROCESSING"}, now)
	if err != nil {
		logger.Log.Errorf("svc.worker.checkAndUpdate: %v", err.Error())
		return
	}
	if len(orders) == 0 {
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"count_orders": len(orders),
		"orders":       fmt.Sprintf("%s", listOrdersString(orders)),
	}).Infof("worker.checkAndUpdate run")
	ctxBatch, cancel := context.WithCancel(ctx)
	defer cancel()
	once := sync.Once{}
	var wg sync.WaitGroup
Loop:
	for _, order := range orders {
		o := order
		select {
		case <-ctxBatch.Done():
			break Loop
		case w.chLimit <- struct{}{}:
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-w.chLimit }()
			responseAccrual, err := w.accrualClient.GetOrder(ctxBatch, o.Number)
			if errors.Is(err, ErrToManyRequests) {
				logger.Log.Warnf("too many requests to accrual service: code 429")
				w.rateLimitid.Store(true)
				once.Do(cancel)
				return
			}
			if err != nil {
				logger.Log.Errorf("svc.worker.checkAndUpdate: %s", err.Error())
				return
			}
			if responseAccrual == nil {
				logger.Log.Error("svc.worker.checkAndUpdate: responseAccrual == nil ")
				return
			}
			if responseAccrual.Status != o.OrderStatus {
				if responseAccrual.Status == "PROCESSED" {
					if err := w.checkerDb.ApplyAccrual(ctx, o.Number, int64(*responseAccrual.Accrual), o.UserID); err != nil {
						logger.Log.Errorf("svc.worker.checkAndUpdate: %s", err.Error())
						return
					}
				} else if err := w.checkerDb.UpdateFromAccrual(ctx, o.Number, responseAccrual.Status, now.Add(time.Second*120)); err != nil {
					logger.Log.Errorf("svc.worker.checkAndUpdate: %s", err.Error())
					return
				}
			} else {
				if err := w.checkerDb.UpdateSyncTime(ctx, o.Number, now.Add(time.Second*120)); err != nil {
					logger.Log.Errorf("svc.worker.checkAndUpdate: %s", err.Error())
					return
				}
			}
		}()
	}
	wg.Wait()
}

func listOrdersString(orders []Order) string {
	if len(orders) == 0 {
		return "[]"
	}
	listNumbers := make([]string, 0, len(orders))
	for _, order := range orders {
		listNumbers = append(listNumbers, order.Number)
	}
	return strings.Join(listNumbers, ",")
}
