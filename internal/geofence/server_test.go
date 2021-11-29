package geofence

import (
	"context"
	"fmt"
	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/storage/geocache"
	"github.com/X-Keeper/geoborder/internal/storage/postgres"
	gf "github.com/X-Keeper/geoborder/pkg/api/proto"
	"github.com/X-Keeper/geoborder/pkg/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os"
	"testing"
)

var cache *geocache.MemoryGeoCache

func init() {
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

	cache, err = geocache.NewMemoryCache(geoDB)

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

// Тест интеграционный, нужен дамп геозон
func TestGeoborderServer_GetGeofencesByUserId(t *testing.T) {
	tests := []struct {
		name    string
		req     *gf.UserPoints
		res     *gf.Geofences
		errCode gf.Status
		errMsg  string
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
			nil,
			gf.Status_OK,
			fmt.Sprintf("cannot deposit %v", -1.11),
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

			response, err := client.GetGeofencesByUserId(ctx, tt.req)

			if response != nil {
				if response.GetStatus() != tt.errCode {
					t.Error("response: expected", tt.res.GetStatus(), "received", response.GetStatus())
				}
			}

			if err != nil {
				if er, ok := status.FromError(err); ok {
					//if er.Code() != tt.errCode {
					//	t.Error("error code: expected", codes.InvalidArgument, "received", er.Code())
					//}
					if er.Message() != tt.errMsg {
						t.Error("error message: expected", tt.errMsg, "received", er.Message())
					}
				}
			}
		})
	}
}

func TestGeoborderServer_CheckGeofenceByPoint(t *testing.T) {
	tests := []struct {
		name    string
		req     *gf.PointWithGeofence
		res     *gf.Geofences
		errCode gf.Status
		errMsg  string
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
				GeofenceId:           []uint64{221,50},
			},
			nil,
			gf.Status_OK,
			fmt.Sprintf("cannot deposit %v", -1.11),
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

			response, err := client.CheckGeofenceByPoint(ctx, tt.req)

			if response != nil {
				if response.GetStatus() != tt.errCode {
					t.Error("response: expected", tt.res.GetStatus(), "received", response.GetStatus())
				}
			}

			if err != nil {
				if er, ok := status.FromError(err); ok {
					//if er.Code() != tt.errCode {
					//	t.Error("error code: expected", codes.InvalidArgument, "received", er.Code())
					//}
					if er.Message() != tt.errMsg {
						t.Error("error message: expected", tt.errMsg, "received", er.Message())
					}
				}
			}
		})
	}
}