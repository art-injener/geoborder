package config

import (
	"github.com/spf13/viper"

	"github.com/X-Keeper/geoborder/pkg/logger"
)

const DebugLevel = "debug"

type Config struct {
	LogLevel   string `mapstructure:"LOG_LEVEL"`
	ServerPort int    `mapstructure:"PORT"`
	UseMocks   bool   `mapstructure:"USE_MOCK"`
	DBDevicesConfig
	GRPCConfig
	Log *logger.Logger
}

type DBConfig struct {
	Host     string
	Port     uint16
	NameDB   string
	User     string
	Password string
}

type GRPCConfig struct {
	Host string `mapstructure:"GRPC_HOST"`
	Port uint16 `mapstructure:"GRPC_PORT"`
}

type DBDevicesConfig struct {
	Host     string `mapstructure:"DEVICES_DB_HOST"`
	Port     uint16 `mapstructure:"DEVICES_DB_PORT"`
	NameDB   string `mapstructure:"DEVICES_DB_DATABASE"`
	User     string `mapstructure:"DEVICES_DB_USERNAME"`
	Password string `mapstructure:"DEVICES_DB_PASSWORD"`
}

func LoadConfig(path string) (*Config, error) {

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("USE_MOCK", &cfg.UseMocks); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("DEVICES_DB_HOST", &cfg.DBDevicesConfig.Host); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("DEVICES_DB_PORT", &cfg.DBDevicesConfig.Port); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("DEVICES_DB_USERNAME", &cfg.DBDevicesConfig.User); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("DEVICES_DB_PASSWORD", &cfg.DBDevicesConfig.Password); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("DEVICES_DB_DATABASE", &cfg.DBDevicesConfig.NameDB); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("GRPC_HOST", &cfg.GRPCConfig.Host); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("GRPS_PORT", &cfg.GRPCConfig.Port); err != nil {
		return nil, err
	}

	return &cfg, nil
}
