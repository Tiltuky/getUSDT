package run

import (
	"fmt"
	"getUSDT/config"
	"getUSDT/internal/modules/ratesService/service"
	"getUSDT/internal/modules/ratesService/storage"
	"getUSDT/internal/monitoring"
	"net"
	"net/http"

	grpchealth "getUSDT/internal/modules/health/gRPC"
	healthservice "getUSDT/internal/modules/health/service"
	grpcrate "getUSDT/internal/modules/ratesService/gRPC"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	log        *zap.Logger
	gRPCServer *grpc.Server
	port       int
}

func NewApp(log *zap.Logger, cfg *config.Config, dbPostgres *sqlx.DB, tr trace.Tracer) *App {
	// Создаем метрики
	metrics := monitoring.NewMetrics()
	// Создаем новый gRPC сервер с логированием
	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zap.UnaryServerInterceptor(log),
			monitoring.UnaryInterceptor(metrics),
		),
		grpc.ChainStreamInterceptor(
			grpc_zap.StreamServerInterceptor(log),
			monitoring.StreamInterceptor(metrics),
		),
	)

	// Инициализация хранилища и сервисов для RatesService
	PostgresStorage := storage.NewRatesStorage(dbPostgres)
	RatesService := service.NewRatesService(PostgresStorage)

	// Регистрация RatesServer
	grpcrate.NewRatesServer(RatesService, tr)
	grpcrate.Register(gRPCServer, RatesService, tr)

	// Регистрация HealthServer
	HealthService := healthservice.NewHealthService()
	grpchealth.NewHealthServer(HealthService, tr)
	grpchealth.Register(gRPCServer, HealthService, tr)

	// Экспозиция метрик через HTTP
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":9100", nil); err != nil {
			log.Error("HTTP server error", zap.Error(err))
		}
	}()

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       cfg.Local.Port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server is running", zap.String("address", l.Addr().String()))
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.log.With(zap.String("operation", op)).
		Info("grpc server is stopping", zap.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
