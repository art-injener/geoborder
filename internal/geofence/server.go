package geofence

import (
	"context"
	"fmt"

	"github.com/paulmach/orb"

	"github.com/X-Keeper/geoborder/internal/storage"
	gf "github.com/X-Keeper/geoborder/pkg/api/proto"
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

		fmt.Println(err)
		fmt.Println(geofences)
		if err != nil {
			return nil, err
		}

		geoInfo := make([]*gf.GeofenceInfo, 0, len(geofences))

		for j := 0; j < len(geofences); j++ {
			geoInfo = append(geoInfo, &gf.GeofenceInfo{
				GeofenceId: geofences[j].GeofenceID,
				PolygonId:  geofences[j].PolygonID,
				Title:      geofences[j].Title,
				Distance:   geofences[j].Distance,
			})
		}
		grpcResponse = append(grpcResponse, &gf.Geofence{
			PointId: points.Items[i].PointId,
			GeoInfo: geoInfo,
		})
	}

	fmt.Println(grpcResponse)

	return &gf.Geofences{
		UserId:   points.UserId,
		Geofence: grpcResponse,
		Status:   gf.Status_OK,
		Error:    "",
	}, nil
}

func (s *GeoborderServer) CheckGeofenceByPoint(_ context.Context, req *gf.PointWithGeofence) (*gf.Geofences, error) {
	grpcResponse := make([]*gf.Geofence, 0, 1)

	for i := 0; i < len(req.Points); i++ {
		geofences, err := s.geoCache.CheckGeofenceByPoint(
			orb.Point{
				req.Points[i].Longitude,
				req.Points[i].Latitude},
			req.GeofenceId,
		)

		if err != nil {
			return nil, err
		}

		geoInfo := make([]*gf.GeofenceInfo, 0, len(geofences))

		for j := 0; j < len(geofences); j++ {
			geoInfo = append(geoInfo, &gf.GeofenceInfo{
				GeofenceId: geofences[j].GeofenceID,
				PolygonId:  geofences[j].PolygonID,
				Title:      geofences[j].Title,
				Distance:   geofences[j].Distance,
			})
		}
		grpcResponse = append(grpcResponse, &gf.Geofence{
			PointId: req.Points[i].PointId,
			GeoInfo: geoInfo,
		})
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
		geofence, err := s.geoCache.GetDistanceToGeofence(orb.Point{request.Points[i].Longitude, request.Points[i].Latitude})
		if err != nil {
			return nil, err
		}
		geoInfo := make([]*gf.GeofenceInfo, 0, len(geofence))

		for j := 0; j < len(geofence); j++ {
			geoInfo = append(geoInfo, &gf.GeofenceInfo{
				GeofenceId: geofence[j].GeofenceID,
				PolygonId:  geofence[j].PolygonID,
				Title:      geofence[j].Title,
				Distance:   geofence[j].Distance,
			})
		}
		grpcResponse = append(grpcResponse, &gf.Geofence{
			PointId: request.Points[i].PointId,
			GeoInfo: geoInfo,
		})
	}

	return &gf.Geofences{
		UserId:   0,
		Geofence: grpcResponse,
		Status:   gf.Status_OK,
		Error:    "",
	}, nil
}
