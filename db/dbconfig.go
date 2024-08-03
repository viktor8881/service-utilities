package db

import (
	"errors"
	"time"
)

type DatabaseConfig struct {
	DSN                string
	DBType             string // Тип базы данных, например "mongo", "mysql" и т.д.
	SetMaxOpenConns    int
	SetMaxIdleConns    int
	SetConnMaxLifetime time.Duration
}

func (cfg *DatabaseConfig) Validate() error {
	if cfg.DSN == "" {
		return errors.New("DSN is required")
	}
	if cfg.DBType == "" {
		return errors.New("DBType is required")
	}
	if cfg.SetMaxOpenConns <= 0 {
		return errors.New("SetMaxOpenConns must be greater than 0")
	}
	if cfg.SetMaxIdleConns < 0 {
		return errors.New("SetMaxIdleConns cannot be negative")
	}
	if cfg.SetConnMaxLifetime <= 0 {
		return errors.New("SetConnMaxLifetime must be greater than 0")
	}
	return nil
}
