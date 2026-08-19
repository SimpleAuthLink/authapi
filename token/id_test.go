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
	// nil app ID
	if nilID = new(AppID).SetBytes(nil); nilID != nil {
		t.Errorf("expected nil, got %v", nilID)
	}
	var nilAppID *AppID
	if bNilAppID := nilAppID.Bytes(); bNilAppID != nil {
		t.Errorf("expected nil, got %v", bNilAppID)
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

func TestGenerateTokenVerifyToken(t *testing.T) {
	t.Parallel()
	var nilAppID *AppID
	if res := nilAppID.GenerateToken(nil, nil); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if nilAppID.VerifyToken(nil, *testAppSecret) {
		t.Errorf("expected token to be invalid")
	}
	if res := new(AppID).GenerateToken(nil, nil); res != nil {
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
	if id.VerifyToken([]byte{}, *testAppSecret) {
		t.Errorf("expected token to be invalid")
	}
	dummyToken := new(Token).SetExpiration(*new(Expiration).SetDuration(minDuration * 3))
	if id.VerifyToken(*dummyToken, *testAppSecret) {
		t.Errorf("expected token to be invalid")
	}
	dummyToken = new(Token).SetSignature([]byte("test"))
	if id.VerifyToken(*dummyToken, *testAppSecret) {
		t.Errorf("expected token to be invalid")
	}
	if invalidToken := id.GenerateToken(*testAppSecret, nil); invalidToken != nil {
		t.Fatalf("expected nil, got %v", invalidToken)
	}

	userID := new(Email).SetString("test@email.com")
	token := id.GenerateToken(*testAppSecret, *userID)
	if token == nil {
		t.Fatalf("error creating token")
	}
	if !id.VerifyToken(token, *testAppSecret) {
		t.Errorf("expected token to be valid")
	}
	time.Sleep(app.SessionDuration + time.Second)
	if id.VerifyToken(token, *testAppSecret) {
		t.Errorf("expected token to be invalid")
	}
	if id.VerifyToken(nil, *testAppSecret) {
		t.Errorf("expected token to be invalid")
	}
	exp := new(Expiration).SetDuration(minDuration)
	if id.VerifyToken(exp.Marshal(), *testAppSecret) {
		t.Errorf("expected token to be invalid")
	}
}

func TestTokenModificationRejected(t *testing.T) {
	t.Parallel()

	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: minDuration,
	}
	app.SetSecret(secret)
	id := app.ID(secret)
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	
	userID := new(Email).SetString("test@email.com")
	token := id.GenerateToken(*secret, *userID)
	if token == nil {
		t.Fatalf("error creating token")
	}
	if !id.VerifyToken(token, *secret) {
		t.Fatalf("original token should be valid")
	}

	parts := bytes.Split([]byte(token.String()), []byte{tokenSeparator})
	if len(parts) != numOfTokenParts {
		t.Fatalf("expected %d parts in token, got %d", numOfTokenParts, len(parts))
	}
	expPart, sigPart := string(parts[0]), string(parts[1])

	if len(expPart) < 4 {
		t.Fatalf("expiration part too short")
	}
	modifiedExp := string(expPart[0]) + string(expPart[1]) + "B" + expPart[3:]
	modifiedToken1 := modifiedExp + string(tokenSeparator) + sigPart
	tkn1 := new(Token).SetString(modifiedToken1)
	if id.VerifyToken(*tkn1, *secret) {
		t.Errorf("modified expiration token should be invalid")
	}

	if len(sigPart) < 30 {
		t.Fatalf("signature part too short")
	}
	modifiedSig := sigPart[:28] + "f" + sigPart[29:]
	modifiedToken2 := expPart + string(tokenSeparator) + modifiedSig
	tkn2 := new(Token).SetString(modifiedToken2)
	if id.VerifyToken(*tkn2, *secret) {
		t.Errorf("modified signature token should be invalid")
	}
}

func TestTokenCrossAppRejected(t *testing.T) {
	t.Parallel()

	servicePart := []byte("service-secret")

	app1Part := []byte("app-one-secret")
	secret1 := new(Secret).SetParts(servicePart, app1Part)
	app1 := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: minDuration,
	}
	app1.SetSecret(secret1)
	id1 := app1.ID(secret1)
	if id1 == nil {
		t.Fatalf("error decoding app1 ID")
	}

	app2Part := []byte("app-two-secret")
	secret2 := new(Secret).SetParts(servicePart, app2Part)
	app2 := &App{
		Name:            "OtherApp",
		RedirectURI:     testRedirectURI,
		SessionDuration: minDuration,
	}
	app2.SetSecret(secret2)
	id2 := app2.ID(secret2)
	if id2 == nil {
		t.Fatalf("error decoding app2 ID")
	}

	userID := new(Email).SetString("test@email.com")
	token := id1.GenerateToken(*secret1, *userID)
	if token == nil {
		t.Fatalf("error creating token")
	}
	if !id1.VerifyToken(token, *secret1) {
		t.Fatalf("original token should be valid for app1")
	}
	if id2.VerifyToken(token, *secret2) {
		t.Errorf("app1's token should be invalid for app2")
	}
	if id1.VerifyToken(token, *secret2) {
		t.Errorf("token should be invalid with wrong secret")
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
