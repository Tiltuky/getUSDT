package main

import (
	"context"
	"getUSDT/config"
	"getUSDT/internal/infrastructure/db/postgres"
	"getUSDT/run"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	jaegerPropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.0"
	"go.uber.org/zap"
)

const applicationID = "getUSDT-service"
const tracerURL = "http://host.docker.internal:14268/api/traces"

func main() {
	// Создание контекста для работы с завершением программы
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настройка провайдера трассировок с использованием экспортера Jaeger
	tp, err := setupTracerProvider(tracerURL)
	if err != nil {
		log.Fatalf("failed to initialize tracer provider: %v", err)
	}
	defer shutdownTracerProvider(ctx, tp)

	// Создание трассировщика
	tr := tp.Tracer(applicationID)

	// Загрузка конфигурации приложения
	cfg := config.MustLoad()

	// Инициализация логера с помощью zap
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync() // Закрытие логера
	}()

	// Инициализация подключения к базе данных PostgreSQL
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Warn("Failed to close database", zap.Error(err))
		}
	}()

	// Создание и запуск основного приложения
	application := run.NewApp(logger, cfg, db, tr)

	// Запуск приложения в отдельной горутине
	go application.MustRun()

	// Ожидание системного сигнала для завершения работы приложения
	waitForShutdown(logger, application)
}

// setupTracerProvider настраивает OpenTelemetry с использованием экспортера Jaeger
func setupTracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Создание экспортера Jaeger для отправки трассировок в указанный URL
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	// Настройка OpenTelemetry для использования Jaeger как пропагатора
	otel.SetTextMapPropagator(jaegerPropagator.Jaeger{})

	// Создание и настройка провайдера трассировок
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(applicationID),
		)),
	)

	return tp, nil
}

// shutdownTracerProvider корректно завершает работу провайдера трассировок
func shutdownTracerProvider(ctx context.Context, tp *tracesdk.TracerProvider) {
	// Создание контекста с тайм-аутом для корректного завершения работы
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Завершаем работу провайдера трассировок
	if err := tp.Shutdown(shutdownCtx); err != nil {
		log.Printf("failed to shut down tracer provider: %v", err)
	}
}

// waitForShutdown обрабатывает корректное завершение работы при получении системных сигналов
func waitForShutdown(logger *zap.Logger, application *run.App) {
	// Канал для получения системных сигналов
	stop := make(chan os.Signal, 1)

	// Подписываемся на сигналы завершения работы
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Ожидаем получения сигнала
	sig := <-stop
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	// Останавливаем приложение
	application.Stop()
	logger.Info("Application stopped gracefully")
}
