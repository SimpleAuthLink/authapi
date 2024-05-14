package mongo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/simpleauthlink/authapi/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	tokensCollection  = "tokens"
	secretsCollection = "secrets"
	appsCollection    = "apps"
)

type Config struct {
	MongoURI string
	Database string
}

type MongoDriver struct {
	ctx      context.Context
	cancel   context.CancelFunc
	config   Config
	client   *mongo.Client
	keysLock sync.RWMutex

	tokens *mongo.Collection
	apps   *mongo.Collection
}

func (md *MongoDriver) Init(config any) error {
	// validate config
	cfg, ok := config.(Config)
	if !ok {
		return db.ErrInvalidConfig
	}
	if cfg.Database == "" {
		return fmt.Errorf("%w: no database name provided", db.ErrInvalidConfig)
	}
	if cfg.MongoURI == "" {
		return fmt.Errorf("%w: no database url provided", db.ErrInvalidConfig)
	}
	// init the client options
	opts := options.Client()
	opts.ApplyURI(cfg.MongoURI)
	opts.SetMaxConnecting(200)
	timeout := time.Second * 10
	opts.ConnectTimeout = &timeout
	// connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return errors.Join(db.ErrOpenConn, err)
	}
	// check if the connection is available
	ctx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return errors.Join(db.ErrOpenConn, err)
	}
	// create the internal context
	md.ctx, md.cancel = context.WithCancel(context.Background())
	// set the client and config
	md.client = client
	md.config = cfg
	// instantiate the collections
	md.tokens = client.Database(cfg.Database).Collection(tokensCollection)
	md.apps = client.Database(cfg.Database).Collection(appsCollection)
	// create the indexes
	if err := md.createIndexes(); err != nil {
		return errors.Join(db.ErrOpenConn, err)
	}
	return nil
}

func (md *MongoDriver) Close() error {
	md.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := md.client.Disconnect(ctx); err != nil {
		return errors.Join(db.ErrCloseConn, err)
	}
	return nil
}

func (md *MongoDriver) createIndexes() error {
	ctx, cancel := context.WithTimeout(md.ctx, 20*time.Second)
	defer cancel()
	// create an index for app secrets
	if _, err := md.apps.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "secrets", Value: 1}}, // 1 for ascending order
		Options: nil,
	}); err != nil {
		return err
	}
	// create an index for token expiration
	if _, err := md.tokens.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expiration", Value: 1}},
		Options: nil,
	}); err != nil {
		return err
	}
	return nil
}
