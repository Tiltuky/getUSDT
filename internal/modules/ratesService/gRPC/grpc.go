package grpcrates

import (
	"context"
	"fmt"
	"getUSDT/internal/models"
	"getUSDT/proto/usdt/proto"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type RatesServer struct {
	tr trace.Tracer
	proto.RatesServiceServer
	ratesService RatesService
}

type RatesService interface {
	GetRatesFromAPI(ctx context.Context) (*models.Rate, error)
	SaveRate(ctx context.Context, rate *models.Rate) error
}

func NewRatesServer(ratesService RatesService, tr trace.Tracer) *RatesServer {
	return &RatesServer{
		tr:           tr,
		ratesService: ratesService,
	}
}

func Register(gRPC *grpc.Server, rate RatesService, tr trace.Tracer) {
	proto.RegisterRatesServiceServer(gRPC, NewRatesServer(rate, tr))
}

// GetRates возвращает текущий курс USDT.
func (s *RatesServer) GetRates(ctx context.Context, req *proto.GetRatesRequest) (*proto.GetRatesResponse, error) {
	// Start a span
	ctx, span := s.tr.Start(ctx, "GetRates")
	defer span.End()

	// Добавляем атрибуты запроса к спану
	span.SetAttributes(
		attribute.String("rpc.method", "GetRates"),
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", "RatesService"),
	)

	// Получаем последний курс через сервис
	rate, err := s.ratesService.GetRatesFromAPI(ctx)
	if err != nil {
		// Добавляем атрибуты ошибки к спану
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "failed to fetch rate from API"))
		return nil, fmt.Errorf("failed to fetch rate from API: %w", err)
	}

	// Сохраняем курс
	if err := s.ratesService.SaveRate(ctx, rate); err != nil {
		// Добавляем атрибуты ошибки к спану
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "failed to save rate"))
		return nil, fmt.Errorf("failed to save rate: %w", err)
	}

	// Добавляем атрибуты успешного результата
	span.SetAttributes(
		attribute.Float64("rate.ask", rate.Ask),
		attribute.Float64("rate.bid", rate.Bid),
		attribute.Int64("rate.timestamp", rate.Timestamp.Unix()),
	)

	// Возвращаем данные через gRPC
	return &proto.GetRatesResponse{
		Ask:       rate.Ask,
		Bid:       rate.Bid,
		Timestamp: rate.Timestamp.Unix(),
	}, nil
}
