package main

import (
	"fmt"
	"github.com/X-Keeper/geoborder/internal/geofence"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"

	"github.com/pkg/errors"

	"github.com/X-Keeper/geoborder/internal/config"
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

	memoryGeoCache, err := geocache.NewMemoryCache(geoDB)

	if err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error create geocache"), cfg.Log)
		os.Exit(1)
	}
	if _, err := memoryGeoCache.Load(); err != nil {
		logger.LogError(errors.Wrap(err, "[MAIN] : error create geocache"), cfg.Log)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln(err)
	}

	server := grpc.NewServer()

	geoborderServer := geofence.NewGeoborderServer(memoryGeoCache)

	gf.RegisterGeofenceServiceServer(server, geoborderServer)

	log.Fatalln(server.Serve(listener))
}


