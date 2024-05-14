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

type Token struct {
	Token      db.Token `bson:"_id"`
	Expiration int64    `bson:"expiration"`
}

func (md *MongoDriver) TokenExpiration(token db.Token) (time.Time, error) {
	var dbToken Token
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	if err := md.tokens.FindOne(ctx, bson.M{"_id": token}).Decode(&dbToken); err != nil {
		if err == mongo.ErrNoDocuments {
			return time.Time{}, db.ErrTokenNotFound
		}
		return time.Time{}, errors.Join(db.ErrGetToken, err)
	}
	return time.Unix(0, dbToken.Expiration), nil
}

func (md *MongoDriver) SetToken(token db.Token, expiration time.Time) error {
	md.keysLock.Lock()
	defer md.keysLock.Unlock()
	// set token in the database
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	dbToken := Token{
		Token:      token,
		Expiration: expiration.UnixNano(),
	}
	opts := options.Replace().SetUpsert(true)
	if _, err := md.tokens.ReplaceOne(ctx, bson.M{"_id": token}, dbToken, opts); err != nil {
		return errors.Join(db.ErrSetToken, err)
	}
	return nil
}

func (md *MongoDriver) DeleteToken(token db.Token) error {
	md.keysLock.Lock()
	defer md.keysLock.Unlock()
	// delete token from the database
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	if _, err := md.tokens.DeleteOne(ctx, bson.M{"_id": token}); err != nil {
		if err == mongo.ErrNoDocuments {
			return db.ErrTokenNotFound
		}
		return errors.Join(db.ErrDelToken, err)
	}
	return nil
}

func (md *MongoDriver) HasToken(tokenPrefix string) (db.Token, error) {
	// check if there is a token with the provided prefix in the database
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	var dbToken Token
	if err := md.tokens.FindOne(ctx, bson.M{"_id": bson.M{"$regex": "^" + tokenPrefix}}).Decode(&dbToken); err != nil {
		if err == mongo.ErrNoDocuments {
			return db.Token(""), db.ErrTokenNotFound
		}
		return db.Token(""), errors.Join(db.ErrGetToken, err)
	}
	return dbToken.Token, nil
}

func (md *MongoDriver) DeleteExpiredTokens() error {
	md.keysLock.Lock()
	defer md.keysLock.Unlock()
	// delete expired tokens from the database, filter by expiration time less
	// than now
	ctx, cancel := context.WithTimeout(md.ctx, 5*time.Second)
	defer cancel()
	dbNow := time.Now().UnixNano()
	if _, err := md.tokens.DeleteMany(ctx, bson.M{"expiration": bson.M{"$lt": dbNow}}); err != nil {
		return errors.Join(db.ErrDelToken, err)
	}
	return nil
}
