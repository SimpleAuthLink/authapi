package token

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"
)

// App represents an application that can request tokens. It has a name, a
// redirect URI, and a session duration.
type App struct {
	Name            string
	RedirectURI     string
	SessionDuration time.Duration
	AppSecretHash   []byte
}

// Valid method returns true if the app is valid, false otherwise. An app is
// considered valid if its name is between 3 and 20 characters, its redirect
// URI is a valid URI, and its session duration is between 5 minutes and 24
// hours.
func (app *App) Valid(secretHash []byte) bool {
	if app == nil {
		return false
	}
	// check if the app name is between the min and max length
	if len(app.Name) < appNameMinLen || len(app.Name) > appNameMaxLen {
		return false
	}
	// check if the redirect URI is valid
	if !uriRegexp.MatchString(app.RedirectURI) || len(app.RedirectURI) > redirectURIMaxLen {
		return false
	}
	// check if the session duration is between the min and max duration
	if app.SessionDuration < minDuration || app.SessionDuration > maxDuration {
		return false
	}
	if secretHash != nil {
		return bytes.Equal(app.AppSecretHash, secretHash)
	}
	return true
}

// Attributes method returns the app's attributes as a slice of strings. This
// is useful for encoding the app.
func (app *App) Attributes() []string {
	return []string{app.Name, app.RedirectURI, app.SessionDuration.String(), hex.EncodeToString(app.AppSecretHash)}
}

// SetAttributes method sets the app's attributes from a slice of strings. This
// is useful for decoding the app.
func (app *App) SetAttributes(attrs []string) *App {
	// check if the slice has the correct number of attributes
	if len(attrs) != 4 {
		return nil
	}
	// parse the session duration
	duration, err := time.ParseDuration(attrs[2])
	if err != nil {
		return nil
	}
	// if the app is nil, create a new app
	if app == nil {
		app = new(App)
	}
	// set the app's attributes
	app.Name = attrs[0]
	app.RedirectURI = attrs[1]
	app.SessionDuration = duration
	appSecretHash, err := hex.DecodeString(attrs[3])
	if err != nil {
		return nil
	}
	if len(appSecretHash) != secretHashSize {
		return nil
	}
	app.AppSecretHash = appSecretHash
	// check if the app is valid and return it if it is
	if !app.Valid(nil) {
		return nil
	}
	return app
}

// String method returns the app as a string. This is useful for debugging
// and encoding the app. The resulting string is the app's attributes joined
// by the app data separator.
func (app *App) String() string {
	if !app.Valid(nil) {
		return ""
	}
	// join the app's attributes with the app data separator
	return strings.Join(app.Attributes(), appDataSeparator)
}

// SetString method sets the app from a string. This is useful for decoding
// the app. The string should be the app's attributes joined by the app data
// separator.
func (app *App) SetString(data string) *App {
	b := strings.Split(data, appDataSeparator)
	return app.SetAttributes(b)
}

// Bytes method returns the app as a byte slice. This is useful for encoding
// the app. It is equivalent to converting the app to a string and then
// converting the string to a byte slice.
func (app *App) Bytes() []byte {
	return []byte(app.String())
}

// SetBytes method sets the app from a byte slice. This is useful for decoding
// the app. It is equivalent to converting the byte slice to a string and then
// converting the string to the app.
func (app *App) SetBytes(data []byte) *App {
	return app.SetString(string(data))
}

// Marshal method returns the app as a base64-encoded byte slice. It is used
// to be included in the app ID, which makes it self-contained.
func (app *App) Marshal() []byte {
	if !app.Valid(nil) {
		return nil
	}
	bApp := app.Bytes()
	b := make([]byte, base64.RawStdEncoding.EncodedLen(len(bApp)))
	base64.RawStdEncoding.Encode(b, bApp)
	return b
}

// Unmarshal method sets the app from a base64-encoded byte slice. It is used
// to extract the app from the app ID.
func (app *App) Unmarshal(data []byte) *App {
	b := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	if _, err := base64.RawStdEncoding.Decode(b, data); err != nil {
		return nil
	}
	return app.SetBytes(b)
}

// ID method returns the app ID of the app. The app ID is a self-contained
// representation of the app that can be used to generate tokens. It is
// created by encoding the app as a base64-encoded byte slice using the
// Marshal method.
func (app *App) ID(secret *Secret) *AppID {
	if !app.Valid(secret.Hash()) {
		return nil
	}
	return new(AppID).SetBytes(app.Marshal())
}

// SetID method sets the app from an app ID. The app ID is a self-contained
// representation of the app that can be used to generate tokens. The app is
// extracted from the app ID by decoding the app as a base64-encoded byte
// slice using the Unmarshal method.
func (app *App) SetID(id *AppID) *App {
	if id == nil {
		return nil
	}
	return app.Unmarshal(id.Bytes())
}

func (app *App) SetSecret(secret *Secret) *App {
	if app == nil {
		return nil
	}
	if secret == nil {
		return app
	}
	// set the app secret hash
	app.AppSecretHash = secret.Hash()
	return app
}
