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

type AppConfig struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Feed      FeedConfig
	Scheduler SchedulerConfig
}

func Load() (*AppConfig, error) {
	_ = godotenv.Load()

	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "3000")
	viper.SetDefault("DATABASE_DSN", "data/forex.db")
	viper.SetDefault("CALENDAR_FEED_URL", "https://nfs.faireconomy.media/ff_calendar_thisweek.json")
	viper.SetDefault("SCHEDULER_INTERVAL", "1h")

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
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}

	return cfg, nil
}
