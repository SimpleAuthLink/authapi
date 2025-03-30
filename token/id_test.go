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
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	app.SetSecret(secret)
	id := app.ID(secret)
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
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	app.SetSecret(secret)
	id := app.ID(secret)
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
	var nilID *AppID
	nilID = nilID.SetBytes(app.Marshal())
	if nilID == nil {
		t.Fatalf("error decoding app ID")
	}
	if !bytes.Equal(nilID.Bytes(), id.Bytes()) {
		t.Errorf("expected %v, got %v", id.Bytes(), nilID.Bytes())
	}
}

func TestPrivKeySignVerifyAppID(t *testing.T) {
	var nilAppID *AppID
	if privKey := nilAppID.PrivKey(*testAppSecret); privKey != nil {
		t.Errorf("expected nil, got %v", privKey)
	}
	if nilSig := nilAppID.Sign(*testAppSecret, []byte("test data")); nilSig != nil {
		t.Errorf("expected nil, got %v", nilSig)
	}
	if nilVerify := nilAppID.Verify(*testAppSecret, []byte("test data"), []byte("test sig")); nilVerify {
		t.Errorf("expected signature to be invalid")
	}
	badSecret := new(Secret)
	if nilPrivKey := new(AppID).PrivKey(*badSecret); nilPrivKey != nil {
		t.Errorf("expected nil, got %v", nilPrivKey)
	}
	if privKey := new(AppID).PrivKey(*testAppSecret); privKey != nil {
		t.Errorf("expected nil, got %v", privKey)
	}
	if sig := new(AppID).Sign(*testAppSecret, []byte("test data")); sig != nil {
		t.Errorf("expected nil, got %v", sig)
	}
	if new(AppID).Verify(*testAppSecret, []byte("test data"), []byte("test sig")) {
		t.Errorf("expected signature to be invalid")
	}
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	app.SetSecret(secret)
	id := app.ID(secret)
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	data := []byte("test data")
	sig := id.Sign(*testAppSecret, data)
	if sig == nil {
		t.Fatalf("error signing data")
	}
	if !id.Verify(*testAppSecret, data, sig) {
		t.Errorf("expected signature to be valid")
	}
	if id.Verify(*testAppSecret, data, []byte("invalid sig")) {
		t.Errorf("expected signature to be invalid")
	}
}

func TestMessage(t *testing.T) {
	if privKey := new(AppID).PrivKey(*testAppSecret); privKey != nil {
		t.Errorf("expected nil, got %v", privKey)
	}
	if sig := new(AppID).Sign(*testAppSecret, []byte("test data")); sig != nil {
		t.Errorf("expected nil, got %v", sig)
	}
}

func TestGenerateTokenVerifyToken(t *testing.T) {
	t.Parallel()
	var nilAppID *AppID
	if res := nilAppID.GenerateToken(nil, ""); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if nilAppID.VerifyToken(nil, *testAppSecret, "") {
		t.Errorf("expected token to be invalid")
	}
	if res := new(AppID).GenerateToken(nil, ""); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: minDuration,
	}
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	app.SetSecret(secret)
	id := app.ID(secret)
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	email := "test@email.com"
	if id.VerifyToken([]byte{}, *testAppSecret, email) {
		t.Errorf("expected token to be invalid")
	}
	dummyToken := new(Token).SetExpiration(*new(Expiration).SetDuration(minDuration * 3))
	if id.VerifyToken(*dummyToken, *testAppSecret, email) {
		t.Errorf("expected token to be invalid")
	}
	dummyToken = new(Token).SetSignature([]byte("test"))
	if id.VerifyToken(*dummyToken, *testAppSecret, email) {
		t.Errorf("expected token to be invalid")
	}
	if invalidToken := id.GenerateToken(*testAppSecret, ""); invalidToken != nil {
		t.Fatalf("expected nil, got %v", invalidToken)
	}

	token := id.GenerateToken(*testAppSecret, email)
	if token == nil {
		t.Fatalf("error creating token")
	}
	if id.VerifyToken(token, *testAppSecret, "") {
		t.Errorf("expected token to be invalid")
	}
	if !id.VerifyToken(token, *testAppSecret, email) {
		t.Errorf("expected token to be valid")
	}
	time.Sleep(app.SessionDuration + time.Second)
	if id.VerifyToken(token, *testAppSecret, email) {
		t.Errorf("expected token to be invalid")
	}
	if id.VerifyToken(nil, *testAppSecret, email) {
		t.Errorf("expected token to be invalid")
	}
	exp := new(Expiration).SetDuration(minDuration)
	if id.VerifyToken(exp.Marshal(), *testAppSecret, email) {
		t.Errorf("expected token to be invalid")
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
