package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/X-Keeper/geoborder/internal/config"
	"github.com/X-Keeper/geoborder/internal/geofence"
	"github.com/X-Keeper/geoborder/internal/storage/geocache"
	"github.com/X-Keeper/geoborder/internal/storage/postgres"
	gf "github.com/X-Keeper/geoborder/pkg/api/proto"
	"github.com/X-Keeper/geoborder/pkg/logger"
)

func main() {
	// читаем конфигурационные настройки
	cfg, err := config.LoadConfig("configs")
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

	memoryGeoCache, err := geocache.NewMemoryCache(geoDB, cfg.Log)

	if err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error create geocache"), cfg.Log)
		os.Exit(1)
	}
	if _, err = memoryGeoCache.Load(); err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error load geocache"), cfg.Log)
		os.Exit(1)
	}

	const defaultTimeout = 5
	ticker := time.NewTicker(defaultTimeout * time.Second)
	done := make(chan bool)

	dbUpdater(done, ticker, memoryGeoCache, cfg)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCConfig.Port))
	if err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error listen tcp"), cfg.Log)
		os.Exit(1)
	}

	logger.LogDebug(fmt.Sprintf("[MAIN]::Start listen  %s", listener.Addr()), cfg.Log)

	server := grpc.NewServer()

	geoborderServer := geofence.NewGeoborderServer(memoryGeoCache)

	gf.RegisterGeofenceServiceServer(server, geoborderServer)

	if err := server.Serve(listener); err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error start server"), cfg.Log)
		os.Exit(1)
	}
	ticker.Stop()
	done <- true
}

func dbUpdater(done chan bool, ticker *time.Ticker, memoryGeoCache *geocache.MemoryGeoCache, cfg *config.Config) {
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_, err := memoryGeoCache.Update()
				if err != nil {
					logger.LogError(errors.Wrap(err, "[MAIN] : error update cache"), cfg.Log)

					return
				}
			}
		}
	}()
}
