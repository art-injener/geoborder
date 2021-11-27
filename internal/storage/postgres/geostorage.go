package postgres

import (
	"context"

	"github.com/dhconnelly/rtreego"
	"github.com/jackc/pgx/v4"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/pkg/errors"

	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/storage/models"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

// GeoStorage - структура для работы с postgress.
type GeoStorage struct {
	Storage
}

// NewGeoStorage - Конструктор.
func NewGeoStorage(cfg *config.Config) *GeoStorage {
	return &GeoStorage{
		Storage{
			db:  nil,
			log: cfg.Log,
		},
	}
}

func (s *GeoStorage) GetAllGeozones() ([]models.Geofences, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT json_build_object("+
			"'id',       g.id, "+
			"'title',    g.title, "+
			"'userId',   g.user_id, "+
			"'geometry', ST_AsGeoJSON(polygon)::json) "+
			" FROM geo.gz_polygon gp inner join geo.geozone g on gp.gz_id = g.id;")

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.Wrap(err, "QueryRow failed")
	}

	defer rows.Close()

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	var geozones []models.Geofences

	for rows.Next() {
		var g models.Geofences
		err := rows.Scan(&g)

		if err != nil {
			logger.LogError(err, s.log)

			continue
		}

		geozones = append(geozones, g)
	}

	if err := rows.Err(); err != nil {
		logger.LogError(err, s.log)
	}

	return geozones, nil
}

// GetFullGeometry - загружаем всю информацию об геозонах.
func (s *GeoStorage) GetFullGeometry() (map[uint64]*models.GeozoneExt, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT json_build_object(  "+
			"'poligonId',  gp.id,"+
			"'geofenceId', g.id,"+
			"'title',      g.title,"+
			"'userId',     g.user_id, "+
			"'geometryFull',   ST_AsGeoJSON(polygon::geometry)::json,"+
			"'geometrySimplify', ST_Simplify(polygon::geometry,0.1,true)::json,"+
			"'geometryBoundingBox',  ST_AsBinary(ST_Extent(polygon::geometry))) "+
			"FROM geo.gz_polygon gp inner join geo.geozone g on gp.gz_id = g.id group by gp.id,g.id;")

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.Wrap(err, "QueryRow failed")
	}

	defer rows.Close()

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	var geozones = make(map[uint64]*models.GeozoneExt)

	for rows.Next() {
		var g models.GeozoneExt
		err := rows.Scan(&g)

		if err != nil {
			logger.LogError(err, s.log)

			continue
		}

		var p orb.Bound

		if err = wkb.Scanner(&p).Scan([]byte(g.GeometryBoundingBox)); err != nil {
			logger.LogError(err, s.log)

			continue
		}

		if g.BoundingBox, err = rtreego.NewRectFromPoints(
			rtreego.Point{p.Min.X(), p.Min.Y()},
			rtreego.Point{p.Max.X(), p.Max.Y()}); err != nil {
			logger.LogError(err, s.log)

			continue
		}

		gs, ok := g.GeometrySimplify.Coordinates.(orb.Polygon)
		// функция ST_Simplify в PostGis может обрезать полигон,
		// для такого случая будем использовать полный полигон геозоны
		if ok && (len(gs) > 0 && len(gs[0]) <= 4) {
			g.GeometrySimplify = *g.GeometryFull
		}

		g.GeometryFull = nil
		geozones[g.ID] = &g
	}

	if err := rows.Err(); err != nil {
		logger.LogError(err, s.log)
	}

	return geozones, nil
}
