package db

import (
	"context"
	"database/sql"

	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DatabaseConfig struct {
	DSN    string
	DBType string // Тип базы данных, например "postgres", "mysql" и т.д.
}

type DB struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewDb(ctx context.Context, cfg DatabaseConfig, logger *zap.Logger) (*DB, error) {
	db, err := sqlx.ConnectContext(ctx, cfg.DBType, cfg.DSN)
	if err != nil {
		logger.Error("Failed to connect to database", zap.String("database_type", cfg.DBType), zap.Error(err))
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.String("database_type", cfg.DBType), zap.Error(err))
		return nil, err
	}

	logger.Info("Database connection established successfully", zap.String("database_type", cfg.DBType))

	return &DB{db: db, logger: logger}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Get(ctx context.Context, name string, query string, dest interface{}, args ...interface{}) error {
	db.logger.Info(fmt.Sprintf("sql: %s", name), zap.String("query", query), zap.Any("args", args))

	err := db.db.GetContext(ctx, dest, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			db.logger.Error(fmt.Sprintf("sql: Failed to execute query %s", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		}

		return err
	}

	return nil
}

func (db *DB) FetchAll(ctx context.Context, name string, query string, dest interface{}, args ...interface{}) error {
	db.logger.Info(fmt.Sprintf("sql: %s", name), zap.String("query", query), zap.Any("args", args))

	err := db.db.SelectContext(ctx, dest, query, args...)
	if err != nil {
		db.logger.Error(fmt.Sprintf("sql: Failed to execute query %s", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		return err
	}

	return nil
}

func (db *DB) Create(ctx context.Context, name string, query string, args ...interface{}) (int64, error) {
	db.logger.Info("sql: Executing SQL command", zap.String("command", name), zap.String("query", query), zap.Any("args", args))

	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		db.logger.Error("sql: Failed to execute command", zap.String("command", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		return 0, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		db.logger.Error("sql: Failed to get last insert newID", zap.String("command", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		return 0, err

	}

	return newID, nil
}

func (db *DB) Update(ctx context.Context, name string, query string, args ...interface{}) (int64, error) {
	db.logger.Info("sql: Executing SQL command", zap.String("command", name), zap.String("query", query), zap.Any("args", args))

	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		db.logger.Error("sql: Failed to execute command", zap.String("command", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		db.logger.Error("sql: Failed to get rows affected", zap.String("command", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		return 0, err
	}

	return rowsAffected, nil
}

func (db *DB) Delete(ctx context.Context, name string, query string, args ...interface{}) (int64, error) {
	db.logger.Info("sql: Executing SQL command", zap.String("command", name), zap.String("query", query), zap.Any("args", args))

	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		db.logger.Error("sql: Failed to execute command", zap.String("command", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		db.logger.Error("sql: Failed to get rows affected", zap.String("command", name), zap.String("query", query), zap.Error(err), zap.Any("args", args))
		return 0, err
	}

	return rowsAffected, nil
}

type TxFunc func(tx *sql.Tx) error

func (db *DB) ExecuteTx(ctx context.Context, name string, txFunc TxFunc) error {
	db.logger.Info("sql: Starting transaction", zap.String("transaction", name))

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		db.logger.Error("sql: Failed to begin transaction", zap.String("transaction", name), zap.Error(err))
		return err
	}

	if err := txFunc(tx); err != nil {
		db.logger.Error("sql: Failed to execute transaction function", zap.String("transaction", name), zap.Error(err))
		errRollback := tx.Rollback()
		if errRollback != nil {
			db.logger.Error("sql: Failed to rollback transaction", zap.String("transaction", name), zap.Error(err))
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		db.logger.Error("sql: Failed to commit transaction", zap.String("transaction", name), zap.Error(err))
		errRollback := tx.Rollback()
		if errRollback != nil {
			db.logger.Error("sql: Failed to rollback transaction", zap.String("transaction", name), zap.Error(err))
		}

		return err
	}

	db.logger.Info("sql: Transaction executed successfully", zap.String("transaction", name))
	return nil
}
