package integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/storage/geocache"
	"github.com/X-Keeper/geoborder/internal/storage/postgres"
	gf "github.com/X-Keeper/geoborder/pkg/api/proto"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

type GeoborderSuite struct {
	suite.Suite
	ctx        context.Context
	clientConn *grpc.ClientConn
	client     gf.GeofenceServiceClient
	cache      *geocache.MemoryGeoCache
}

func TestGeoborderSuite(t *testing.T) {
	suite.Run(t, new(GeoborderSuite))
}

func (s *GeoborderSuite) SetupSuite() {
	// читаем конфигурационные настройки
	cfg, err := config.LoadConfig("../../configs")

	s.Require().NoError(err)

	cfg.Log = logger.NewConsole(cfg.LogLevel == config.DebugLevel)

	geoDB := postgres.NewGeoStorage(cfg)

	connect, err := geoDB.Connect(
		&config.DBConfig{
			Host:     cfg.DBDevicesConfig.Host,
			Port:     cfg.DBDevicesConfig.Port,
			NameDB:   cfg.DBDevicesConfig.NameDB,
			User:     cfg.DBDevicesConfig.User,
			Password: cfg.DBDevicesConfig.Password,
		})

	s.Require().NoError(err)
	s.Require().Equal(true, connect)

	s.cache, err = geocache.NewMemoryCache(geoDB, cfg.Log)

	s.Require().NoError(err)

	cnt, err := s.cache.Load()

	s.Require().Greater(cnt, 0)
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.clientConn, err = grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cfg.GRPCConfig.Port), grpc.WithInsecure())

	s.Require().NoError(err)
	s.Require().NotNil(s.clientConn)

	s.client = gf.NewGeofenceServiceClient(s.clientConn)

}

func (s *GeoborderSuite) TestGetGeofencesByUserId() {
	u := &gf.UserPoints{
		UserId:       0,
		WithDistance: false,
		Items: []*gf.Point{
			{
				PointId:   1,
				Latitude:  47.23571,
				Longitude: 39.70151,
				Accuracy:  0,
			},
		},
	}

	response, err := s.client.GetGeofencesByUserId(s.ctx, u)

	s.Require().NoError(err)
	s.Require().NotNil(response)

	actualGeofenceInfo := map[uint64]*gf.GeofenceInfo{
		366: {
			GeofenceId: 95,
			PolygonId:  366,
			Title:      "Южный федеральный округ",
			Distance:   0,
		},
		3904: {
			GeofenceId: 1,
			PolygonId:  3904,
			Title:      "Россия",
			Distance:   0,
		},
		3806: {
			GeofenceId: 37,
			PolygonId:  3806,
			Title:      "Ростовская область",
			Distance:   0,
		},
	}

	s.checkGeofenceInfo(u.UserId, actualGeofenceInfo, response)

}

func (s *GeoborderSuite) checkGeofenceInfo(userID uint64, actual map[uint64]*gf.GeofenceInfo, response *gf.Geofences) {
	s.Require().NotNil(response)

	s.Require().Equal(userID, response.UserId)

	s.Require().Equal(1, len(response.Geofence))

	info := response.Geofence[0].GeoInfo
	for i := 0; i < len(info); i++ {
		v, ok := actual[info[i].PolygonId]
		s.Require().True(ok)
		s.Require().Equal(v.GeofenceId, info[i].GeofenceId)
		s.Require().Equal(v.PolygonId, info[i].PolygonId)
		s.Require().Equal(v.Title, info[i].Title)
	}
}

func (s *GeoborderSuite) TestCheckGeofenceByPoint() {
	p := &gf.PointWithGeofence{
		Points: []*gf.Point{
			{
				PointId:   1,
				Latitude:  47.23571,
				Longitude: 39.70151,
				Accuracy:  0,
			},
			{
				PointId:   2,
				Latitude:  55.558741,
				Longitude: 37.378847,
				Accuracy:  0,
			},
		},
		GeofenceId: []uint64{221, 50},
	}

	actualGeofenceInfo := map[uint64]*gf.Geofence{
		7452: {
			PointId: 1,
			GeoInfo: []*gf.GeofenceInfo{{
				GeofenceId: 221,
				PolygonId:  7452,
				Title:      "Ростов",
				Distance:   0,
			}},
		},
		3732: {
			PointId: 2,
			GeoInfo: []*gf.GeofenceInfo{{
				GeofenceId: 50,
				PolygonId:  3732,
				Title:      "Москва",
				Distance:   0,
			}},
		},
	}

	response, err := s.client.CheckGeofenceByPoint(s.ctx, p)

	s.Require().NoError(err)
	s.Require().NotNil(response)

	s.checkGeofence(actualGeofenceInfo, response)
}

func (s *GeoborderSuite) checkGeofence(actual map[uint64]*gf.Geofence, response *gf.Geofences) {
	s.Require().NotNil(response)

	s.Require().Equal(len(actual), len(response.Geofence))

	for _, geofence := range response.Geofence {
		for i := 0; i < len(geofence.GeoInfo); i++ {
			v, ok := actual[geofence.GeoInfo[i].PolygonId]
			s.Require().True(ok)
			s.Require().Equal(1, len(v.GeoInfo))
			s.Require().Equal(v.GeoInfo[0].GeofenceId, geofence.GeoInfo[i].GeofenceId)
			s.Require().Equal(v.GeoInfo[0].PolygonId, geofence.GeoInfo[i].PolygonId)
			s.Require().Equal(v.GeoInfo[0].Title, geofence.GeoInfo[i].Title)
			s.Require().Equal(v.GeoInfo[0].Distance, geofence.GeoInfo[i].Distance)
		}
	}
}

func (s *GeoborderSuite) TestGetDistanceToGeofence() {
	p := &gf.Points{
		Points: []*gf.Point{
			{
				PointId:   1,
				Latitude:  47.23571,
				Longitude: 39.70151,
				Accuracy:  0,
			},
			{
				PointId:   2,
				Latitude:  55.558741,
				Longitude: 37.378847,
				Accuracy:  0,
			},
		},
	}

	response, err := s.client.GetDistanceToGeofence(s.ctx, p)

	s.Require().NoError(err)
	s.Require().NotNil(response)

	for i := 0; i < len(response.Geofence); i++ {
		s.Require().True(len(response.Geofence[i].GeoInfo) > 0)
	}
}
