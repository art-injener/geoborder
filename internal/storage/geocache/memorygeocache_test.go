package geocache

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/dhconnelly/rtreego"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"

	"github.com/X-Keeper/geoborder/internal/storage"
	"github.com/X-Keeper/geoborder/internal/storage/models"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

var geoJSONRostovOnDon = `{
  "poligonId": 7452,
  "geofenceId": 221,
  "title": "Ростов-на-Дону",
  "userId": 22217,
  "geometrySimplify": {
    "type": "Polygon",
    "coordinates": [
      [
        [
          39.593353,
          47.318345
        ],
        [
          39.943542,
          47.322069
        ],
        [
          39.881744,
          47.174778
        ],
        [
          39.571381,
          47.173845
        ],
        [
          39.593353,
          47.318345
        ]
      ]
    ]
  },
  "geometryBoundingBox": "\\x0103000000010000000500000024473a0323c943405cc98e8d4096474024473a0323c943401807978e39a947408333f8fbc5f843401807978e39a947408333f8fbc5f843405cc98e8d4096474024473a0323c943405cc98e8d40964740"
}`

var geoJsonMoscow = `{
  "id": 3734,
  "geoZoneId": 50,
  "title": "Москва",
  "userId": 0,
  "geometryFull": {
    "type": "Polygon",
    "coordinates": [
      [
        [
          37.308213,
          55.770701
        ],
        [
          37.310755,
          55.775369
        ],
        [
          37.322191,
          55.776973
        ],
        [
          37.328227,
          55.781203
        ],
        [
          37.31266,
          55.786308
        ],
        [
          37.31139,
          55.79258
        ],
        [
          37.322826,
          55.800893
        ],
        [
          37.333311,
          55.79958
        ],
        [
          37.347924,
          55.806873
        ],
        [
          37.354595,
          55.804831
        ],
        [
          37.352053,
          55.797101
        ],
        [
          37.357772,
          55.792725
        ],
        [
          37.367938,
          55.789954
        ],
        [
          37.352689,
          55.777848
        ],
        [
          37.352371,
          55.774348
        ],
        [
          37.344429,
          55.76866
        ],
        [
          37.320285,
          55.767055
        ],
        [
          37.318696,
          55.769535
        ],
        [
          37.308213,
          55.770701
        ],
        [
          37.308213,
          55.770701
        ]
      ]
    ]
  },
  "geometrySimplify": {
    "type": "Polygon",
    "coordinates": [
      [
        [
          37.308213,
          55.770701
        ],
        [
          37.322826,
          55.800893
        ],
        [
          37.367938,
          55.789954
        ],
        [
          37.308213,
          55.770701
        ]
      ]
    ]
  },
  "geometryBoundingBox": "\\x01030000000100000005000000dd99098673a74240dc9db5db2ee24b40dd99098673a7424041834d9d47e74b40537aa69718af424041834d9d47e74b40537aa69718af4240dc9db5db2ee24b40dd99098673a74240dc9db5db2ee24b40"
}`

func TestMemoryGeoCache_FindGeoZoneByPont(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		db       storage.GeoStorage
		geozones map[uint64]*models.GeozoneExt
		rtree    *rtreego.Rtree
		log      *logger.Logger
	}
	type args struct {
		point orb.Point
	}

	geofence := models.GeozoneExt{}

	if err := json.Unmarshal([]byte(geoJSONRostovOnDon),&geofence);err != nil {
		fmt.Errorf(err.Error())
	}

	var p orb.Bound
	var err error

	if err = wkb.Scanner(&p).Scan([]byte(geofence.GeometryBoundingBox)); err != nil{
		fmt.Errorf(err.Error())
	}

	if geofence.BoundingBox, err = rtreego.NewRectFromPoints(
		rtreego.Point{p.Min.X(), p.Min.Y()},
		rtreego.Point{p.Max.X(), p.Max.Y()}); err != nil {
		fmt.Println(err.Error())
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "TestRostove",
			fields:  fields{
				RWMutex:  sync.RWMutex{},
				db:       nil,
				geozones: func() map[uint64]*models.GeozoneExt {
					m := make(map[uint64]*models.GeozoneExt)
					m[7452] = &geofence

					return m
				}(),
				rtree:   func() *rtreego.Rtree{
					rt := rtreego.NewTree(2, 25, 2)
					rt.Insert(&models.Geofences{
						PoligonID:   geofence.ID,
						GeofenceID:  geofence.GeofenceID,
						Title:       geofence.Title,
						UserID:      geofence.UserID,
						BoundingBox: geofence.BoundingBox,
					})

					return rt
				} () ,
				log:      nil,
			},
			args:    args{orb.Point{39.70151,47.23571}},
			want:   1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemoryGeoCache{
				RWMutex:  tt.fields.RWMutex,
				db:       tt.fields.db,
				geozones: tt.fields.geozones,
				rtree:    tt.fields.rtree,
				log:      tt.fields.log,
			}

			got, err := m.FindGeoZoneByPont(tt.args.point)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindGeoZoneByPont() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(len(got), tt.want) {
				t.Errorf("FindGeoZoneByPont() got = %v (%d), want %v", got,len(got), tt.want)
			}
		})
	}
}
