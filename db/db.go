package db

import (
	"context"
	"database/sql"
	"time"

	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DatabaseConfig struct {
	DSN                string
	DBType             string // Тип базы данных, например "postgres", "mysql" и т.д.
	SetMaxOpenConns    int
	SetMaxIdleConns    int
	SetConnMaxLifetime time.Duration
}

type DB struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewDb(ctx context.Context, cfg DatabaseConfig, logger *zap.Logger) (*DB, func(), error) {
	db, err := sqlx.ConnectContext(ctx, cfg.DBType, cfg.DSN)
	if err != nil {
		logger.Error("db: Failed to connect to database", zap.String("database_type", cfg.DBType), zap.Error(err))
		return nil, nil, err
	}

	db.SetMaxOpenConns(cfg.SetMaxOpenConns)
	db.SetMaxIdleConns(cfg.SetMaxIdleConns)
	db.SetConnMaxLifetime(cfg.SetConnMaxLifetime)

	if err = db.Ping(); err != nil {
		logger.Error("db: Failed to ping database", zap.String("database_type", cfg.DBType), zap.Error(err))
		return nil, nil, err
	}

	logger.Info("db: Database connection established successfully", zap.String("database_type", cfg.DBType))

	closeFunc := func() {
		if err := db.Close(); err != nil {
			logger.Error("db: close db connection error", zap.Error(err))
		} else {
			logger.Info("db: close db connection success")
		}
	}

	return &DB{db: db, logger: logger}, closeFunc, nil
}

func (db *DB) Get(ctx context.Context, name string, query string, dest interface{}, args ...interface{}) error {
	fields := []zap.Field{zap.String("command", name), zap.String("query", query), zap.Any("args", args)}
	db.logger.Info("db: execute sql:", fields...)

	err := db.db.GetContext(ctx, dest, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fields = append(fields, zap.Error(err))
			db.logger.Error("db: failed sql", fields...)
		}

		return err
	}

	return nil
}

func (db *DB) FetchAll(ctx context.Context, name string, query string, dest interface{}, args ...interface{}) error {
	fields := []zap.Field{zap.String("command", name), zap.String("query", query), zap.Any("args", args)}
	db.logger.Info("db: execute sql:", fields...)

	err := db.db.SelectContext(ctx, dest, query, args...)
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql:", fields...)
		return err
	}

	return nil
}

func (db *DB) Create(ctx context.Context, name string, query string, args ...interface{}) (int64, error) {
	fields := []zap.Field{zap.String("command", name), zap.String("query", query), zap.Any("args", args)}
	db.logger.Info("db: execute sql:", fields...)

	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql:", fields...)
		return 0, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql: Failed to get last insert newID", fields...)
		return 0, err

	}

	return newID, nil
}

func (db *DB) Update(ctx context.Context, name string, query string, args ...interface{}) (int64, error) {
	fields := []zap.Field{zap.String("command", name), zap.String("query", query), zap.Any("args", args)}
	db.logger.Info("db: execute sql:", fields...)

	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql:", fields...)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql: Failed to get rows affected", fields...)
		return 0, err
	}

	return rowsAffected, nil
}

func (db *DB) Delete(ctx context.Context, name string, query string, args ...interface{}) (int64, error) {
	fields := []zap.Field{zap.String("command", name), zap.String("query", query), zap.Any("args", args)}
	db.logger.Info("db: execute sql:", fields...)

	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql:", fields...)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql: Failed to get rows affected", fields...)
		return 0, err
	}

	return rowsAffected, nil
}

type TxFunc func(tx *sql.Tx) error

func (db *DB) ExecuteTx(ctx context.Context, name string, txFunc TxFunc) error {
	fields := []zap.Field{zap.String("command", name)}
	db.logger.Info("db: execute sql in transaction:", fields...)

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql: Failed to begin transaction", fields...)
		return err
	}

	if err := txFunc(tx); err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql: Failed to execute transaction function", fields...)
		errRollback := tx.Rollback()
		if errRollback != nil {
			db.logger.Error("db: failed sql: Failed to rollback transaction", zap.String("command", name), zap.Error(err))
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("db: failed sql: Failed to commit transaction function", fields...)
		errRollback := tx.Rollback()
		if errRollback != nil {
			db.logger.Error("db: failed sql: Failed to rollback transaction", zap.String("command", name), zap.Error(err))
		}

		return err
	}

	db.logger.Info("db: success sql: Transaction executed successfully", fields...)
	return nil
}
