package geofence

import (
	"context"
	"github.com/X-Keeper/geoborder/internal/storage"
	gf "github.com/X-Keeper/geoborder/pkg/api/proto"
	"github.com/paulmach/orb"
)

type GeoborderServer struct {
	gf.UnimplementedGeofenceServiceServer
	geoCache storage.MemoryGeoCache
}

func NewGeoborderServer(geoCache storage.MemoryGeoCache) *GeoborderServer {
	return &GeoborderServer{
		geoCache: geoCache,
	}
}

func (s *GeoborderServer) GetGeofencesByUserId(ctx context.Context, points *gf.UserPoints) (*gf.Geofences, error) {
	grpcResponse := make([]*gf.Geofence, 0, 1)

	for i := 0; i < len(points.Items); i++ {
		geofences, err := s.geoCache.FindGeofenceByPoint(
			orb.Point{
				points.Items[i].Longitude,
				points.Items[i].Latitude},
			&points.UserId,
			points.WithDistance)

		if err != nil {
			return nil, err
		}

		geoInfo := make([]*gf.GeofenceInfo, 0, len(geofences))

		for i := 0; i < len(geofences); i++ {
			geoInfo = append(geoInfo, &gf.GeofenceInfo{
				GeofenceId: geofences[i].GeofenceID,
				PolygonId:  geofences[i].PolygonID,
				Title:      geofences[i].Title,
				Distance:   geofences[i].Distance,
			})
		}
		grpcResponse = append(grpcResponse, &gf.Geofence{
			PointId: points.Items[i].PointId,
			GeoInfo: geoInfo,
		})
	}

	return &gf.Geofences{
		UserId:   points.UserId,
		Geofence: grpcResponse,
		Status:   gf.Status_OK,
		Error:    "",
	}, nil
}

func (s *GeoborderServer) CheckGeofenceByPoint(ctx context.Context, request *gf.PointWithGeofence) (*gf.Geofences, error) {
	grpcResponse := make([]*gf.Geofence, 0, 1)

	for i := 0; i < len(request.Points); i++ {

		s.geoCache.CheckGeofenceByPoint(
			orb.Point{
				request.Points[i].Longitude,
				request.Points[i].Latitude},
			request.GeofenceId,
		)
	}

	return &gf.Geofences{
		UserId:   0,
		Geofence: grpcResponse,
		Status:   gf.Status_OK,
		Error:    "",
	}, nil
}

func (s *GeoborderServer) GetDistanceToGeofence(_ context.Context, request *gf.Points) (*gf.Geofences, error) {
	grpcResponse := make([]*gf.Geofence, 0, 1)

	for i := 0; i < len(request.Points); i++ {
		s.geoCache.GetDistanceToGeofence(orb.Point{request.Points[i].Longitude,	request.Points[i].Latitude})
	}

	return &gf.Geofences{
		UserId:   0,
		Geofence: grpcResponse,
		Status:   gf.Status_OK,
		Error:    "",
	}, nil
}