package token

import (
	"encoding/base64"
	"strings"
	"time"
)

type App struct {
	Name            string
	RedirectURI     string
	SessionDuration time.Duration
}

func (app *App) Valid() bool {
	if app == nil {
		return false
	}
	if len(app.Name) < appNameMinLen || len(app.Name) > appNameMaxLen {
		return false
	}
	if !uriRegexp.MatchString(app.RedirectURI) || len(app.RedirectURI) > redirectURIMaxLen {
		return false
	}
	if app.SessionDuration < minDuration || app.SessionDuration > maxDuration {
		return false
	}
	return true
}

func (app *App) Attributes() []string {
	return []string{app.Name, app.RedirectURI, app.SessionDuration.String()}
}

func (app *App) SetAttributes(attrs []string) *App {
	if len(attrs) != 3 {
		return nil
	}
	duration, err := time.ParseDuration(attrs[2])
	if err != nil {
		return nil
	}
	app.Name = attrs[0]
	app.RedirectURI = attrs[1]
	app.SessionDuration = duration
	if !app.Valid() {
		return nil
	}
	return app
}

func (app *App) String() string {
	if !app.Valid() {
		return ""
	}
	return strings.Join(app.Attributes(), appDataSeparator)
}

func (app *App) SetString(data string) *App {
	b := strings.Split(data, appDataSeparator)
	return app.SetAttributes(b)
}

func (app *App) Bytes() []byte {
	return []byte(app.String())
}

func (app *App) SetBytes(data []byte) *App {
	return app.SetString(string(data))
}

func (app *App) Marshal() []byte {
	if !app.Valid() {
		return nil
	}
	bApp := app.Bytes()
	b := make([]byte, base64.RawStdEncoding.EncodedLen(len(bApp)))
	base64.RawStdEncoding.Encode(b, bApp)
	return b
}

func (app *App) Unmarshal(data []byte) *App {
	b := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	if _, err := base64.RawStdEncoding.Decode(b, data); err != nil {
		return nil
	}
	return app.SetBytes(b)
}

func (app *App) ID() *AppID {
	if !app.Valid() {
		return nil
	}
	return new(AppID).SetBytes(app.Marshal())
}
