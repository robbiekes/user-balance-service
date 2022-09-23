package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"os"
	"user-balance-api/internal/service"
)

func NewRouter(handler *echo.Echo, services *service.Service) {
	handler.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}", "method":"${method}","uri":"${uri}", "status":${status},"error":"${error}"}` + "\n",
		Output: setLogsFile(),
	}))

	auth := handler.Group("/auth")
	{
		newAuthRoutes(auth, services)
	}

	authMiddleware := &AuthMiddleware{services.Auth}
	api := handler.Group("/api", authMiddleware.UserIdentity)
	{
		account := api.Group("/account")
		{
			newAccountRoutes(account, services.Account, services.History)

		}
		history := api.Group("/history")
		{
			newHistoryRoutes(history, services)
		}
	}
}

func setLogsFile() *os.File {
	file, err := os.OpenFile("/logs/requests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
