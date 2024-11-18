package storage

import (
	"context"
	"fmt"
	"getUSDT/internal/models"

	"github.com/jmoiron/sqlx"
)

type RatesStorage struct {
	db *sqlx.DB
}

func NewRatesStorage(db *sqlx.DB) *RatesStorage {
	return &RatesStorage{
		db: db,
	}
}

func (s *RatesStorage) Close() error {
	return s.db.Close()
}

// SaveRate сохраняет курс USDT (Ask, Bid, Timestamp) в базе данных
func (s *RatesStorage) SaveRate(ctx context.Context, rate *models.Rate) error {
	query := `INSERT INTO rates (ask, bid) VALUES ($1, $2)`
	_, err := s.db.Exec(query, rate.Ask, rate.Bid)
	if err != nil {
		return fmt.Errorf("failed to execute insert query: %w", err)
	}
	return nil
}

// GetRatesFromAPI заглушка для обеспечения интерфейсной совместимости
func (s *RatesStorage) GetRatesFromAPI(ctx context.Context) (*models.Rate, error) {
	return nil, fmt.Errorf("GetRatesFromAPI is not implemented")
}
