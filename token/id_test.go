package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
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
	if privKey := new(AppID).PrivKey(testAppSecret); privKey != nil {
		t.Errorf("expected nil, got %v", privKey)
	}
	if sig := new(AppID).Sign(testAppSecret, []byte("test data")); sig != nil {
		t.Errorf("expected nil, got %v", sig)
	}
	if new(AppID).Verify(testAppSecret, []byte("test data"), []byte("test sig")) {
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
	sig := id.Sign(testAppSecret, data)
	if sig == nil {
		t.Fatalf("error signing data")
	}
	if !id.Verify(testAppSecret, data, sig) {
		t.Errorf("expected signature to be valid")
	}
	if id.Verify(testAppSecret, data, []byte("invalid sig")) {
		t.Errorf("expected signature to be invalid")
	}
}

func TestNewTokenVerifyToken(t *testing.T) {
	t.Parallel()
	if res := new(AppID).NewToken(nil, ""); res != nil {
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
	token := id.NewToken(testAppSecret, email)
	if token == nil {
		t.Fatalf("error creating token")
	}
	if !id.VerifyToken(testAppSecret, token, email) {
		t.Errorf("expected token to be valid")
	}
	time.Sleep(app.SessionDuration + 1)
	if id.VerifyToken(testAppSecret, token, email) {
		t.Errorf("expected token to be invalid")
	}
	if id.VerifyToken(testAppSecret, nil, email) {
		t.Errorf("expected token to be invalid")
	}
	exp := NewExpiration(minDuration)
	if id.VerifyToken(testAppSecret, exp.Marshal(), email) {
		t.Errorf("expected token to be invalid")
	}
}

func Test_signMsg(t *testing.T) {
	expected := []byte("testcombineddata")
	if res := signMsg([]byte("test"), []byte("combined"), []byte("data")); !bytes.Equal(res, expected) {
		t.Errorf("expected %v, got %v", expected, res)
	}
}

func Test_fmtToken(t *testing.T) {
	sig := []byte("testsig")
	exp := []byte("testexp")
	expected := append(exp, tokenSeparator)
	expected = append(expected, sig...)
	if res := fmtToken(exp, sig); !bytes.Equal(res, expected) {
		t.Errorf("expected %v, got %v", expected, res)
	}
}

func Test_hedgedNonce(t *testing.T) {
	if res := hedgedNonce(); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if res := hedgedNonce(nil); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	seed := []byte("test")
	hFn := hmac.New(sha256.New, seed)
	expected := hFn.Sum(nil)
	if res := hedgedNonce(seed); !bytes.Equal(res, expected) {
		t.Errorf("expected %v, got %v", expected, res)
	}

	seed = []byte("test")
	hFn = hmac.New(sha256.New, seed)
	in := []byte("data")
	hFn.Write(in)
	expected = hFn.Sum(nil)
	if res := hedgedNonce(seed, in); !bytes.Equal(res, expected) {
		t.Errorf("expected %v, got %v", expected, res)
	}
}
