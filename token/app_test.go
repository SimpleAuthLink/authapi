package token

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"
)

const (
	testAppName         = "MySuperMegaApp"
	testRedirectURI     = "https://example.com/login?app=MySuperMegaApp"
	testSessionDuration = time.Minute * 30
)

var testAppSecret = new(Secret).SetParts([]byte("super_secret_key"), []byte("super_secret_salt"))

func TestValidApp(t *testing.T) {
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	if err := app.Valid(nil); err != nil {
		t.Errorf("expected valid app data")
	}
	// test app name
	app.Name = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	if err := app.Valid(nil); err == nil {
		t.Errorf("expected invalid app data")
	}
	app.Name = "no"
	if err := app.Valid(nil); err == nil {
		t.Errorf("expected invalid app data")
	}
	app.Name = testAppName
	// test redirect URI
	app.RedirectURI = "https://example.com/login?app=lorem_ipsum_dolor_sit_amet_consectetur_adipiscing_elit_sed_do_eiusmod_tempor_incididunt_ut_labore_et_dolore_magna_aliqua"
	if err := app.Valid(nil); err == nil {
		t.Errorf("expected invalid app data")
	}
	app.RedirectURI = "no_url"
	if err := app.Valid(nil); err == nil {
		t.Errorf("expected invalid app data")
	}
	app.RedirectURI = testRedirectURI
	// test session duration
	app.SessionDuration = minDuration - 1
	if err := app.Valid(nil); err == nil {
		t.Errorf("expected invalid app data")
	}
	app.SessionDuration = maxDuration + 1
	if err := app.Valid(nil); err == nil {
		t.Errorf("expected invalid app data")
	}
	var nilApp *App
	if err := nilApp.Valid(nil); err == nil {
		t.Errorf("expected invalid app data")
	}
	app.AppSecretHash = testAppSecret.Hash()
	servicePart := []byte("invalid-service-secret")
	appPart := []byte("invalid-app-secret")
	invalidSecret := new(Secret).SetParts(servicePart, appPart)
	if err := app.Valid(invalidSecret.Hash()); err == nil {
		t.Errorf("expected invalid app data")
	}
}

func TestAttributesSetAttributesApp(t *testing.T) {
	if res := new(App).SetAttributes([]string{}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	strHash := hex.EncodeToString(testAppSecret.Hash())
	if res := new(App).SetAttributes([]string{testAppName, testRedirectURI, "no_duration", strHash}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if res := new(App).SetAttributes([]string{testAppName, "no_url", testSessionDuration.String()}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if res := new(App).SetAttributes([]string{testAppName, testRedirectURI, testSessionDuration.String(), "invalid_hash"}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if res := new(App).SetAttributes([]string{testAppName, testRedirectURI, testSessionDuration.String(), "2bbc94cb9c916e1f6f1354ef30c1c80767b85159570304baa402c088180a0ec5"}); res != nil {
		t.Errorf("expected nil, got %v", res)
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
	var nilApp *App
	nilData := nilApp.SetAttributes(app.Attributes())
	if nilData == nil {
		t.Fatalf("error decoding app data")
	}
	if nilData.Name != testAppName {
		t.Errorf("expected app name %q, got %q", testAppName, nilData.Name)
	}
	if nilData.RedirectURI != testRedirectURI {
		t.Errorf("expected redirect URI %q, got %q", testRedirectURI, nilData.RedirectURI)
	}
	if nilData.SessionDuration != testSessionDuration {
		t.Errorf("expected session duration %v, got %v", testSessionDuration, nilData.SessionDuration)
	}
	data := new(App).SetAttributes(app.Attributes())
	if data == nil {
		t.Fatalf("error decoding app data")
	}
	if data.Name != testAppName {
		t.Errorf("expected app name %q, got %q", testAppName, data.Name)
	}
	if data.RedirectURI != testRedirectURI {
		t.Errorf("expected redirect URI %q, got %q", testRedirectURI, data.RedirectURI)
	}
	if data.SessionDuration != testSessionDuration {
		t.Errorf("expected session duration %v, got %v", testSessionDuration, data.SessionDuration)
	}
	// set an out of range duration to test if the app is valid
	attrs := app.Attributes()
	attrs[2] = "1s"
	if res := new(App).SetAttributes(attrs); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
}

func TestStringSetStringApp(t *testing.T) {
	if res := new(App).String(); res != "" {
		t.Errorf("expected empty string, got %q", res)
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
	data := new(App).SetString(app.String())
	if data == nil {
		t.Fatalf("error decoding app data")
	}
	if data.Name != testAppName {
		t.Errorf("expected app name %q, got %q", testAppName, data.Name)
	}
	if data.RedirectURI != testRedirectURI {
		t.Errorf("expected redirect URI %q, got %q", testRedirectURI, data.RedirectURI)
	}
	if data.SessionDuration != testSessionDuration {
		t.Errorf("expected session duration %v, got %v", testSessionDuration, data.SessionDuration)
	}
}

func TestBytesSetBytesApp(t *testing.T) {
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	app.SetSecret(secret)
	data := new(App).SetBytes(app.Bytes())
	if data == nil {
		t.Fatalf("error decoding app data")
	}
	if data.Name != testAppName {
		t.Errorf("expected app name %q, got %q", testAppName, data.Name)
	}
	if data.RedirectURI != testRedirectURI {
		t.Errorf("expected redirect URI %q, got %q", testRedirectURI, data.RedirectURI)
	}
	if data.SessionDuration != testSessionDuration {
		t.Errorf("expected session duration %v, got %v", testSessionDuration, data.SessionDuration)
	}
}

func TestMarshalUnmarshalApp(t *testing.T) {
	if res := new(App).Marshal(); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if res := new(App).Unmarshal([]byte{1}); res != nil {
		t.Errorf("expected nil, got %v", res)
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
	data := new(App).Unmarshal(app.Marshal())
	if data == nil {
		t.Fatalf("error decoding app data")
	}
	if data.Name != testAppName {
		t.Errorf("expected app name %q, got %q", testAppName, data.Name)
	}
	if data.RedirectURI != testRedirectURI {
		t.Errorf("expected redirect URI %q, got %q", testRedirectURI, data.RedirectURI)
	}
	if data.SessionDuration != testSessionDuration {
		t.Errorf("expected session duration %v, got %v", testSessionDuration, data.SessionDuration)
	}
}

func TestAppID(t *testing.T) {
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	if id := new(App).ID(secret); id != nil {
		t.Errorf("expected nil, got %v", id)
	}
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	app.SetSecret(secret)
	id := app.ID(secret)
	if id == nil {
		t.Fatalf("error decoding app ID")
	}
	if !bytes.Equal(id.Bytes(), app.Marshal()) {
		t.Errorf("expected %v, got %v", app.Marshal(), id.Bytes())
	}
	if res := new(App).SetID(nil); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	newApp := new(App).SetID(id)
	if newApp == nil {
		t.Fatalf("error decoding app ID")
	}
	if newApp.String() != app.String() {
		t.Errorf("expected %s, got %s", app.String(), newApp.String())
	}
}

func TestSetSecretApp(t *testing.T) {
	var nilApp *App
	if res := nilApp.SetSecret(nil); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	if res := app.SetSecret(nil); res == nil {
		t.Errorf("expected nil, got %v", res)
	}
	if app.AppSecretHash != nil {
		t.Errorf("expected nil, got %v", app.AppSecretHash)
	}
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	app.SetSecret(secret)
	if !bytes.Equal(app.AppSecretHash, secret.Hash()) {
		t.Errorf("expected %v, got %v", secret.Hash(), app.AppSecretHash)
	}
}
