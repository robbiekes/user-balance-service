package v1

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-balance-service/internal/service"
	mock_service "user-balance-service/internal/service/mock"
)

func TestAuthMiddleware_UserIdentity(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	type MockBehaviour func(s *mock_service.MockAuth, token string)

	testCases := []struct {
		name            string
		args            args
		headerName      string
		headerValue     string
		token           string
		mockBehaviour   MockBehaviour
		wantStatusCode  int
		wantRequestBody string
	}{
		{
			name:        "OK",
			args:        args{ctx: context.Background()},
			headerName:  "Authorization",
			headerValue: "Bearer token",
			token:       "token",
			mockBehaviour: func(s *mock_service.MockAuth, token string) {
				s.EXPECT().ParseToken(token).Return(1, nil)
			},
			wantStatusCode:  200,
			wantRequestBody: "1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			auth := mock_service.NewMockAuth(ctrl)
			tc.mockBehaviour(auth, tc.token)
			services := service.Service{Auth: auth}
			authMiddleware := AuthMiddleware{services}

			e := echo.New()
			var handlerFunc echo.HandlerFunc
			e.GET("/api", authMiddleware.UserIdentity(handlerFunc))

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api", nil)

			req.Header.Set(tc.headerName, tc.headerValue)

			e.ServeHTTP(w, req)

			assert.Equal(t, w.Code, tc.wantStatusCode)
			assert.Equal(t, w.Body.String(), tc.wantRequestBody)
		})
	}
}
