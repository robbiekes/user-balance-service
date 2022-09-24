package v1

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"user-balance-service/internal/service"
)

const (
	userIdCtx = "userId"
)

type AuthMiddleware struct {
	s service.Auth
}

func (h *AuthMiddleware) UserIdentity(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, ok := bearerToken(c.Request())
		if !ok {
			newErrorResponse(c, http.StatusUnauthorized, "invalid auth header")
			return nil
		}

		userId, err := h.s.ParseToken(token)
		if err != nil {
			newErrorResponse(c, http.StatusUnauthorized, err.Error())
			return err
		}

		c.Set(userIdCtx, userId)

		return next(c)
	}
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "

	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}

	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return header[len(prefix):], true
	}

	return "", false
}
