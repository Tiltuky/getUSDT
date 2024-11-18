package grpchealth

import (
	"context"
	"fmt"
	"getUSDT/internal/models"
	"getUSDT/proto/health/proto"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// HealthServer — структура для сервера health-сервиса
type HealthServer struct {
	tr            trace.Tracer
	healthService HealthService
	proto.HealthServer
}

type HealthService interface {
	CheckHealthStatus(ctx context.Context) (*models.HealthStatus, error)
}

// NewHealthServer создаёт новый HealthServer
func NewHealthServer(healthChecker HealthService, tr trace.Tracer) *HealthServer {
	return &HealthServer{
		healthService: healthChecker,
		tr:            tr,
	}
}

// Register регистрирует Health-сервис в gRPC сервере
func Register(gRPC *grpc.Server, chek HealthService, tr trace.Tracer) {
	proto.RegisterHealthServer(gRPC, NewHealthServer(chek, tr))
}

// CheckHealth проверяет состояние сервиса
func (s *HealthServer) Check(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	// Начинаем новый спан для трассировки
	ctx, span := s.tr.Start(ctx, "CheckHealth")
	defer span.End()

	// Проверка состояния здоровья через сервис
	status, err := s.healthService.CheckHealthStatus(ctx)
	if err != nil {
		return nil, err
	}

	// Логика для возвращения статуса через gRPC
	var servingStatus proto.HealthCheckResponse_ServingStatus
	if status.Status == "Healthy" {
		servingStatus = proto.HealthCheckResponse_SERVING
	} else if status.Status == "Unhealthy" {
		servingStatus = proto.HealthCheckResponse_NOT_SERVING
	} else {
		return nil, fmt.Errorf("unknown health status: %s", status.Status)
	}

	return &proto.HealthCheckResponse{
		Status: servingStatus,
	}, nil
}
