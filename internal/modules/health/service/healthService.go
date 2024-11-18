package healthservice

import (
	"context"
	"fmt"
	"getUSDT/internal/models"
	"log"
	"time"
)

// HealthService структура для реализации HealthService
type HealthService struct {
	startTime time.Time
}

// NewHealthService создаёт новый экземпляр HealthService
func NewHealthService() *HealthService {
	// Инициализируем startTime, чтобы отслеживать время работы приложения
	return &HealthService{
		startTime: time.Now(),
	}
}

// CheckHealthStatus проверяет статус здоровья приложения
func (h *HealthService) CheckHealthStatus(ctx context.Context) (*models.HealthStatus, error) {
	// Проверка на nil
	if h == nil {
		log.Println("HealthService is nil!")
		return nil, fmt.Errorf("HealthService is nil")
	}
	// Используем select для проверки отмены контекста
	select {
	case <-ctx.Done():
		return nil, ctx.Err() // Если контекст отменен, возвращаем ошибку
	default:
		healthyDuration := time.Since(h.startTime)
		if healthyDuration < time.Second*5 {
			// Если приложение только что запустилось, статус может быть "Initializing"
			return &models.HealthStatus{
				Status: "Initializing",
			}, nil
		}

		// Если прошло достаточно времени, приложение считается "Healthy"
		return &models.HealthStatus{
			Status: "Healthy",
		}, nil
	}
}
