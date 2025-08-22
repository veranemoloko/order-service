package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order/internal/config"
	handler "order/internal/delivery/http"
	"order/internal/delivery/kafka"
	model "order/internal/entity"
	"order/internal/infrastructure/cache"
	"order/internal/infrastructure/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/logger"
)

// must — хелпер для инициализации зависимостей.
// Если вернулась ошибка → логируем и выходим.
func must[T any](val T, err error, logger *slog.Logger, msg string) T {
	if err != nil {
		logger.Error(msg, "err", err)
		os.Exit(1)
	}
	return val
}

func main() {
	cfg := config.LoadConfig()

	// базовый логгер
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // в проде лучше брать из cfg.Service.LogLevel
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("Config loaded", "config", cfg)

	// --- инициализация зависимостей ---
	orderCache := must(cache.NewOrderCache(cfg.Service.CacheSize), logger, "failed to init cache")
	logger.Info("Order cache initialized")

	db := must(database.NewPostgresDB(cfg.DB), logger, "failed to connect to db")
	db.Config.Logger = gormLogger()
	logger.Info("Database connected with SQL logging enabled")

	repo := database.NewRepository(db, orderCache)
	logger.Info("Repository initialized")

	// --- запуск Kafka consumer ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewConsumer(cfg.Kafka.Broker, cfg.Kafka.Topic, cfg.Kafka.Group)
	go runConsumer(ctx, consumer, repo, logger.With("component", "kafka"))

	// --- запуск HTTP сервера ---
	srv := runHTTPServer(cfg.Service.HTTPPort, repo, logger.With("component", "http"))

	// --- graceful shutdown ---
	waitForShutdownSignal(logger)
	cancel()

	// закрываем Kafka
	if err := consumer.Close(); err != nil {
		logger.Error("failed to close Kafka consumer", "err", err)
	}

	// закрываем HTTP
	ctxShut, cancelShut := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShut()
	if err := srv.Shutdown(ctxShut); err != nil {
		logger.Error("HTTP server shutdown error", "err", err)
	}

	logger.Info("Service stopped gracefully")
}

// runConsumer запускает Kafka consumer с обработкой заказов.
func runConsumer(ctx context.Context, consumer *kafka.Consumer, repo *database.Repository, logger *slog.Logger) {
	err := consumer.Start(ctx, func(orders []model.Order) {
		for i := range orders {
			reqLogger := logger.With("order_uid", orders[i].OrderUID)

			err, updated := repo.InsertOrUpdateOrder(&orders[i])
			switch {
			case err != nil:
				reqLogger.Error("failed to save order", "err", err)
			case updated != nil:
				reqLogger.Info("order saved/updated")
			default:
				reqLogger.Debug("order already up-to-date")
			}
		}
	})

	if err != nil {
		logger.Error("Kafka consumer stopped", "err", err)
	}
}

// runHTTPServer поднимает Gin HTTP сервер.
func runHTTPServer(port string, repo *database.Repository, logger *slog.Logger) *http.Server {
	router := gin.Default()
	h := handler.NewHandler(repo)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Info("HTTP server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "err", err)
			os.Exit(1)
		}
	}()

	return srv
}

// waitForShutdownSignal блокирует выполнение до SIGINT/SIGTERM.
func waitForShutdownSignal(logger *slog.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Warn("Shutting down service", "signal", sig)
}

// gormLogger возвращает SQL логгер для GORM.
func gormLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)
}
