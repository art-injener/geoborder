package geofence

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/storage/geocache"
	"github.com/X-Keeper/geoborder/internal/storage/postgres"
	gf "github.com/X-Keeper/geoborder/pkg/api/proto"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

var cache *geocache.MemoryGeoCache //nolint:gochecknoglobals // используется только для тестов

func init() { //nolint:gochecknoinits // инициализация тестового окружения
	// читаем конфигурационные настройки
	cfg, err := config.LoadConfig("../../configs")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

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

	if err != nil || !connect {
		logger.LogError(errors.Wrap(err, "[MAIN] : error connect to devicesDb"), cfg.Log)
		os.Exit(1)
	}

	cache, err = geocache.NewMemoryCache(geoDB, cfg.Log)

	if err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error create geocache"), cfg.Log)
		os.Exit(1)
	}
	if _, err := cache.Load(); err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error create geocache"), cfg.Log)
		os.Exit(1)
	}
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	geoborderServer := NewGeoborderServer(cache)

	gf.RegisterGeofenceServiceServer(server, geoborderServer)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

// Тест интеграционный, нужен дамп геозон.
//nolint:funlen // тест
func TestGeoborderServer_GetGeofencesByUserId(t *testing.T) {
	tests := []struct {
		name    string
		req     *gf.UserPoints
		res     *gf.Geofences
		errCode gf.Status
		wantErr error
	}{
		{
			"valid request",
			&gf.UserPoints{
				UserId: 0,
				Items: []*gf.Point{
					{
						PointId:   1,
						Latitude:  47.23571,
						Longitude: 39.70151,
						Accuracy:  0,
					},
				},
			},
			&gf.Geofences{
				UserId: 1,
				Geofence: []*gf.Geofence{
					{PointId: 1},
				},
			},
			gf.Status_OK,
			nil,
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := gf.NewGeofenceServiceClient(conn)


	for _, tt := range tests { //nolint:dupl // тесты
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.GetGeofencesByUserId(ctx, tt.req)

			if response != nil {
				if response.GetStatus() != tt.errCode {
					t.Error("response: expected", tt.res.GetStatus(), "received", response.GetStatus())
				}
			}

			if response.Geofence == nil {
				t.Error("empty response")
			}

			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Error("error message: expected", tt.wantErr, "received", err)
				}
			}

			for i := 0; i < len(tt.req.Items); i++ {
				if response.Geofence[i].PointId != tt.res.Geofence[i].PointId {
					t.Errorf("GetDistanceToGeofence() got = %v, want %v", response, tt.res)
				}

				if response.Geofence[i].GeoInfo == nil {
					t.Errorf("GetDistanceToGeofence()  empty GeoInfo")
				}
			}
		})
	}
}

//nolint:funlen // тест
func TestGeoborderServer_CheckGeofenceByPoint(t *testing.T) {
	tests := []struct {
		name    string
		req     *gf.PointWithGeofence
		res     *gf.Geofences
		errCode gf.Status
		wantErr error
	}{
		{
			"valid request",
			&gf.PointWithGeofence{
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
			},
			&gf.Geofences{
				UserId: 0,
				Geofence: []*gf.Geofence{
					{PointId: 1},
					{PointId: 2},
				},
				Status: gf.Status_OK,
			},
			gf.Status_OK,
			nil,
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := gf.NewGeofenceServiceClient(conn)

	for _, tt := range tests { //nolint:dupl // тест
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.CheckGeofenceByPoint(ctx, tt.req)

			if response != nil {
				if response.GetStatus() != tt.errCode {
					t.Error("response: expected", tt.res.GetStatus(), "received", response.GetStatus())
				}
			}

			if response.Geofence == nil {
				t.Error("empty response")
			}

			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Error("error message: expected", tt.wantErr, "received", err)
				}
			}

			for i := 0; i < len(tt.req.Points); i++ {
				if response.Geofence[i].PointId != tt.res.Geofence[i].PointId {
					t.Errorf("GetDistanceToGeofence() got = %v, want %v", response, tt.res)
				}

				if response.Geofence[i].GeoInfo == nil {
					t.Errorf("GetDistanceToGeofence()  empty GeoInfo")
				}
			}
		})
	}
}

func TestGeoborderServer_GetDistanceToGeofence(t *testing.T) {
	tests := []struct {
		name    string
		req     *gf.Points
		res     *gf.Geofences
		errCode gf.Status
		wantErr error
	}{
		{
			"valid request",
			&gf.Points{
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
			},
			&gf.Geofences{
				UserId: 0,
				Geofence: []*gf.Geofence{
					{PointId: 1},
					{PointId: 2},
				},
				Status: gf.Status_OK,
			},
			gf.Status_OK,
			nil,
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := gf.NewGeofenceServiceClient(conn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := client.GetDistanceToGeofence(ctx, tt.req)

			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Error("error message: expected", tt.wantErr, "received", err)
				}
			}

			if got.Geofence == nil {
				t.Errorf("GetDistanceToGeofence()  empty response")
			}

			for i := 0; i < len(tt.req.Points); i++ {
				if got.Geofence[i].PointId != tt.res.Geofence[i].PointId {
					t.Errorf("GetDistanceToGeofence() got = %v, want %v", got, tt.res)
				}

				if got.Geofence[i].GeoInfo == nil {
					t.Errorf("GetDistanceToGeofence()  empty GeoInfo")
				}
			}

		})
	}
}
