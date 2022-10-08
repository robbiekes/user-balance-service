package v1

import (
	"bytes"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-balance-service/internal/entity"
	"user-balance-service/internal/service"
	mock_service "user-balance-service/internal/service/mock"
)

func TestControllerAuth_signUp(t *testing.T) {
	type args struct {
		ctx  context.Context
		user entity.User
	}

	type MockBehaviour func(s *mock_service.MockAuth, args args)

	testCases := []struct {
		name            string
		args            args
		inputBody       string
		mockBehaviour   MockBehaviour
		wantStatusCode  int
		wantRequestBody string
	}{
		{
			name: "OK",
			args: args{
				ctx: context.Background(),
				user: entity.User{
					Username: "test",
					Password: "qwerty",
				},
			},
			inputBody: `{"username":"test","password":"qwerty"}`,
			mockBehaviour: func(s *mock_service.MockAuth, args args) {
				s.EXPECT().CreateUser(args.ctx, args.user).Return(1, nil)
			},
			wantStatusCode:  200,
			wantRequestBody: `{"id":1}` + "\n",
		},
		{
			name: "Empty fields", // test-case doesn't work due to a bug in binding request body
			args: args{
				ctx: context.Background(),
			},
			inputBody:       `{"username":"test"}`,
			mockBehaviour:   func(s *mock_service.MockAuth, args args) {},
			wantStatusCode:  400,
			wantRequestBody: `{"message":"invalid input body"}` + "\n",
		},
		{
			name: "Service failure",
			args: args{
				ctx: context.Background(),
				user: entity.User{
					Username: "test",
					Password: "qwerty",
				},
			},
			inputBody: `{"username":"test","password":"qwerty"}`,
			mockBehaviour: func(s *mock_service.MockAuth, args args) {
				s.EXPECT().CreateUser(args.ctx, args.user).Return(1, errors.New("service failure"))
			},
			wantStatusCode:  500,
			wantRequestBody: `{"message":"service failure"}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// init deps
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create service mock
			auth := mock_service.NewMockAuth(ctrl)
			tc.mockBehaviour(auth, tc.args)
			services := &service.Service{Auth: auth}

			// create test server
			e := echo.New()
			g := e.Group("/auth")
			newAuthRoutes(g, services.Auth)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/auth/sign-up",
				bytes.NewBufferString(tc.inputBody))

			req.Header.Set("Content-Type", "application/json")

			// perform request
			e.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatusCode, w.Code)
			assert.Equal(t, tc.wantRequestBody, w.Body.String())
		})
	}
}
