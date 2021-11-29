package models

import (
	"fmt"
	"github.com/dhconnelly/rtreego"
	"github.com/paulmach/orb/geojson"
)

type Geofence struct {
	PolygonID   uint64 `json:"polygonId"`
	GeofenceID  uint64 `json:"geofenceId"`
	UserID      uint64 `json:"userId"`
	Title       string `json:"title"`
	Distance    float64
	BoundingBox *rtreego.Rect
}

func (t Geofence) Bounds() *rtreego.Rect {
	return t.BoundingBox
}

func (t Geofence) String() string {
	return fmt.Sprintf(" Geofence : %s, distanse = %f ", t.Title, t.Distance)
}

type GeofenceExt struct {
	PolygonID           uint64            `json:"polygonId"`
	GeofenceID          uint64            `json:"geofenceId"`
	Title               string            `json:"title"`
	UserID              uint64            `json:"userId"`
	GeometryFull        *geojson.Geometry `json:"geometryFull"`
	GeometrySimplify    geojson.Geometry  `json:"geometrySimplify"`
	GeometryBoundingBox string            `json:"geometryBoundingBox"`
	BoundingBox         *rtreego.Rect
}
