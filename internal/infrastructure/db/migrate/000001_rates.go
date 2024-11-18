package migrate

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upCreateRatesTable, downCreateRatesTable)
}

func upCreateRatesTable(tx *sql.Tx) error {
	// Создание таблицы rates
	_, err := tx.Exec(`
        CREATE TABLE IF NOT EXISTS rates (
            id SERIAL PRIMARY KEY,               -- Уникальный ID записи курса
            ask NUMERIC,       -- Лучшая цена продажи (ask)
            bid NUMERIC,       -- Лучшая цена покупки (bid)
            timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW() -- Временная метка получения курса
        );
    `)
	if err != nil {
		return fmt.Errorf("could not create rates table: %v", err)
	}

	return nil
}

func downCreateRatesTable(tx *sql.Tx) error {
	// Удаление таблицы rates
	_, err := tx.Exec(`
        DROP TABLE IF EXISTS rates;
    `)
	if err != nil {
		return fmt.Errorf("could not drop rates table: %v", err)
	}

	return nil
}
