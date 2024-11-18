package service

import (
	"context"
	"encoding/json"
	"fmt"
	"getUSDT/internal/models"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source=rateservice.go -destination=mocks/mock_rateservice.go -package=mocks

// RatesService структура для работы с курсами
type RatesService struct {
	storage RatesStorage
}

// RatesStorage интерфейс для взаимодействия с хранилищем данных
type RatesStorage interface {
	GetRatesFromAPI(ctx context.Context) (*models.Rate, error)
	SaveRate(ctx context.Context, rate *models.Rate) error
}

// NewRatesService создает новый экземпляр RatesService
func NewRatesService(storage RatesStorage) *RatesService {
	return &RatesService{
		storage: storage,
	}
}

// Структуры для парсинга ответа от Garantex API
type AskBid struct {
	Price  string `json:"price"`  // Цена на покупку/продажу
	Volume string `json:"volume"` // Объем
	Amount string `json:"amount"` // Сумма
	Factor string `json:"factor"` // Коэффициент
	Type   string `json:"type"`   // Тип
}

type ApiResponse struct {
	Asks []AskBid `json:"asks"` // Список заявок на покупку
	Bids []AskBid `json:"bids"` // Список заявок на продажу
}

// Получаем текущие курсы с биржи Garantex с трассировкой
func (s *RatesService) GetRatesFromAPI(ctx context.Context) (*models.Rate, error) {
	// Создаем трассировщик для отслеживания выполнения этой операции
	tracer := otel.Tracer("getUSDT.service")
	_, span := tracer.Start(ctx, "GetRatesFromAPI")
	defer span.End()

	// URL для запроса к API
	url := "https://garantex.org/api/v2/depth?market=usdtrub"
	span.SetAttributes(
		attribute.String("http.method", "GET"), // Метод HTTP запроса
		attribute.String("http.url", url),      // URL запроса
	)

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now() // Засекаем время начала запроса

	// Отправляем запрос и логируем события
	span.AddEvent("Sending HTTP request")
	resp, err := client.Get(url)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch rate from API")
		return nil, fmt.Errorf("failed to fetch rate from API: %w", err)
	}
	span.AddEvent("HTTP response received") // Ответ получен
	duration := time.Since(start)           // Время ответа
	span.SetAttributes(attribute.Float64("http.duration_ms", float64(duration.Milliseconds())))
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Non-200 status code")
		return nil, err
	}

	// Декодируем JSON ответ от API в структуру
	var apiResponse ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to decode API response")
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Проверяем наличие цен на покупку и продажу
	if len(apiResponse.Asks) == 0 || len(apiResponse.Bids) == 0 {
		err := fmt.Errorf("no ask/bid prices available in API response")
		span.RecordError(err)
		span.SetStatus(codes.Error, "No ask/bid prices available")
		return nil, err
	}

	// Преобразуем цены из строкового формата в числа с плавающей запятой
	askPrice, err := strconv.ParseFloat(apiResponse.Asks[0].Price, 64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to parse ask price")
		return nil, fmt.Errorf("failed to parse ask price: %w", err)
	}
	bidPrice, err := strconv.ParseFloat(apiResponse.Bids[0].Price, 64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to parse bid price")
		return nil, fmt.Errorf("failed to parse bid price: %w", err)
	}

	// Создаем объект модели курса и добавляем информацию в трассировку
	rate := &models.Rate{Ask: askPrice, Bid: bidPrice}
	span.SetAttributes(
		attribute.Float64("rate.ask", askPrice), // Цена на покупку
		attribute.Float64("rate.bid", bidPrice), // Цена на продажу
	)
	span.SetStatus(codes.Ok, "Operation completed successfully")

	return rate, nil
}

// Сохраняем курс с трассировкой
func (s *RatesService) SaveRate(ctx context.Context, rate *models.Rate) error {
	// Создаем трассировщик для отслеживания этой операции
	tracer := otel.Tracer("getUSDT.service")
	ctx, span := tracer.Start(ctx, "SaveRate")
	defer span.End()

	// Логируем попытку сохранения курса
	span.AddEvent("Attempting to save rate", trace.WithAttributes(
		attribute.Float64("rate.ask", rate.Ask), // Цена на покупку
		attribute.Float64("rate.bid", rate.Bid), // Цена на продажу
	))

	// Сохраняем курс в хранилище
	err := s.storage.SaveRate(ctx, rate)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to save rate")
		return fmt.Errorf("failed to save rate: %w", err)
	}

	// Логируем успешное сохранение курса
	span.AddEvent("Rate saved successfully")
	span.SetStatus(codes.Ok, "Rate saved successfully")
	return nil
}
