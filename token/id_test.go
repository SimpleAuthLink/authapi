package token

import (
	"bytes"
	"testing"
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
}
