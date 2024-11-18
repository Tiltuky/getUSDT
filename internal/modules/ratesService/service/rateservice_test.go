package service

import (
	"context"
	"errors"
	"getUSDT/internal/models"
	"getUSDT/internal/modules/ratesService/service/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSaveRate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockRatesStorage(ctrl)

	// Создаем тестовый курс
	rate := &models.Rate{Ask: 100.5, Bid: 99.5}

	// Задаем ожидаемое поведение для мок-метода SaveRate
	mockStorage.EXPECT().SaveRate(gomock.Any(), rate).Return(nil).Times(1)

	// Создаем экземпляр RatesService с мок-стореджем
	service := NewRatesService(mockStorage)

	// Выполняем тестируемую функцию
	err := service.SaveRate(context.Background(), rate)

	// Проверяем результаты
	assert.NoError(t, err)
}

func TestSaveRate_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockRatesStorage(ctrl)

	// Создаем тестовый курс
	rate := &models.Rate{Ask: 100.5, Bid: 99.5}

	// Задаем ожидаемое поведение для мок-метода SaveRate, который вернет ошибку
	mockStorage.EXPECT().SaveRate(gomock.Any(), rate).Return(errors.New("save error")).Times(1)

	// Создаем экземпляр RatesService с мок-стореджем
	service := NewRatesService(mockStorage)

	// Выполняем тестируемую функцию
	err := service.SaveRate(context.Background(), rate)

	// Проверяем результаты
	assert.Error(t, err)
}
