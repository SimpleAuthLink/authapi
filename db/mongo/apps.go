package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/simpleauthlink/authapi/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type App struct {
	ID              string `bson:"_id"`
	Name            string `bson:"name"`
	AdminEmail      string `bson:"admin_email"`
	SessionDuration int64  `bson:"session_duration"`
	Callback        string `bson:"callback"`
	Secret          string `bson:"secret"`
}

func (md *MongoDriver) AppById(appId string) (*db.App, error) {
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	// get app from the database based on the app id
	var app App
	if err := md.apps.FindOne(ctx, bson.M{"_id": appId}).Decode(&app); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, db.ErrAppNotFound
		}
		return nil, errors.Join(db.ErrGetApp, err)
	}
	// return app
	return &db.App{
		Name:            app.Name,
		AdminEmail:      app.AdminEmail,
		SessionDuration: app.SessionDuration,
		Callback:        app.Callback,
	}, nil
}

func (md *MongoDriver) AppBySecret(secret string) (*db.App, string, error) {
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	// get app from the database based on the app secret
	var app App
	if err := md.apps.FindOne(ctx, bson.M{"secret": secret}).Decode(&app); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, "", db.ErrAppNotFound
		}
		return nil, "", errors.Join(db.ErrGetApp, err)
	}
	// return app and app id
	return &db.App{
		Name:            app.Name,
		AdminEmail:      app.AdminEmail,
		SessionDuration: app.SessionDuration,
		Callback:        app.Callback,
	}, app.ID, nil
}

func (md *MongoDriver) SetApp(appId string, app *db.App) error {
	md.keysLock.Lock()
	defer md.keysLock.Unlock()
	// create or update app in the database
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	dbApp := App{
		ID:              appId,
		Name:            app.Name,
		AdminEmail:      app.AdminEmail,
		SessionDuration: app.SessionDuration,
		Callback:        app.Callback,
	}
	opts := options.Replace().SetUpsert(true)
	if _, err := md.apps.ReplaceOne(ctx, bson.M{"_id": appId}, dbApp, opts); err != nil {
		return errors.Join(db.ErrSetApp, err)
	}
	return nil
}

func (md *MongoDriver) DeleteApp(appId string) error {
	md.keysLock.Lock()
	defer md.keysLock.Unlock()
	// delete secret from the database by the app id
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	if _, err := md.apps.DeleteOne(ctx, bson.M{"_id": appId}); err != nil {
		if err == mongo.ErrNoDocuments {
			return db.ErrAppNotFound
		}
		return errors.Join(db.ErrDelApp, err)
	}
	return nil
}

func (md *MongoDriver) SetSecret(secret, appId string) error {
	md.keysLock.Lock()
	defer md.keysLock.Unlock()
	// set secret to app in the database by the app id
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	if _, err := md.apps.UpdateOne(ctx, bson.M{"_id": appId}, bson.M{"$set": bson.M{"secret": secret}}); err != nil {
		if err == mongo.ErrNoDocuments {
			return db.ErrAppNotFound
		}
		return errors.Join(db.ErrSetSecret, err)
	}
	return nil
}

func (md *MongoDriver) DeleteSecret(secret string) error {
	md.keysLock.Lock()
	defer md.keysLock.Unlock()
	// delete secret of the app from the database
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	if _, err := md.apps.UpdateOne(ctx, bson.M{"secret": secret}, bson.M{"$unset": bson.M{"secret": ""}}); err != nil {
		if err == mongo.ErrNoDocuments {
			return db.ErrAppNotFound
		}
		return errors.Join(db.ErrDelSecret, err)
	}
	return nil
}
