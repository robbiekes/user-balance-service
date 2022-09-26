package app

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"user-balance-service/config"
	v1 "user-balance-service/internal/controller/http/v1"
	"user-balance-service/internal/service"
	"user-balance-service/internal/service/repo"
	"user-balance-service/internal/service/webapi"
	"user-balance-service/pkg/httpserver"
	"user-balance-service/pkg/postgres"
	"user-balance-service/pkg/rediscache"
)

func Run(cfg *config.Config) {
	// Cache
	log.Info("Initializing Redis...")
	redisCache := rediscache.New(redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}))

	// Repository
	log.Info("Initializing repository")
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Web API
	log.Info("Initializing webapi...")
	converterWebApi := webapi.NewConverterAPI(http.DefaultClient, cfg.Converter.URL, cfg.Converter.ApiKey)

	// Service
	log.Info("Initializing service...")
	services := service.New(
		repo.New(pg, redisCache),
		converterWebApi,
	)

	// HTTP Server
	log.Info("Initializing http server...")
	handler := echo.New()
	v1.NewRouter(handler, services)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	log.Info("Configuring graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Graceful shutdown
	log.Info("Shutting down...")
	err = httpServer.Shutdown()
	if err != nil {
		log.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
