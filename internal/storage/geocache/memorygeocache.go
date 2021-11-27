package geocache

import (
	"fmt"
	"sync"

	"github.com/dhconnelly/rtreego"
	geo "github.com/kellydunn/golang-geo"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/pkg/errors"

	"github.com/X-Keeper/geoborder/internal/storage"
	"github.com/X-Keeper/geoborder/internal/storage/models"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

// MemoryGeoCache - in-memory cache для хранения информации о геозонах.
type MemoryGeoCache struct {
	sync.RWMutex

	// БД для синхронизации данных
	db storage.GeoStorage
	// данные о геозонах, key - это id полигона, который описывает геозону
	geozones map[uint64]*models.GeozoneExt
	// сбалансированное дерево поиска для хранения и запросов bounding box геозон
	rtree *rtreego.Rtree
	// логгирование
	log *logger.Logger
}

func NewMemoryCache(db storage.GeoStorage) (*MemoryGeoCache, error) {
	if db == nil {
		return nil, errors.New("no database connection")
	}

	return &MemoryGeoCache{
		db:       db,
		geozones: make(map[uint64]*models.GeozoneExt),
		rtree:    rtreego.NewTree(2, 25, 4096),
	}, nil
}

func (m *MemoryGeoCache) Load() (count int, err error) {
	m.RLock()
	defer m.RUnlock()

	if m.geozones, err = m.db.GetFullGeometry(); err != nil {
		logger.LogError(err, m.log)

		return 0, errors.Wrap(err, "error load full geometry")
	}

	for _, geofences := range m.geozones {
		m.rtree.Insert(&models.Geofences{
			PoligonID:   geofences.ID,
			GeofenceID:  geofences.GeofenceID,
			Title:       geofences.Title,
			UserID:      geofences.UserID,
			BoundingBox: geofences.BoundingBox,
		})
	}

	logger.LogDebug(fmt.Sprintf("[MEMORY_GEO_CAHCE]::Load : loaded %d geofences", m.rtree.Size()), m.log)

	return m.rtree.Size(), nil
}

func (m *MemoryGeoCache) Update() (count int, err error) {
	m.Lock()
	defer m.Unlock()

	// придумать как обновлять данные в кэше

	return 0, nil
}

// FindGeoZoneByPont - поиск вхождения точки в геозону
// поиск разбит на 2 этапа:
// 1 этап - ищем в rtree пересечение точки с описывающим геозону прямоугольником
// 2 этап - проверяем по списку полученных прямоугольников вхождение точки в упрощенный полигон геозоны
func (m *MemoryGeoCache) FindGeoZoneByPont(point orb.Point) ([]models.Geofences, error) {
	rpt := rtreego.Point{point.X(), point.Y()}
	// выполняем поиск пересечения точки с описывающим геозону прямоугольником
	intersects := m.rtree.SearchIntersect(rpt.ToRect(0.001 / 2.))

	geozones := make([]models.Geofences, 0, len(intersects))

	var gz *models.Geofences
	var gzExt *models.GeozoneExt
	var isGeozone, ok bool

	for i := 0; i < len(intersects); i++ {
		// получаем информацию об описывающем геозону прямоугольнике
		if gz, isGeozone = intersects[i].(*models.Geofences); !isGeozone {
			continue
		}
		// делаем поиск расширенного описания геозоны по id полигона, который её описывает
		if gzExt, ok = m.geozones[gz.PoligonID]; !ok {
			continue
		}

		// получаем описание полигона
		if polygon, isPoly := gzExt.GeometrySimplify.Geometry().(orb.Polygon); isPoly {
			if planar.PolygonContains(polygon, point) {
				geozones = append(geozones, *gz)
			}
		}
	}

	return geozones, nil
}

func (m *MemoryGeoCache) PolygonContainsGeo(polygon orb.Polygon, point orb.Point) bool {

	p := geo.NewPoint(point.Lat(), point.Lon())

	v := polygon[0]

	points := make([]*geo.Point, 0, len(v))
	for i := 0; i < len(v); i++ {
		points = append(points, geo.NewPoint(v[i].Lat(), v[i].Lon()))
	}

	pl := geo.NewPolygon(points)

	return pl.Contains(p)
}
