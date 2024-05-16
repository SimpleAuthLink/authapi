package db

import (
	"strings"
	"sync"
	"time"
)

type TempDriver struct {
	apps        map[string]App
	secretToApp map[string]string
	tokens      map[Token]int64
	lock        sync.RWMutex
}

func (tdb *TempDriver) Init(_ any) error {
	tdb.apps = make(map[string]App)
	tdb.secretToApp = make(map[string]string)
	tdb.tokens = make(map[Token]int64)
	return nil
}

func (tdb *TempDriver) Close() error {
	return nil
}

func (tdb *TempDriver) AppById(appId string) (*App, error) {
	tdb.lock.RLock()
	defer tdb.lock.RUnlock()
	app, ok := tdb.apps[appId]
	if !ok {
		return nil, ErrAppNotFound
	}
	return &app, nil
}

func (tdb *TempDriver) AppBySecret(secret string) (*App, string, error) {
	tdb.lock.RLock()
	defer tdb.lock.RUnlock()
	appId, ok := tdb.secretToApp[secret]
	if !ok {
		return nil, "", ErrAppNotFound
	}
	app, ok := tdb.apps[appId]
	if !ok {
		return nil, "", ErrAppNotFound
	}
	return &app, appId, nil
}

func (tdb *TempDriver) SetApp(appId string, app *App) error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	tdb.apps[appId] = *app
	return nil
}

func (tdb *TempDriver) DeleteApp(appId string) error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	delete(tdb.apps, appId)
	return nil
}

func (tdb *TempDriver) ValidSecret(secret, appId string) (bool, error) {
	tdb.lock.RLock()
	defer tdb.lock.RUnlock()
	appIdFromSecret, ok := tdb.secretToApp[secret]
	if !ok {
		return false, nil
	}
	return appIdFromSecret == appId, nil
}

func (tdb *TempDriver) SetSecret(secret, appId string) error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	tdb.secretToApp[secret] = appId
	return nil
}

func (tdb *TempDriver) DeleteSecret(secret string) error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	delete(tdb.secretToApp, secret)
	return nil
}

func (tdb *TempDriver) TokenExpiration(token Token) (time.Time, error) {
	tdb.lock.RLock()
	defer tdb.lock.RUnlock()
	exp, ok := tdb.tokens[token]
	if !ok {
		return time.Time{}, ErrTokenNotFound
	}
	return time.Unix(0, exp), nil
}

func (tdb *TempDriver) SetToken(token Token, expiration time.Time) error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	tdb.tokens[token] = expiration.UnixNano()
	return nil
}

func (tdb *TempDriver) DeleteToken(token Token) error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	delete(tdb.tokens, token)
	return nil
}

func (tdb *TempDriver) DeleteTokensByPrefix(prefix string) error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	if prefix == "" {
		return nil
	}
	for token := range tdb.tokens {
		if strings.HasPrefix(string(token), prefix) {
			delete(tdb.tokens, token)
		}
	}
	return nil
}

func (tdb *TempDriver) DeleteExpiredTokens() error {
	tdb.lock.Lock()
	defer tdb.lock.Unlock()
	now := time.Now().UnixNano()
	for token, expiration := range tdb.tokens {
		if now > expiration {
			delete(tdb.tokens, token)
		}
	}
	return nil
}

func (tdb *TempDriver) CountTokens(prefix string) (int64, error) {
	tdb.lock.RLock()
	defer tdb.lock.RUnlock()
	if prefix == "" {
		return int64(len(tdb.tokens)), nil
	}
	var count int64
	for token := range tdb.tokens {
		if strings.HasPrefix(string(token), prefix) {
			count++
		}
	}
	return count, nil
}
