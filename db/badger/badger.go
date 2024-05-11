package badger

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/lucasmenendez/authapi/db"
)

const (
	tokenPrefix   = "token_"
	secretsPrefix = "secrets_"
	appPrefix     = "app_"
)

type BadgerDriver struct {
	path string
	db   *badger.DB
}

func (b *BadgerDriver) Init(config any) error {
	path, ok := config.(string)
	if !ok {
		return db.ErrInvalidConfig
	}
	b.path = path
	var err error
	if b.db, err = badger.Open(badger.DefaultOptions(path)); err != nil {
		return errors.Join(db.ErrOpenConn, err)
	}
	return nil
}

func (b *BadgerDriver) Close() error {
	if err := b.db.Close(); err != nil {
		return errors.Join(db.ErrCloseConn, err)
	}
	return nil
}

func (b *BadgerDriver) AppById(appId string) (*db.App, error) {
	app := &db.App{}
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(appPrefix + appId))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return db.ErrAppNotFound
			}
			return errors.Join(db.ErrGetApp, err)
		}
		if err := item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, app); err != nil {
				return errors.Join(db.ErrGetApp, err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("error getting app: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return app, nil
}

func (b *BadgerDriver) AppBySecret(secret string) (*db.App, string, error) {
	app := &db.App{}
	var appId string
	if err := b.db.View(func(txn *badger.Txn) error {
		// get app id from the database based on the app secret
		secretKey := []byte(secretsPrefix + secret)
		item, err := txn.Get(secretKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return db.ErrAppNotFound
			}
			return errors.Join(db.ErrGetApp, err)
		}
		if err := item.Value(func(val []byte) error {
			appId = string(val)
			return nil
		}); err != nil {
			return errors.Join(db.ErrGetApp, err)
		}
		bApp, err := txn.Get([]byte(appPrefix + appId))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return db.ErrAppNotFound
			}
			return errors.Join(db.ErrGetApp, err)
		}
		if err := bApp.Value(func(val []byte) error {
			if err := json.Unmarshal(val, app); err != nil {
				return errors.Join(db.ErrGetApp, err)
			}
			return nil
		}); err != nil {
			return errors.Join(db.ErrGetApp, err)
		}
		return nil
	}); err != nil {
		return nil, "", err
	}
	return app, appId, nil
}

func (b *BadgerDriver) SetApp(appId string, app *db.App) error {
	return b.db.Update(func(txn *badger.Txn) error {
		bApp, err := json.Marshal(app)
		if err != nil {
			return errors.Join(db.ErrSetApp, err)
		}
		if err := txn.Set([]byte(appPrefix+appId), bApp); err != nil {
			return errors.Join(db.ErrSetApp, err)
		}
		return nil
	})
}

func (b *BadgerDriver) DeleteApp(appId string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete([]byte(appPrefix + appId)); err != nil {
			return errors.Join(db.ErrDelApp, err)
		}
		return nil
	})
}

func (b *BadgerDriver) SetSecret(secret, appId string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte(secretsPrefix+secret), []byte(appId)); err != nil {
			return errors.Join(db.ErrSetSecret, err)
		}
		return nil
	})
}

func (b *BadgerDriver) DeleteSecret(secret string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete([]byte(secretsPrefix + secret)); err != nil {
			return errors.Join(db.ErrDelSecret, err)
		}
		return nil
	})
}

func (b *BadgerDriver) TokenExpiration(token db.Token) (time.Time, error) {
	var expiration time.Time
	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(tokenPrefix + token))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return db.ErrTokenNotFound
			}
			return errors.Join(db.ErrGetToken, err)
		}
		if err := item.Value(func(val []byte) error {
			iExpiration, err := strconv.Atoi(string(val))
			if err != nil {
				return errors.Join(db.ErrGetToken, err)
			}
			expiration = time.Unix(0, int64(iExpiration))
			return nil
		}); err != nil {
			return errors.Join(db.ErrGetToken, err)
		}
		return nil
	}); err != nil {
		return time.Time{}, err
	}
	return expiration, nil
}

func (b *BadgerDriver) SetToken(token db.Token, expiration time.Time) error {
	return b.db.Update(func(txn *badger.Txn) error {
		strExpiration := strconv.Itoa(int(expiration.UnixNano()))
		if err := txn.Set([]byte(tokenPrefix+token), []byte(strExpiration)); err != nil {
			return errors.Join(db.ErrSetToken, err)
		}
		return nil
	})
}

func (b *BadgerDriver) DeleteToken(token db.Token) error {
	return b.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete([]byte(tokenPrefix + token)); err != nil {
			return errors.Join(db.ErrDelToken, err)
		}
		return nil
	})
}

func (b *BadgerDriver) DeleteExpiredTokens() error {
	if err := b.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(tokenPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				iExpiration, err := strconv.Atoi(string(val))
				if err != nil {
					return errors.Join(db.ErrGetToken, err)
				}
				expiration := time.Unix(0, int64(iExpiration))
				if expiration.Before(time.Now()) {
					if err := txn.Delete(item.Key()); err != nil {
						return errors.Join(db.ErrDelToken, err)
					}
				}
				return nil
			}); err != nil {
				return errors.Join(db.ErrGetToken, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
