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

func TestWithdrawHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name        string
		contentType string
		body        []byte
		setupMock   func(m *MockWithdrawer)
		want        want
	}{
		{
			name:        "missing content-type -> 400",
			contentType: "",
			body:        []byte(`{"order":"12345678903","sum":10}`),
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{statusCode: http.StatusBadRequest},
		},
		{
			name:        "wrong content-type -> 400",
			contentType: "text/plain",
			body:        []byte(`{"order":"12345678903","sum":10}`),
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{statusCode: http.StatusBadRequest},
		},
		{
			name:        "content-type with charset -> ok to parse; success -> 200",
			contentType: applicationJSONValue + "; charset=utf-8",
			body:        []byte(`{"order":"12345678903","sum":10}`),
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().
					Withdraw(gomock.Any(), "12345678903", 10.0).
					Return(nil).
					Times(1)
			},
			want: want{statusCode: http.StatusOK},
		},
		{
			name:        "bad json -> 400",
			contentType: applicationJSONValue,
			body:        []byte(`{"order":"12345678903","sum":10`), // missing }
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{statusCode: http.StatusBadRequest},
		},
		{
			name:        "not enough money -> 402",
			contentType: applicationJSONValue,
			body:        []byte(`{"order":"12345678903","sum":10}`),
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().
					Withdraw(gomock.Any(), "12345678903", 10.0).
					Return(ErrEnoughMoney).
					Times(1)
			},
			want: want{statusCode: http.StatusPaymentRequired},
		},
		{
			name:        "invalid order number -> 422",
			contentType: applicationJSONValue,
			body:        []byte(`{"order":"bad","sum":10}`),
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().
					Withdraw(gomock.Any(), "bad", 10.0).
					Return(ErrInvalidOrderNumber).
					Times(1)
			},
			want: want{statusCode: http.StatusUnprocessableEntity},
		},
		{
			name:        "invalid summa -> 422",
			contentType: applicationJSONValue,
			body:        []byte(`{"order":"12345678903","sum":-1}`),
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().
					Withdraw(gomock.Any(), "12345678903", -1.0).
					Return(ErrInvalidSumma).
					Times(1)
			},
			want: want{statusCode: http.StatusUnprocessableEntity},
		},
		{
			name:        "unexpected error -> 500",
			contentType: applicationJSONValue,
			body:        []byte(`{"order":"12345678903","sum":10}`),
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().
					Withdraw(gomock.Any(), "12345678903", 10.0).
					Return(errors.New("db down")).
					Times(1)
			},
			want: want{statusCode: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wd := NewMockWithdrawer(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(wd)
			}

			h := WithdrawHandler(wd)

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(tt.body))
			req = req.WithContext(context.Background())

			if tt.contentType != "" {
				req.Header.Set(contentTypeKey, tt.contentType)
			}

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tt.want.statusCode {
				t.Fatalf("expected status %d, got %d", tt.want.statusCode, rr.Code)
			}
		})
	}
}

func TestListWithdrawHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		bodyCheck   func(t *testing.T, body []byte)
	}

	tests := []struct {
		name      string
		setupMock func(m *MockWithdrawer)
		want      want
	}{
		{
			name: "service error -> 500",
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().
					ListWithdraws(gomock.Any()).
					Return(nil, errors.New("db down")).
					Times(1)
			},
			want: want{statusCode: http.StatusInternalServerError},
		},
		{
			name: "empty list -> 204",
			setupMock: func(m *MockWithdrawer) {
				m.EXPECT().
					ListWithdraws(gomock.Any()).
					Return([]Withdraw{}, nil).
					Times(1)
			},
			want: want{statusCode: http.StatusNoContent},
		},
		{
			name: "non-empty list -> 200 and json",
			setupMock: func(m *MockWithdrawer) {
				ws := []Withdraw{
					{
						OrderNumber: "12345678903",
						Summa:       12.34,
						ProcessedAt: RFC3339Time(time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)),
					},
					{
						OrderNumber: "55555555555",
						Summa:       1.00,
						ProcessedAt: RFC3339Time(time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)),
					},
				}
				m.EXPECT().
					ListWithdraws(gomock.Any()).
					Return(ws, nil).
					Times(1)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, body []byte) {
					t.Helper()

					var got []Withdraw
					if err := json.Unmarshal(body, &got); err != nil {
						t.Fatalf("invalid json body %q: %v", string(body), err)
					}
					if len(got) != 2 {
						t.Fatalf("expected 2 withdraws, got %d", len(got))
					}
					if got[0].OrderNumber != "12345678903" {
						t.Fatalf("expected first order %q, got %q", "12345678903", got[0].OrderNumber)
					}
					if got[0].Summa != 12.34 {
						t.Fatalf("expected first sum %.2f, got %.2f", 12.34, got[0].Summa)
					}
				},
			},
		},
		{
			name: "json is array on success",
			setupMock: func(m *MockWithdrawer) {
				ws := []Withdraw{
					{
						OrderNumber: "1",
						Summa:       1,
						ProcessedAt: RFC3339Time(time.Now().UTC()),
					},
				}
				m.EXPECT().ListWithdraws(gomock.Any()).Return(ws, nil).Times(1)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, body []byte) {
					t.Helper()
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

			wd := NewMockWithdrawer(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(wd)
			}

			h := ListWithdrawHandler(wd)

			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
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
