package storage

import (
	"github.com/paulmach/orb"

	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/storage/models"
)

type Connector interface {
	Connect(cfg *config.DBConfig) (bool, error)
	Close() error
}

// GeoStorage - интерфейс для работы с БД , где хранятся геоданные.
type GeoStorage interface {
	Connector
	GetAllGeozones() ([]models.Geofences, error)
	GetFullGeometry() (map[uint64]*models.GeozoneExt, error)
}

// DevStorage - интерфейс для работы с БД.
type DevStorage interface {
	Connector
	GetGeoPoints([]orb.Point, error)
}

type MemoryGeoCache interface {
	Load() (count int, err error)
	Update() (count int, err error)
	FindGeoZoneByPont(point orb.Point) models.Geofences
}
