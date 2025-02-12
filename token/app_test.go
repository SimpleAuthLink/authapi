package token

import (
	"bytes"
	"testing"
	"time"
)

const (
	testAppName         = "MySuperMegaApp"
	testRedirectURI     = "https://example.com/login?app=MySuperMegaApp"
	testSessionDuration = time.Minute * 30
)

func TestValidApp(t *testing.T) {
	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
	}
	if !app.Valid() {
		t.Errorf("expected valid app data")
	}
	// test app name
	app.Name = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	if app.Valid() {
		t.Errorf("expected invalid app data")
	}
	app.Name = "no"
	if app.Valid() {
		t.Errorf("expected invalid app data")
	}
	app.Name = testAppName
	// test redirect URI
	app.RedirectURI = "https://example.com/login?app=lorem_ipsum_dolor_sit_amet_consectetur_adipiscing_elit_sed_do_eiusmod_tempor_incididunt_ut_labore_et_dolore_magna_aliqua"
	if app.Valid() {
		t.Errorf("expected invalid app data")
	}
	app.RedirectURI = "no_url"
	if app.Valid() {
		t.Errorf("expected invalid app data")
	}
	app.RedirectURI = testRedirectURI
	// test session duration
	app.SessionDuration = minDuration - 1
	if app.Valid() {
		t.Errorf("expected invalid app data")
	}
	app.SessionDuration = maxDuration + 1
	if app.Valid() {
		t.Errorf("expected invalid app data")
	}
}

func TestAttributesSetAttributesApp(t *testing.T) {
	if res := new(App).SetAttributes([]string{}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}

	if res := new(App).SetAttributes([]string{testAppName, testRedirectURI, "no_duration"}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}

	if res := new(App).SetAttributes([]string{testAppName, "no_url", testSessionDuration.String()}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}

	app := &App{
		Name:            testAppName,
		RedirectURI:     testRedirectURI,
		SessionDuration: testSessionDuration,
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
	if id := new(App).ID(); id != nil {
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
