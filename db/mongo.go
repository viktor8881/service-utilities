package db

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"time"
)

type MongoDB struct {
	Client *mongo.Client
	dbName string
	logger *zap.Logger
}

func NewMongoDb(ctx context.Context, cfg DatabaseConfig, logger *zap.Logger) (*MongoDB, func(), error) {
	dbName, err := extractDatabaseName(cfg.DSN)
	if err != nil {
		logger.Error("Failed to extract database name from DSN", zap.Error(err))
		return nil, nil, err
	}

	clientOpts := options.Client().ApplyURI(cfg.DSN).
		SetMaxPoolSize(uint64(cfg.SetMaxOpenConns)).             // SetMaxOpenConns аналог
		SetMinPoolSize(uint64(cfg.SetMaxIdleConns)).             // SetMaxIdleConns аналог
		SetMaxConnIdleTime(cfg.SetConnMaxLifetime * time.Minute) // SetConnMaxLifetime аналог

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil, nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Failed to ping MongoDB", zap.Error(err))
		return nil, nil, err
	}

	logger.Info("MongoDB connection established successfully")

	closeFunc := func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("Failed to close MongoDB connection", zap.Error(err))
		} else {
			logger.Info("MongoDB connection closed successfully")
		}
	}

	return &MongoDB{Client: client, dbName: dbName, logger: logger}, closeFunc, nil
}

func (db *MongoDB) Get(ctx context.Context, name string, collection string, dest interface{}, bsonFilter bson.M) error {
	fields := []zap.Field{zap.String("command", name), zap.String("collection", collection), zap.Any("args", bsonFilter)}
	db.logger.Info("mongo: Executing Get operation", fields...)

	coll := db.Client.Database(db.dbName).Collection(collection)
	err := coll.FindOne(ctx, bsonFilter).Decode(dest)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			db.logger.Error("mongo: Failed to execute Get operation", append(fields, zap.Error(err))...)
		}

		return err
	}

	db.logger.Info("mongo: Get operation executed successfully", fields...)
	return nil
}

func (db *MongoDB) FetchAll(ctx context.Context, name string, collection string, dest interface{}, bsonFilter bson.M) error {
	fields := []zap.Field{zap.String("command", name), zap.String("collection", collection), zap.Any("args", bsonFilter)}
	db.logger.Info("mongo: Executing FetchAll operation", fields...)

	coll := db.Client.Database(db.dbName).Collection(collection)
	cursor, err := coll.Find(ctx, bsonFilter)
	if err != nil {
		db.logger.Error("mongo: Failed to execute FetchAll operation", append(fields, zap.Error(err))...)
		return err
	}
	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			db.logger.Error("mongo: Failed to close cursor", zap.Error(cerr))
		}
	}()

	if err := cursor.All(ctx, dest); err != nil {
		db.logger.Error("mongo: Failed to decode FetchAll results", append(fields, zap.Error(err))...)
		return err
	}

	db.logger.Info("mongo: FetchAll operation executed successfully", fields...)
	return nil
}

func (db *MongoDB) Create(ctx context.Context, name string, collection string, document interface{}) (string, error) {
	fields := []zap.Field{zap.String("command", name), zap.String("database", db.dbName), zap.String("collection", collection), zap.Any("document", document)}
	db.logger.Info("mongo: Executing Create operation", fields...)

	coll := db.Client.Database(db.dbName).Collection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		db.logger.Error("mongo: Failed to execute Create operation", append(fields, zap.Error(err))...)
		return "", err
	}

	insertedID := result.InsertedID.(primitive.ObjectID).Hex()
	db.logger.Info("mongo: Create operation executed successfully", append(fields, zap.String("insertedID", insertedID))...)

	return insertedID, nil
}

func (db *MongoDB) Update(ctx context.Context, name string, collection string, filter, update bson.M) (int64, error) {
	fields := []zap.Field{zap.String("command", name), zap.String("database", db.dbName), zap.String("collection", collection), zap.Any("filter", filter), zap.Any("update", update)}
	db.logger.Info("mongo: Executing Update operation", fields...)

	coll := db.Client.Database(db.dbName).Collection(collection)
	result, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		db.logger.Error("mongo: Failed to execute Update operation", append(fields, zap.Error(err))...)
		return 0, err
	}

	db.logger.Info("mongo: Update operation executed successfully", fields...)
	return result.ModifiedCount, nil
}

func (db *MongoDB) Delete(ctx context.Context, name string, collection string, filter bson.M) (int64, error) {
	fields := []zap.Field{zap.String("command", name), zap.String("database", db.dbName), zap.String("collection", collection), zap.Any("filter", filter)}
	db.logger.Info("mongo: Executing Delete operation", fields...)

	coll := db.Client.Database(db.dbName).Collection(collection)
	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		db.logger.Error("mongo: Failed to execute Delete operation", append(fields, zap.Error(err))...)
		return 0, err
	}

	db.logger.Info("mongo: Delete operation executed successfully", fields...)
	return result.DeletedCount, nil
}

type TxMongoFunc func(sessCtx mongo.SessionContext) error

func (db *MongoDB) ExecuteTx(ctx context.Context, name string, txFunc TxMongoFunc) error {
	db.logger.Info("mongo: Executing operation in transaction", zap.String("command", name))

	session, err := db.Client.StartSession()
	if err != nil {
		db.logger.Error("mongo: Failed to start session", zap.Error(err))
		return err
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		if err := txFunc(sessCtx); err != nil {
			db.logger.Error("mongo: Transaction function failed", zap.Error(err))
			_ = session.AbortTransaction(sessCtx)
			return err
		}

		if err := session.CommitTransaction(sessCtx); err != nil {
			db.logger.Error("mongo: Failed to commit transaction", zap.Error(err))
			return err
		}

		db.logger.Info("mongo: Transaction executed successfully")
		return nil
	})

	return err
}

func extractDatabaseName(dsn string) (string, error) {
	uri, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}

	path := strings.TrimPrefix(uri.Path, "/")

	return path, nil
}
