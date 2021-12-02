package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

// Storage - структура для работы с postgress.
type Storage struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

// Connect - подключение к БД.
func (s *Storage) Connect(cfg *config.DBConfig) (bool, error) {
	if s == nil || cfg == nil {
		return false, errors.New("[STORAGE]::Connect : empty connection param")
	}

	var dbURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.NameDB)

	var err error

	if s.db, err = pgxpool.Connect(context.Background(), dbURL); err != nil {
		return false, errors.Wrap(err, "unable to connection to database: %v\n")
	}

	if err := s.db.Ping(context.Background()); err != nil {
		return false, err
	}

	s.log.Debug().Msg("[STORAGE]::Connect : success")

	return true, nil
}

// Close - закрытие подключения к БД.
func (s *Storage) Close() error {
	if s == nil || s.db == nil {
		return errors.New("[STORAGE]::Close : pgx.Conn == nil")
	}

	s.db.Close()
	s.log.Debug().Msg("[STORAGE]::Close : success")

	return nil
}
