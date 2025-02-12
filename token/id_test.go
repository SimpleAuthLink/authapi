package token

import (
	"bytes"
	"testing"
	"time"
)

func TestStringSetStringAppID(t *testing.T) {
	if id := new(AppID).SetString("testID"); id != nil {
		t.Errorf("expected nil, got %v", id)
	}
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	id := app.ID()
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	if id.String() != string(app.Marshal()) {
		t.Errorf("expected %s, got %s", string(app.Marshal()), id.String())
	}
	newID := new(AppID).SetString(string(app.Marshal()))
	if newID == nil {
		t.Fatalf("error decoding app ID")
	}
	if newID.String() != id.String() {
		t.Errorf("expected %s, got %s", id.String(), newID.String())
	}
}

func TestBytesSetBytesAppID(t *testing.T) {
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	id := app.ID()
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	if !bytes.Equal(id.Bytes(), app.Marshal()) {
		t.Errorf("expected %v, got %v", app.Marshal(), id.Bytes())
	}
	newID := new(AppID).SetBytes(app.Marshal())
	if newID == nil {
		t.Fatalf("error decoding app ID")
	}
	if !bytes.Equal(newID.Bytes(), id.Bytes()) {
		t.Errorf("expected %v, got %v", id.Bytes(), newID.Bytes())
	}
}

func TestPrivKeySignVerifyAppID(t *testing.T) {
	if privKey := new(AppID).PrivKey(); privKey != nil {
		t.Errorf("expected nil, got %v", privKey)
	}
	if sig := new(AppID).Sign([]byte("test data")); sig != nil {
		t.Errorf("expected nil, got %v", sig)
	}
	if new(AppID).Verify([]byte("test data"), []byte("test sig")) {
		t.Errorf("expected signature to be invalid")
	}
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	id := app.ID()
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	data := []byte("test data")
	sig := id.Sign(data)
	if sig == nil {
		t.Fatalf("error signing data")
	}
	if !id.Verify(data, sig) {
		t.Errorf("expected signature to be valid")
	}
	if id.Verify(data, []byte("invalid sig")) {
		t.Errorf("expected signature to be invalid")
	}
}

func TestNewTokenVerifyToken(t *testing.T) {
	t.Parallel()
	if res := new(AppID).NewToken("", ""); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: 30 * time.Second,
	}
	id := app.ID()
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	email := "test@email.com"
	secret := "api_secret"
	token := id.NewToken(secret, email)
	if token == nil {
		t.Fatalf("error creating token")
	}
	if !id.VerifyToken(token, secret, email) {
		t.Errorf("expected token to be valid")
	}
	time.Sleep(app.SessionDuration + 1)
	if id.VerifyToken(token, secret, email) {
		t.Errorf("expected token to be invalid")
	}
	if id.VerifyToken(nil, secret, email) {
		t.Errorf("expected token to be invalid")
	}
	exp := NewExpiration(minDuration)
	if id.VerifyToken(exp.Marshal(), secret, email) {
		t.Errorf("expected token to be invalid")
	}
}
