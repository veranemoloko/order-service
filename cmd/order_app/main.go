package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order/internal/config"
	httpDelivery "order/internal/delivery/http"
	"order/internal/delivery/kafka"
	"order/internal/infrastructure/cache"
	"order/internal/infrastructure/database"
	"order/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

func main() {
	cfg := config.LoadConfig()

	initLogger(&cfg.Service)

	orderCache := initCache(cfg)
	repo := initRepo(cfg, orderCache)

	ctx, cancel := context.WithCancel(context.Background())

	consumer := kafka.NewConsumer(
		cfg.Kafka.Broker,
		cfg.Kafka.Topic,
		cfg.Kafka.Group,
		cfg.Kafka.TopicDLQ,
	)

	go runKafkaConsumer(ctx, consumer, repo)

	srv := runHTTPServer(cfg, repo)

	gracefulShutdown(srv, consumer, cancel)
}

func initLogger(cfg *config.ServiceConfig) {
	var handler slog.Handler
	switch cfg.LogFormat {
	case "text":
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      cfg.LogLevel,
			TimeFormat: time.Kitchen,
			NoColor:    false,
		})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: cfg.LogLevel,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logger initialized", "level", cfg.LogLevel.String(), "format", cfg.LogFormat)
}

func initCache(cfg *config.Config) *cache.OrderCache {
	c, err := cache.NewOrderCache(cfg.Service.CacheSize)
	if err != nil {
		slog.Error("Failed to initialize cache", "err", err)
		os.Exit(1)
	}
	slog.Info("Order cache initialized")
	return c
}

func initRepo(cfg *config.Config, orderCache *cache.OrderCache) *database.Repository {
	db, err := database.NewPostgresDB(cfg.DB)
	if err != nil {
		slog.Error("Database connection failed", "err", err)
		os.Exit(1)
	}

	repo := database.NewRepository(db, orderCache)
	slog.Info("Repository initialized")

	if err := repo.RestoreCache(cfg.Service.CacheSize); err != nil {
		slog.Error("Failed to restore cache from DB", "err", err)
		os.Exit(1)
	}

	return repo
}

func runKafkaConsumer(ctx context.Context, consumer *kafka.Consumer, repo *database.Repository) {
	cLogger := slog.With("component", "kafka")

	err := consumer.Start(ctx, func(orders []model.Order) error {
		for i := range orders {
			cLogger.Debug("Processing order", "order", orders[i])
			_, err := repo.AddOrder(&orders[i])

			if err != nil {
				cLogger.Error("Failed to save order", "order_uid", orders[i].OrderUID, "err", err)
			} else {
				cLogger.Info("Order saved/updated", "order_uid", orders[i].OrderUID)
			}
		}
		return nil
	})

	if err != nil {
		cLogger.Error("Kafka consumer stopped", "err", err)
	}
}

func runHTTPServer(cfg *config.Config, repo *database.Repository) *http.Server {
	router := gin.New()
	h := httpDelivery.NewHandler(repo)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":" + cfg.Service.HTTPPort,
		Handler: router,
	}

	go func() {
		slog.Info("HTTP server starting", "port", cfg.Service.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "err", err)
		}
	}()

	return srv
}

func gracefulShutdown(srv *http.Server, consumer *kafka.Consumer, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	slog.Warn("Shutting down service", "signal", sig)

	// Cancel Kafka consumer context
	cancel()

	// Close Kafka consumer
	if err := consumer.Close(); err != nil {
		slog.Error("Error closing Kafka consumer", "err", err)
	} else {
		slog.Info("Kafka consumer closed successfully")
	}

	// Shutdown HTTP server with timeout
	ctxShut, cancelShut := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShut()
	if err := srv.Shutdown(ctxShut); err != nil {
		slog.Error("HTTP server shutdown error", "err", err)
	} else {
		slog.Info("HTTP server stopped gracefully")
	}

	slog.Info("Service stopped")
}
