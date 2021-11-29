package geocache

import (
	"fmt"
	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/storage/postgres"
	"github.com/pkg/errors"
	"os"
	"testing"

	"github.com/X-Keeper/geoborder/pkg/logger"
	"github.com/paulmach/orb"
)

var cache *MemoryGeoCache

func init() {

	// читаем конфигурационные настройки
	cfg, err := config.LoadConfig("../../../configs")
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

	cache, err = NewMemoryCache(geoDB)

	if err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error create geocache"), cfg.Log)
		os.Exit(1)
	}
	if _, err := cache.Load(); err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error create geocache"), cfg.Log)
		os.Exit(1)
	}
}

func TestMemoryGeoCache_FindGeoZoneByPont(t *testing.T) {

	type args struct {
		point orb.Point
		userID uint64
	}

	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "RostovOnDone",
			args:    args{orb.Point{39.70151, 47.23571},
				22217},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cache.FindGeofenceByPoint(tt.args.point, &tt.args.userID, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindGeofenceByPoint() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if len(got) != tt.want {
				t.Errorf("FindGeofenceByPoint() got = %v (%d), want %v", got, len(got), tt.want)
			}
		})
	}
}

func TestMemoryGeoCache_CheckGeofenceByPoint(t *testing.T) {

	type args struct {
		point      orb.Point
		geofenceId []uint64
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "check RostovOnDone",
			args:    args{
				point:      orb.Point{39.70151, 47.23571},
				geofenceId: []uint64{50,221},
			},

			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			
			got, err := cache.CheckGeofenceByPoint(tt.args.point, tt.args.geofenceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckGeofenceByPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("CheckGeofenceByPoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryGeoCache_GetDistanceToGeofence(t *testing.T) {
	type args struct {
		point orb.Point
		userID uint64
	}

	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "RostovOnDone",
			args:    args{orb.Point{39.70151, 47.23571},
				22217},
			want:   9,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cache.GetDistanceToGeofence(tt.args.point)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindGeofenceByPoint() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if len(got) < tt.want {
				t.Errorf("FindGeofenceByPoint() got = %v (%d), want %v", got, len(got), tt.want)
			}
		})
	}
}