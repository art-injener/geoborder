package geocache

import (
	"fmt"
	"sync"

	"github.com/dhconnelly/rtreego"
	gogeo "github.com/kellydunn/golang-geo"
	"github.com/paulmach/orb"
	orbgeo "github.com/paulmach/orb/geo"
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
	geofenceExtCache map[uint64]*models.GeofenceExt
	// данные о геозонах, key - это id геозоны, значение id - полигона
	geofenceLinkedToPolygon map[uint64][]uint64
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
		db:                      db,
		geofenceExtCache:        make(map[uint64]*models.GeofenceExt),
		geofenceLinkedToPolygon: make(map[uint64][]uint64),
		rtree:                   rtreego.NewTree(2, 25, 4096),
	}, nil
}

func (m *MemoryGeoCache) Load() (count int, err error) {

	if m.geofenceExtCache, err = m.db.GetFullGeometry(); err != nil {
		logger.LogError(err, m.log)

		return 0, errors.Wrap(err, "error load full geometry")
	}

	m.RLock()
	defer m.RUnlock()

	for _, geofences := range m.geofenceExtCache {
		m.rtree.Insert(&models.Geofence{
			PolygonID:   geofences.PolygonID,
			GeofenceID:  geofences.GeofenceID,
			Title:       geofences.Title,
			UserID:      geofences.UserID,
			BoundingBox: geofences.BoundingBox,
		})

		polygons := m.geofenceLinkedToPolygon[geofences.GeofenceID]
		polygons = append(polygons, geofences.PolygonID)
		m.geofenceLinkedToPolygon[geofences.GeofenceID] = polygons
	}

	logger.LogDebug(fmt.Sprintf("[MEMORY_GEO_CAHCE]::Load : loaded %d geofences", m.rtree.Size()), m.log)

	return m.rtree.Size(), nil
}

// Update - самый простой вариант обновления кэша.
// Переделать на уведомления от postgres - https://habr.com/ru/company/tensor/blog/484978/
func (m *MemoryGeoCache) Update() (count int, err error) {

	var res map[uint64]*models.GeofenceExt
	ids := make([]uint64, 0, len(m.geofenceLinkedToPolygon))
	for _, v := range m.geofenceLinkedToPolygon {
		ids = append(ids, v...)
	}

	if res, err = m.db.GetNewRecords(ids); err != nil {
		logger.LogError(err, m.log)

		return 0, errors.Wrap(err, "error load full geometry")
	}

	if len(res) == 0 {
		return 0, nil
	}
	m.Lock()
	defer m.Unlock()
	for k, ext := range res {
		m.geofenceExtCache[k] = ext

		polygons := m.geofenceLinkedToPolygon[ext.GeofenceID]
		polygons = append(polygons, ext.PolygonID)
		m.geofenceLinkedToPolygon[ext.GeofenceID] = polygons
	}
	logger.LogDebug(fmt.Sprintf("[MEMORY_GEO_CACHE]::Update : add %d new records", len(res)), m.log)

	return len(res), nil
}

// FindGeofenceByPoint - поиск вхождения точки в геозону
// поиск разбит на 2 этапа:
// 1 этап - ищем в rtree пересечение точки с описывающим геозону прямоугольником
// 2 этап - проверяем по списку полученных прямоугольников вхождение точки в упрощенный полигон геозоны
func (m *MemoryGeoCache) FindGeofenceByPoint(point orb.Point, userID *uint64, withDistance bool) ([]models.Geofence, error) {
	rpt := rtreego.Point{point.X(), point.Y()}
	// выполняем поиск пересечения точки с описывающим геозону прямоугольником
	intersects := m.rtree.SearchIntersect(rpt.ToRect(0.001 / 2.))

	geofences := make([]models.Geofence, 0, len(intersects))

	var gz *models.Geofence
	var gzExt *models.GeofenceExt
	var isGeozone, ok bool

	for i := 0; i < len(intersects); i++ {
		// получаем информацию об описывающем геозону прямоугольнике
		if gz, isGeozone = intersects[i].(*models.Geofence); !isGeozone {
			continue
		}

		// делаем поиск расширенного описания геозоны по id полигона, который её описывает
		if gzExt, ok = m.geofenceExtCache[gz.PolygonID]; !ok {
			continue
		}

		if userID != nil && *userID != gzExt.UserID {
			continue
		}

		// получаем описание полигона
		if polygon, isPoly := gzExt.GeometrySimplify.Geometry().(orb.Polygon); isPoly {
			if planar.PolygonContains(polygon, point) {
				if withDistance {
					_, index := planar.DistanceFromWithIndex(polygon, point)
					gz.Distance = orbgeo.Distance(polygon[0][index], point)
				}
				geofences = append(geofences, *gz)

			}
		}
	}

	return geofences, nil
}
func (m *MemoryGeoCache) CheckGeofenceByPoint(point orb.Point, geofenceId []uint64) ([]models.Geofence, error) {

	geofences := make([]models.Geofence, 0, 2)
	for i := 0; i < len(geofenceId); i++ {

		polygonsId := m.geofenceLinkedToPolygon[geofenceId[i]]
		var gzExt *models.GeofenceExt
		var ok bool
		for i := 0; i < len(polygonsId); i++ {

			// делаем поиск расширенного описания геозоны по id полигона, который её описывает
			if gzExt, ok = m.geofenceExtCache[polygonsId[i]]; !ok {
				continue
			}

			// получаем описание полигона
			if polygon, isPoly := gzExt.GeometrySimplify.Geometry().(orb.Polygon); isPoly {
				if planar.PolygonContains(polygon, point) {
					geofences = append(geofences, models.Geofence{
						PolygonID:  gzExt.PolygonID,
						GeofenceID: gzExt.GeofenceID,
						UserID:     gzExt.UserID,
						Title:      gzExt.Title,
						Distance:   0,
					})
				}
			}
		}
	}
	return geofences, nil
}

func (m *MemoryGeoCache) GetDistanceToGeofence(point orb.Point) ([]models.Geofence, error) {
	return m.FindGeofenceByPoint(point, nil, true)
}

func (m *MemoryGeoCache) PolygonContainsGeo(polygon orb.Polygon, point orb.Point) bool {

	p := gogeo.NewPoint(point.Lat(), point.Lon())

	v := polygon[0]

	points := make([]*gogeo.Point, 0, len(v))
	for i := 0; i < len(v); i++ {
		points = append(points, gogeo.NewPoint(v[i].Lat(), v[i].Lon()))
	}

	pl := gogeo.NewPolygon(points)

	return pl.Contains(p)
}
