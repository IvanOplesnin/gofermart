package mw

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestCheckCookieMiddleware(t *testing.T) {
	type want struct {
		statusCode int
		bodySubstr string
		nextCalled bool
		claims     *Claims 
	}

	tests := []struct {
		name      string
		setCookie func(r *http.Request)
		setupMock func(m *MockTokenChecker)
		want      want
	}{
		{
			name: "no cookie -> 401 unauthorized",
			setCookie: func(r *http.Request) {
			},
			setupMock: func(m *MockTokenChecker) {
				m.EXPECT().CheckToken(gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				bodySubstr: "unauthorized",
				nextCalled: false,
				claims:     nil,
			},
		},
		{
			name: "empty token cookie -> 401 unauthorized",
			setCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: tokenCookieName, Value: ""})
			},
			setupMock: func(m *MockTokenChecker) {
				m.EXPECT().CheckToken(gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				bodySubstr: "unauthorized",
				nextCalled: false,
				claims:     nil,
			},
		},
		{
			name: "valid token -> calls next and sets claims",
			setCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: tokenCookieName, Value: "abc"})
			},
			setupMock: func(m *MockTokenChecker) {
				m.EXPECT().
					CheckToken(gomock.Any(), "abc").
					Return(Claims{UserID: 42}, nil).
					Times(1)
			},
			want: want{
				statusCode: http.StatusOK,
				bodySubstr: "ok",
				nextCalled: true,
				claims:     &Claims{UserID: 42},
			},
		},
		{
			name: "ErrNotUserFound -> 401 invalid token",
			setCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: tokenCookieName, Value: "abc"})
			},
			setupMock: func(m *MockTokenChecker) {
				m.EXPECT().
					CheckToken(gomock.Any(), "abc").
					Return(Claims{}, ErrNotUserFound).
					Times(1)
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				bodySubstr: "invalid token",
				nextCalled: false,
				claims:     nil,
			},
		},
		{
			name: "ErrInvalidPassword -> 401 invalid token",
			setCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: tokenCookieName, Value: "abc"})
			},
			setupMock: func(m *MockTokenChecker) {
				m.EXPECT().
					CheckToken(gomock.Any(), "abc").
					Return(Claims{}, ErrInvalidPassword).
					Times(1)
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				bodySubstr: "invalid token",
				nextCalled: false,
				claims:     nil,
			},
		},
		{
			name: "ErrInvalidToken -> 401 invalid token",
			setCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: tokenCookieName, Value: "abc"})
			},
			setupMock: func(m *MockTokenChecker) {
				m.EXPECT().
					CheckToken(gomock.Any(), "abc").
					Return(Claims{}, ErrInvalidToken).
					Times(1)
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				bodySubstr: "invalid token",
				nextCalled: false,
				claims:     nil,
			},
		},
		{
			name: "unexpected error -> 500 failed to check token",
			setCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: tokenCookieName, Value: "abc"})
			},
			setupMock: func(m *MockTokenChecker) {
				m.EXPECT().
					CheckToken(gomock.Any(), "abc").
					Return(Claims{}, errors.New("db down")).
					Times(1)
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				bodySubstr: "failed to check token",
				nextCalled: false,
				claims:     nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt 
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tokenChecker := NewMockTokenChecker(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(tokenChecker)
			}

			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				if tt.want.claims != nil {
					v := r.Context().Value(ClaimsKey)
					if v == nil {
						t.Fatalf("expected claims in context, got nil")
					}
					got, ok := v.(Claims)
					if !ok {
						t.Fatalf("expected Claims type in context, got %T", v)
					}
					if got.UserID != tt.want.claims.UserID {
						t.Fatalf("expected UserID=%d, got %d", tt.want.claims.UserID, got.UserID)
					}
				} else {
					if v := r.Context().Value(ClaimsKey); v != nil {
						t.Fatalf("expected no claims in context, got %v", v)
					}
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})

			mw := CheckCookie(tokenChecker)
			h := mw(next)

			req := httptest.NewRequest(http.MethodGet, "/any", nil)

			req = req.WithContext(context.Background())

			if tt.setCookie != nil {
				tt.setCookie(req)
			}

			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tt.want.statusCode {
				t.Fatalf("expected status %d, got %d", tt.want.statusCode, rr.Code)
			}

			body := rr.Body.String()
			if tt.want.bodySubstr != "" && !strings.Contains(body, tt.want.bodySubstr) {
				t.Fatalf("expected body to contain %q, got %q", tt.want.bodySubstr, body)
			}

			if nextCalled != tt.want.nextCalled {
				t.Fatalf("expected nextCalled=%v, got %v", tt.want.nextCalled, nextCalled)
			}
		})
	}
}
