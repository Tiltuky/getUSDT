package postgres

import (
	"database/sql"
	"fmt"
	"getUSDT/config"
	"log"
	"time"

	_ "getUSDT/internal/infrastructure/db/migrate"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

func NewPostgresDB(cfg *config.Config) (*sqlx.DB, error) {
	var dsn string
	var err error
	var dbRaw *sql.DB

	dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.DBName, cfg.DB.SSlMode)
	fmt.Println("Connecting with DSN:", dsn)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	timeoutExceeded := time.After(time.Second * cfg.DB.TimeOut)

	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("db connection failed after %d timeout %s", 5, err)
		case <-ticker.C:
			dbRaw, err = sql.Open(cfg.DB.Driver, dsn)
			if err != nil {
				return nil, fmt.Errorf("db connection failed %s", err)
			}
			err = dbRaw.Ping()
			if err == nil {

				db := sqlx.NewDb(dbRaw, cfg.DB.Driver)

				err = goose.Up(dbRaw, "./")
				if err != nil {
					log.Fatal("Goose up failed ", err)
				}
				return db, nil
			}

			log.Fatal("failed to connect to the database", err)
		}
	}
}
