package models

import "time"

type Rate struct {
	ID        int64     `json:"id" db:"id"`               // Уникальный ID записи курса
	Ask       float64   `json:"ask" db:"ask"`             // Лучшая цена продажи (ask)
	Bid       float64   `json:"bid" db:"bid"`             // Лучшая цена покупки (bid)
	Timestamp time.Time `json:"timestamp" db:"timestamp"` // Временная метка получения курса
}

type HealthStatus struct {
	Status string `json:"status"`
}
