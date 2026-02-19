package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestAddOrderHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name        string
		contentType string
		body        []byte
		setupMock   func(m *MockOrdered)
		want        want
	}{
		{
			name:        "wrong content-type -> 400",
			contentType: "application/json",
			body:        []byte("123"),
			setupMock: func(m *MockOrdered) {
				m.EXPECT().AddOrder(gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{statusCode: http.StatusBadRequest},
		},
		{
			name:        "invalid order id -> 422",
			contentType: textPlainValue,
			body:        []byte("not-valid"),
			setupMock: func(m *MockOrdered) {
				m.EXPECT().
					AddOrder(gomock.Any(), "not-valid").
					Return(false, ErrInvalidOrderID).
					Times(1)
			},
			want: want{statusCode: http.StatusUnprocessableEntity},
		},
		{
			name:        "another user order -> 409",
			contentType: textPlainValue,
			body:        []byte("12345678903"),
			setupMock: func(m *MockOrdered) {
				m.EXPECT().
					AddOrder(gomock.Any(), "12345678903").
					Return(false, ErrAnotherUserOrder).
					Times(1)
			},
			want: want{statusCode: http.StatusConflict},
		},
		{
			name:        "unexpected error -> 500",
			contentType: textPlainValue,
			body:        []byte("12345678903"),
			setupMock: func(m *MockOrdered) {
				m.EXPECT().
					AddOrder(gomock.Any(), "12345678903").
					Return(false, errors.New("db down")).
					Times(1)
			},
			want: want{statusCode: http.StatusInternalServerError},
		},
		{
			name:        "order already exists (same user) -> 200",
			contentType: textPlainValue,
			body:        []byte("12345678903"),
			setupMock: func(m *MockOrdered) {
				m.EXPECT().
					AddOrder(gomock.Any(), "12345678903").
					Return(true, nil).
					Times(1)
			},
			want: want{statusCode: http.StatusOK},
		},
		{
			name:        "new order accepted -> 202",
			contentType: textPlainValue,
			body:        []byte("12345678903"),
			setupMock: func(m *MockOrdered) {
				m.EXPECT().
					AddOrder(gomock.Any(), "12345678903").
					Return(false, nil).
					Times(1)
			},
			want: want{statusCode: http.StatusAccepted},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ordered := NewMockOrdered(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(ordered)
			}

			h := AddOrderHandler(ordered)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set(contentTypeKey, tt.contentType)
			}
			// на всякий случай фиксируем контекст
			req = req.WithContext(context.Background())

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tt.want.statusCode {
				t.Fatalf("expected status %d, got %d", tt.want.statusCode, rr.Code)
			}
		})
	}
}

func TestOrdersHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		bodyCheck   func(t *testing.T, body []byte)
	}

	tests := []struct {
		name      string
		setupMock func(m *MockOrdered)
		want      want
	}{
		{
			name: "service error -> 500",
			setupMock: func(m *MockOrdered) {
				m.EXPECT().
					Orders(gomock.Any()).
					Return(nil, errors.New("db down")).
					Times(1)
			},
			want: want{statusCode: http.StatusInternalServerError},
		},
		{
			name: "empty list -> 204",
			setupMock: func(m *MockOrdered) {
				m.EXPECT().
					Orders(gomock.Any()).
					Return([]Order{}, nil).
					Times(1)
			},
			want: want{statusCode: http.StatusNoContent},
		},
		{
			name: "non-empty list -> 200 and json",
			setupMock: func(m *MockOrdered) {
				accrual := 12.34
				orders := []Order{
					{
						Number:     "12345678903",
						Status:     "NEW",
						Accrual:    &accrual,
						UploadedAt: RFC3339Time(time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)),
					},
					{
						Number:     "55555555555",
						Status:     "PROCESSED",
						Accrual:    nil,
						UploadedAt: RFC3339Time(time.Date(2026, 2, 19, 13, 0, 0, 0, time.UTC)),
					},
				}

				m.EXPECT().
					Orders(gomock.Any()).
					Return(orders, nil).
					Times(1)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, body []byte) {
					t.Helper()

					// Проверим, что это валидный JSON-массив Order
					var got []Order
					if err := json.Unmarshal(body, &got); err != nil {
						t.Fatalf("invalid json body %q: %v", string(body), err)
					}
					if len(got) != 2 {
						t.Fatalf("expected 2 orders, got %d", len(got))
					}
					if got[0].Number != "12345678903" {
						t.Fatalf("expected first order number %q, got %q", "12345678903", got[0].Number)
					}
				},
			},
		},
		{
			name: "encoder error path is hard to trigger; still validate no extra text on success",
			setupMock: func(m *MockOrdered) {
				orders := []Order{
					{
						Number:     "1",
						Status:     "NEW",
						Accrual:    nil,
						UploadedAt: RFC3339Time(time.Now().UTC()),
					},
				}
				m.EXPECT().Orders(gomock.Any()).Return(orders, nil).Times(1)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, body []byte) {
					t.Helper()
					// как минимум тело не пустое и похоже на json
					if len(body) == 0 {
						t.Fatalf("expected non-empty body")
					}
					if !strings.HasPrefix(strings.TrimSpace(string(body)), "[") {
						t.Fatalf("expected json array, got %q", string(body))
					}
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ordered := NewMockOrdered(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(ordered)
			}

			h := OrdersHandler(ordered)

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			req = req.WithContext(context.Background())

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tt.want.statusCode {
				t.Fatalf("expected status %d, got %d", tt.want.statusCode, rr.Code)
			}

			if tt.want.contentType != "" {
				ct := rr.Header().Get("Content-Type")
				if ct != tt.want.contentType {
					t.Fatalf("expected Content-Type %q, got %q", tt.want.contentType, ct)
				}
			}

			if tt.want.bodyCheck != nil {
				tt.want.bodyCheck(t, rr.Body.Bytes())
			}
		})
	}
}
