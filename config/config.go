package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config структура для конфигурации приложения
type Config struct {
	Local Local    `yaml:"local"`
	DB    DBConfig `yaml:"db"`
}

// Local структура для конфигурации локальных параметров
type Local struct {
	Port int `yaml:"port"` // Порт для сервера
}

// DBConfig структура для конфигурации базы данных
type DBConfig struct {
	Host     string        `yaml:"host"`     // Хост базы данных
	Port     string        `yaml:"port"`     // Порт базы данных
	Username string        `yaml:"username"` // Имя пользователя базы данных
	Password string        `yaml:"password"` // Пароль базы данных
	DBName   string        `yaml:"dbname"`   // Имя базы данных
	SSlMode  string        `yaml:"sslmode"`  // Режим SSL для подключения
	Driver   string        `yaml:"driver"`   // Драйвер базы данных
	TimeOut  time.Duration `yaml:"timeout"`  // Таймаут подключения
}

// MustLoad загружает конфигурацию из файла и возвращает структуру Config
// Функция завершает выполнение программы с ошибкой, если конфигурацию не удается загрузить
func MustLoad() *Config {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Получаем путь к конфигурационному файлу из переменной окружения
	configPath, _ := os.LookupEnv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("Config path not found in environment file")
	}

	// Проверяем, существует ли файл конфигурации
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("Configuration file not found")
	}

	var cfg Config

	// Читаем конфигурацию из файла и заполняем структуру cfg
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Unable to read config: %s", err)
	}

	// Логируем информацию о загруженной конфигурации
	log.Printf("Config: %+v", cfg)

	return &cfg
}
