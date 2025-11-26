package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

type FeedConfig struct {
	URL string
}

type SchedulerConfig struct {
	Interval time.Duration
}

type LoggingConfig struct {
	Level string
}

type AppConfig struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Feed      FeedConfig
	Scheduler SchedulerConfig
	Logging   LoggingConfig
}

func Load() (*AppConfig, error) {
	_ = godotenv.Load()

	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "5000")
	viper.SetDefault("DATABASE_DSN", "postgres://trading_user:trading_password@127.0.0.1:5432/trading_db?sslmode=disable")
	viper.SetDefault("CALENDAR_FEED_URL", "https://nfs.faireconomy.media/ff_calendar_thisweek.json")
	viper.SetDefault("SCHEDULER_INTERVAL", "1h")
	viper.SetDefault("LOG_LEVEL", "info")

	interval, err := time.ParseDuration(viper.GetString("SCHEDULER_INTERVAL"))
	if err != nil {
		return nil, fmt.Errorf("invalid scheduler interval: %w", err)
	}

	cfg := &AppConfig{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			DSN: viper.GetString("DATABASE_DSN"),
		},
		Feed: FeedConfig{
			URL: viper.GetString("CALENDAR_FEED_URL"),
		},
		Scheduler: SchedulerConfig{
			Interval: interval,
		},
		Logging: LoggingConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}

	return cfg, nil
}
