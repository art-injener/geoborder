package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/paulmach/orb"
	"github.com/pkg/errors"

	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/storage/models"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

// DevStorage - структура для работы с postgress.
type DevStorage struct {
	Storage
}

// NewDevStorage - Конструктор.
func NewDevStorage(cfg *config.Config) *DevStorage {
	return &DevStorage{
		Storage{
			db:  nil,
			log: cfg.Log,
		},
	}
}

func (s *DevStorage) GetGeoPoints() ([]orb.Point, error) {
	rows, err := s.db.Query(context.Background(),
		"select st_asewkt(geo::geometry) from data_processed dp where geo IS NOT NULL;")

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.Wrap(err, "QueryRow failed")
	}

	defer rows.Close()

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	geoPoints := make([]orb.Point, 0)

	for rows.Next() {
		gp := models.GeoPoint{}
		err := rows.Scan(&gp)
		if err != nil {
			logger.LogError(err, s.log)

			continue
		}

		geoPoints = append(geoPoints, orb.Point{gp.Lon, gp.Lat})
	}

	if err := rows.Err(); err != nil {
		logger.LogError(err, s.log)
	}

	return geoPoints, nil
}
