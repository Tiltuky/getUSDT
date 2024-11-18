package monitoring

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// UnaryInterceptor для сбора метрик gRPC
func UnaryInterceptor(metrics *Metrics) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		metrics.RequestsTotal.Inc()                                  // Увеличиваем счетчик запросов
		metrics.RequestsLatency.Observe(time.Since(start).Seconds()) // Фиксируем задержку
		return resp, err
	}
}

// StreamInterceptor для сбора метрик потоковых запросов
func StreamInterceptor(metrics *Metrics) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		metrics.RequestsTotal.Inc()                                  // Увеличиваем счетчик запросов
		metrics.RequestsLatency.Observe(time.Since(start).Seconds()) // Фиксируем задержку
		return err
	}
}
