package models

import (
	"github.com/dhconnelly/rtreego"
	"github.com/paulmach/orb/geojson"
)

type Geofences struct {
	PoligonID   uint64 `json:"poligonId"`
	GeofenceID  uint64 `json:"geofenceId"`
	UserID      uint64 `json:"userId"`
	Title       string `json:"title"`
	BoundingBox *rtreego.Rect
}

func (t Geofences) Bounds() *rtreego.Rect {
	return t.BoundingBox
}

func (t Geofences) String() string {
	return t.Title
}

type GeozoneExt struct {
	ID                  uint64            `json:"poligonId"`
	GeofenceID          uint64            `json:"geofenceId"`
	Title               string            `json:"title"`
	UserID              uint64            `json:"userId"`
	GeometryFull        *geojson.Geometry `json:"geometryFull"`
	GeometrySimplify    geojson.Geometry  `json:"geometrySimplify"`
	GeometryBoundingBox string            `json:"geometryBoundingBox"`
	BoundingBox         *rtreego.Rect
}
