package client

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/api/io"
	"github.com/simpleauthlink/authapi/internal/fakesmtpserver"
	"github.com/simpleauthlink/authapi/notification/email"
	"github.com/simpleauthlink/authapi/notification/templates/login"
	"github.com/simpleauthlink/authapi/token"
)

var (
	// server
	testServer       = "127.0.0.1"
	testPort         = randomPort()
	testEmailPort    = randomPort()
	testEndpoint     = fmt.Sprintf("http://%s:%d", testServer, testPort)
	testServerSecret = "serversecret"
	testServerEmail  = "server@email.com"
	// client
	testAppName      = "TestingClientApp"
	testStrAppID     = "VGVzdGluZ0NsaWVudEFwcHxodHRwczovL2V4YW1wbGUuY29tL2NhbGxiYWNrfDMwc3w0YzM1ZDNjNjYwNzc0ZWE4OWExNjNjYjY"
	testRedirectURI  = "https://example.com/callback"
	testTimeout      = 30 * time.Second
	testClientSecret = "clientsecret"
	testEmailCh      = make(chan string, 1)
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create the email queue
	emailQueue, err := email.NewEmailQueue(ctx, &email.EmailConfig{
		FromName:    "test",
		FromAddress: testServerEmail,
		SMTPServer:  testServer,
		SMTPPort:    testEmailPort,
	})
	if err != nil {
		panic("failed to create email queue: " + err.Error())
	}
	// start the email queue and defer to stop it
	emailQueue.Start()
	defer emailQueue.Stop()
	smtpServer := fakesmtpserver.NewServer(testServer, testEmailPort, testEmailCh)
	if err := smtpServer.Start(ctx); err != nil {
		panic("failed to start fake SMTP server: " + err.Error())
	}
	defer smtpServer.Stop()
	// create the service
	service, err := api.New(ctx, &api.Config{
		Server:     testServer,
		ServerPort: testPort,
		Secret:     testServerSecret,
		DemoMode:   false,
	}, emailQueue)
	if err != nil {
		log.Fatalln("ERR: error creating service:", err)
	}
	// start the service in background
	go func() {
		if err := service.Start(); err != nil {
			panic("failed to start service: " + err.Error())
		}
	}()
	defer service.Stop()
	m.Run()
}

func TestSuccessFlow(t *testing.T) {
	// start the client
	cli, err := New(&Config{
		APIEndpoint: testEndpoint,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	// create a new application ID
	appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
	if err != nil {
		t.Fatalf("failed to create app ID: %v", err)
	}
	// setup the application ID in the client
	cli.SetupAppID(appID, testClientSecret)
	// request a new token for the test email
	testEmail := randomEmail()
	if err := cli.RequestToken(testEmail); err != nil {
		t.Fatalf("failed to request token: %v", err)
	}
	// wait to receive the email
	testToken, err := readTokenFromEmail(testEmail)
	if err != nil {
		t.Fatalf("failed to read token from email: %v", err)
	}
	// verify the token
	valid, _, err := cli.VerifyToken(testToken, testEmail)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}
	if !valid {
		t.Fatalf("token is not valid")
	}
	// wait for the token to expire
	time.Sleep(testTimeout)
	// verify the token after expiration
	valid, _, err = cli.VerifyToken(testToken, testEmail)
	if err != nil {
		t.Fatalf("failed to verify token after expiration: %v", err)
	}
	if valid {
		t.Fatalf("token should not be valid after expiration")
	}
}

func TestConfig(t *testing.T) {
	t.Run("init empty config", func(t *testing.T) {
		config := new(Config)
		config.init()
		if config.APIEndpoint != DefaultAPIEndpoint {
			t.Fatalf("expected APIEndpoint to be %s, got %s", DefaultAPIEndpoint, config.APIEndpoint)
		}
		if config.Timeout != DefaultClientTimeout {
			t.Fatalf("expected Timeout to be %v, got %v", DefaultClientTimeout, config.Timeout)
		}
	})
	t.Run("init non-empty config", func(t *testing.T) {
		config := &Config{
			APIEndpoint: "http://example.com",
			Timeout:     10 * time.Second,
		}
		config.init()
		if config.APIEndpoint != "http://example.com" {
			t.Fatalf("expected APIEndpoint to be http://example.com, got %s", config.APIEndpoint)
		}
		if config.Timeout != 10*time.Second {
			t.Fatalf("expected Timeout to be 10s, got %v", config.Timeout)
		}
	})
	t.Run("validate config without app", func(t *testing.T) {
		config := new(Config)
		if err := config.Validate(false); err == nil {
			t.Error("expected error when validating without app, got nil")
		}
		config.APIEndpoint = "http//example.com"
		if err := config.Validate(false); err == nil {
			t.Error("expected error for invalid APIEndpoint, got nil")
		}
		config.APIEndpoint = DefaultAPIEndpoint
		if err := config.Validate(false); err == nil {
			t.Error("expected error when validating without app, got nil")
		}
		config.init()
		if err := config.Validate(false); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := config.Validate(true); err == nil {
			t.Error("expected error when validating with requireApp true, got nil")
		}
		config.AppID = new(token.AppID)
		if err := config.Validate(true); err == nil {
			t.Error("expected error when validating with requireApp true, got nil")
		}
		config.AppSecret = "test-secret"
		if err := config.Validate(true); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		_, err := New(nil)
		if err != nil {
			t.Fatalf("expected no error when passing nil config, got %v", err)
		}
	})
	t.Run("valid config", func(t *testing.T) {
		cli, err := New(&Config{
			APIEndpoint: testEndpoint,
		})
		if err != nil {
			t.Fatalf("expected no error when passing valid config, got %v", err)
		}
		if cli == nil {
			t.Error("expected client to be created, got nil")
		}
	})
	t.Run("invalid health check request", func(t *testing.T) {
		_, err := New(&Config{
			APIEndpoint: "http//invalid-url",
		})
		if err != ErrInvalidAPIEndpoint {
			t.Fatalf("expected ErrInvalidAPIEndpoint, got %v", err)
		}
	})
	t.Run("invalid health check call", func(t *testing.T) {
		_, err := New(&Config{
			APIEndpoint: "http://nonexistent.host.issue",
		})
		if err != ErrAPIUnavailable {
			t.Fatalf("expected ErrAPIUnavailable, got %v", err)
		}
	})
	t.Run("invalid health check response", func(t *testing.T) {
		_, err := New(&Config{
			APIEndpoint: "http://google.com/invalid-endpoint",
		})
		if err != ErrAPIUnavailable {
			t.Fatalf("expected ErrAPIUnavailable, got %v", err)
		}
	})
}

func TestDefault(t *testing.T) {
	t.Run("invalid app secret", func(t *testing.T) {
		_, err := Default(testStrAppID, "")
		if err != ErrInvalidAppSecret {
			t.Fatalf("expected ErrInvalidAppSecret, got %v", err)
		}
	})
	t.Run("empty app ID", func(t *testing.T) {
		_, err := Default("", testClientSecret)
		if err != ErrInvalidAppID {
			t.Fatalf("expected ErrInvalidAppID, got %v", err)
		}
	})
	t.Run("invalid app ID", func(t *testing.T) {
		_, err := Default("invalid-app-id", testClientSecret)
		if err != ErrInvalidAppID {
			t.Fatalf("expected ErrInvalidAppID, got %v", err)
		}
	})
	t.Run("valid app ID and secret", func(t *testing.T) {
		_, err := Default(testStrAppID, testClientSecret)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestNewAppID(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		cli.config.APIEndpoint = "invalid-endpoint"
		_, err = cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != ErrInvalidAPIEndpoint {
			t.Fatalf("expected ErrInvalidAPIEndpoint, got %v", err)
		}
	})

	t.Run("invalid inputs", func(t *testing.T) {
		// initialize the client with the testing endpoint
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		_, err = cli.NewAppID("", "", "", 0)
		if err != ErrInvalidAppName {
			t.Fatalf("expected ErrInvalidAppName, got %v", err)
		}
		_, err = cli.NewAppID(testAppName, "", "", 0)
		if err != ErrInvalidAppRedirectURI {
			t.Fatalf("expected ErrInvalidAppRedirectURI, got %v", err)
		}
		_, err = cli.NewAppID(testAppName, testRedirectURI, "", 0)
		if err != ErrInvalidAppSecret {
			t.Fatalf("expected ErrInvalidAppSecret, got %v", err)
		}
		_, err = cli.NewAppID(testAppName, testRedirectURI, testClientSecret, 0)
		if err != ErrInvalidAppSessionDuration {
			t.Fatalf("expected ErrInvalidAppSessionDuration, got %v", err)
		}
	})
	t.Run("request error", func(t *testing.T) {
		// initialize the client with the testing endpoint
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		cli.config.APIEndpoint = "http://invalid-endpoint"
		_, err = cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != ErrRequestAppID {
			t.Fatalf("expected ErrRequestAppID, got %v", err)
		}

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer srv.Close()
		cli.config.APIEndpoint = srv.URL
		_, err = cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != io.ErrHTTPStatusNotOk {
			t.Fatalf("expected ErrHTTPStatusNotOk, got %v", err)
		}
	})
	t.Run("success", func(t *testing.T) {
		// initialize the client with the testing endpoint
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if appID == nil {
			t.Error("expected appID to be created, got nil")
		}
		cli.SetupAppID(appID, testClientSecret)
		if cli.config.AppID == nil || cli.config.AppID.String() != appID.String() {
			t.Fatalf("expected client config AppID to be set, got %v", cli.config.AppID)
		}
		if cli.config.AppSecret != testClientSecret {
			t.Fatalf("expected client config AppSecret to be set, got %s", cli.config.AppSecret)
		}
	})
}

func TestRequestToken(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		cli.config.APIEndpoint = "invalid-endpoint"
		testEmail := randomEmail()
		err = cli.RequestToken(testEmail)
		if err != ErrInvalidAPIEndpoint {
			t.Fatalf("expected ErrInvalidAPIEndpoint, got %v", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		if err := cli.RequestToken(""); err != ErrInvalidEmailAddress {
			t.Fatalf("expected ErrInvalidEmailAddress, got %v", err)
		}
		if err := cli.RequestToken("invalid-email"); err != ErrInvalidEmailAddress {
			t.Fatalf("expected ErrInvalidEmailAddress, got %v", err)
		}
	})

	t.Run("request error", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		// invalid API endpoint
		cli.config.APIEndpoint = "http://nonexistent.host.issue"
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != ErrAPIUnavailable {
			t.Fatalf("expected ErrAPIUnavailable, got %v", err)
		}
		// non-200 response
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer srv.Close()
		cli.config.APIEndpoint = srv.URL
		if err := cli.RequestToken(testEmail); err != ErrRequestToken {
			t.Fatalf("expected ErrRequestToken, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestVerifyToken(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		testEmail := randomEmail()
		if _, _, err := cli.VerifyToken(nil, testEmail); err != ErrInvalidAppID {
			t.Fatalf("expected ErrInvalidAppID, got %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		testEmail := randomEmail()
		if _, _, err := cli.VerifyToken(nil, testEmail); err != ErrInvalidToken {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		if _, _, err := cli.VerifyToken(testToken, ""); err != ErrInvalidEmailAddress {
			t.Fatalf("expected ErrInvalidEmailAddress, got %v", err)
		}
	})
	t.Run("request error", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		// invalid API endpoint
		cli.config.APIEndpoint = "http://nonexistent.host.issue"
		if _, _, err := cli.VerifyToken(testToken, testEmail); err != ErrAPIUnavailable {
			t.Fatalf("expected ErrAPIUnavailable, got %v", err)
		}
	})
	t.Run("non-200 response", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer srv.Close()
		cli.config.APIEndpoint = srv.URL
		if _, _, err := cli.VerifyToken(testToken, testEmail); err != io.ErrHTTPStatusNotOk {
			t.Fatalf("expected ErrHTTPStatusNotOk, got %v", err)
		}
	})
	t.Run("success", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		appID, err := cli.NewAppID(testAppName, testRedirectURI, testClientSecret, testTimeout)
		if err != nil {
			t.Fatalf("failed to create app ID: %v", err)
		}
		cli.SetupAppID(appID, testClientSecret)
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		valid, expTime, err := cli.VerifyToken(testToken, testEmail)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !valid {
			t.Fatal("expected token to be valid")
		}
		if expTime.IsZero() {
			t.Fatal("expected expiration time to be set")
		}
	})
}

func readTokenFromEmail(email string) (*token.Token, error) {
	// wait for the email to be sent
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(testTimeout):
			return nil, fmt.Errorf("timeout waiting for email")
		case msg := <-testEmailCh:
			token := login.FindToken(email, msg)
			if token != nil {
				return token, nil
			}
		}
	}
}

func TestAuthorizedRequestURLParams(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		cli.config.APIEndpoint = "invalid-endpoint"
		_, _, err = cli.AuthorizedRequestURLParams(nil)
		if err != ErrInvalidAPIEndpoint {
			t.Fatalf("expected ErrInvalidAPIEndpoint, got %v", err)
		}
	})

	cli, err := New(&Config{
		APIEndpoint: testEndpoint,
		AppID:       new(token.AppID).SetString(testStrAppID),
		AppSecret:   testClientSecret,
		Timeout:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Run("valid params in request url", func(t *testing.T) {
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		reqURL, _ := url.Parse("http://example.com/test")
		reqURL.RawQuery = fmt.Sprintf("%s=%s&%s=%s",
			DefaultAuthEmailURLParam, url.QueryEscape(testEmail),
			DefaultAuthTokenURLParam, url.QueryEscape(testToken.String()))

		req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		email, _, err := cli.AuthorizedRequestURLParams(req)
		if err != nil {
			t.Fatalf("failed to get authorized request URL params: %v", err)
		}
		if email != testEmail {
			t.Fatalf("expected email to be %s, got %s", testEmail, email)
		}
	})

	t.Run("no params in request url", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		_, _, err = cli.AuthorizedRequestURLParams(request)
		if err != ErrMissingAuthURLParams {
			t.Fatalf("expected ErrMissingAuthURLParams, got %v", err)
		}
	})

	t.Run("invalid params in request url", func(t *testing.T) {
		invalidEmailURL, _ := url.Parse("http://example.com/test")
		invalidEmailURL.RawQuery = DefaultAuthEmailHeader + "=invalid%2G"
		request, err := http.NewRequest(http.MethodGet, invalidEmailURL.String(), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		_, _, err = cli.AuthorizedRequestURLParams(request)
		if err != ErrMissingAuthURLParams {
			t.Fatalf("expected ErrMissingAuthURLParams, got %v", err)
		}
	})

	t.Run("invalid email in request params", func(t *testing.T) {
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		invalidEmail := "noemail.com"
		reqURL, _ := url.Parse("http://example.com/test")
		reqURL.RawQuery = fmt.Sprintf("%s=%s&%s=%s",
			DefaultAuthEmailURLParam, url.QueryEscape(invalidEmail),
			DefaultAuthTokenURLParam, url.QueryEscape(testToken.String()))

		req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		email, valid, err := cli.AuthorizedRequestURLParams(req)
		if err != ErrInvalidEmailAddress {
			t.Fatalf("expected ErrInvalidEmailAddress, got %v", err)
		}
		if valid {
			t.Fatal("expected token to be invalid")
		}
		if email != "" {
			t.Fatalf("expected email to be empty, got %s", email)
		}
	})

	t.Run("invalid auth in request params", func(t *testing.T) {
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		otherEmail := "other@email.com"
		reqURL, _ := url.Parse("http://example.com/test")
		reqURL.RawQuery = fmt.Sprintf("%s=%s&%s=%s",
			DefaultAuthEmailURLParam, url.QueryEscape(otherEmail),
			DefaultAuthTokenURLParam, url.QueryEscape(testToken.String()))

		req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		email, valid, err := cli.AuthorizedRequestURLParams(req)
		if err != nil {
			t.Fatalf("failed to get authorized request URL params: %v", err)
		}
		if valid {
			t.Fatal("expected token to be invalid")
		}
		if email != otherEmail {
			t.Fatalf("expected email to be empty, got %s", email)
		}
	})
}

func TestAuthorizedRequestHeaders(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		cli, err := New(&Config{APIEndpoint: testEndpoint})
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		cli.config.APIEndpoint = "invalid-endpoint"
		_, _, err = cli.AuthorizedRequestHeaders(nil)
		if err != ErrInvalidAPIEndpoint {
			t.Fatalf("expected ErrInvalidAPIEndpoint, got %v", err)
		}
	})

	cli, err := New(&Config{
		APIEndpoint: testEndpoint,
		AppID:       new(token.AppID).SetString(testStrAppID),
		AppSecret:   testClientSecret,
		Timeout:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Run("valid params in request headers", func(t *testing.T) {
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		req, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set(DefaultAuthEmailHeader, testEmail)
		req.Header.Set(DefaultAuthTokenHeader, testToken.String())
		email, valid, err := cli.AuthorizedRequestHeaders(req)
		if err != nil {
			t.Fatalf("failed to get authorized request URL params: %v", err)
		}
		if !valid {
			t.Fatal("expected token to be valid")
		}
		if email != testEmail {
			t.Fatalf("expected email to be %s, got %s", testEmail, email)
		}
	})

	t.Run("no params in request headers", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		_, _, err = cli.AuthorizedRequestHeaders(request)
		if err != ErrMissingAuthHeaders {
			t.Fatalf("expected ErrMissingAuthHeaders, got %v", err)
		}
	})

	t.Run("invalid email in request headers", func(t *testing.T) {
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		invalidEmail := "noemail.com"

		req, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set(DefaultAuthEmailHeader, invalidEmail)
		req.Header.Set(DefaultAuthTokenHeader, testToken.String())
		email, valid, err := cli.AuthorizedRequestHeaders(req)
		if err != ErrInvalidEmailAddress {
			t.Fatalf("expected ErrInvalidEmailAddress, got %v", err)
		}
		if valid {
			t.Fatal("expected token to be invalid")
		}
		if email != "" {
			t.Fatalf("expected email to be empty, got %s", email)
		}
	})

	t.Run("invalid auth in request headers", func(t *testing.T) {
		testEmail := randomEmail()
		if err := cli.RequestToken(testEmail); err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		testToken, err := readTokenFromEmail(testEmail)
		if err != nil {
			t.Fatalf("failed to read token from email: %v", err)
		}
		otherEmail := "other@email.com"
		req, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set(DefaultAuthEmailHeader, otherEmail)
		req.Header.Set(DefaultAuthTokenHeader, testToken.String())
		email, valid, err := cli.AuthorizedRequestHeaders(req)
		if err != nil {
			t.Fatalf("failed to get authorized request headers: %v", err)
		}
		if valid {
			t.Fatal("expected token to be invalid")
		}
		if email != otherEmail {
			t.Fatalf("expected email to be empty, got %s", email)
		}
	})
}

func Test_req(t *testing.T) {
	type testData struct{}

	t.Run("invalid config", func(t *testing.T) {
		_, err := req(&Config{}, http.MethodGet, "/test", &testData{})
		if err != ErrInvalidAPIEndpoint {
			t.Fatalf("expected ErrInvalidAPIEndpoint, got %v", err)
		}
	})

	t.Run("failed to create request", func(t *testing.T) {
		exampleConfig := &Config{APIEndpoint: "http://example.com"}
		exampleConfig.init()
		_, err := req(exampleConfig, string(byte(0)), "/test", &testData{})
		if err != ErrCreateRequest {
			t.Fatalf("expected ErrCreateRequest, got %v", err)
		}
	})

	t.Run("without headers", func(t *testing.T) {
		exampleConfig := &Config{APIEndpoint: "http://example.com"}
		exampleConfig.init()
		req, err := req(exampleConfig, http.MethodGet, "/test", &testData{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if req == nil {
			t.Fatal("expected request to be created, got nil")
		}
		if req.Header.Get(api.AppIDHeader) != "" {
			t.Fatalf("expected AppID header to be empty, got %s", req.Header.Get(api.AppIDHeader))
		}
		if req.Header.Get(api.AppSecretHeader) != "" {
			t.Fatalf("expected AppSecret header to be empty, got %s", req.Header.Get(api.AppSecretHeader))
		}
		if req.Method != http.MethodGet {
			t.Fatalf("expected request method to be %s, got %s", http.MethodGet, req.Method)
		}
		if req.URL.String() != "http://example.com/test" {
			t.Fatalf("expected request URL to be http://example.com/test, got %s", req.URL.String())
		}
	})

	t.Run("with headers", func(t *testing.T) {
		exampleConfig := &Config{APIEndpoint: "http://example.com"}
		exampleConfig.init()

		app := &token.App{
			Name:            testAppName,
			RedirectURI:     testRedirectURI,
			SessionDuration: testTimeout,
		}
		servicePart := []byte("service-secret")
		appPart := []byte("app-secret")
		secret := new(token.Secret).SetParts(servicePart, appPart)
		app.SetSecret(secret)
		exampleConfig.AppID = app.ID(secret)
		exampleConfig.AppSecret = "app-secret"

		req, err := req(exampleConfig, http.MethodGet, "/test", &testData{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if req == nil {
			t.Fatal("expected request to be created, got nil")
		}
		if req.Header.Get(api.AppIDHeader) != exampleConfig.AppID.String() {
			t.Fatalf("expected AppID header to be set, got %s", req.Header.Get(api.AppIDHeader))
		}
		if req.Header.Get(api.AppSecretHeader) != exampleConfig.AppSecret {
			t.Fatalf("expected AppSecret header to be set, got %s", req.Header.Get(api.AppSecretHeader))
		}
		if req.Method != http.MethodGet {
			t.Fatalf("expected request method to be %s, got %s", http.MethodGet, req.Method)
		}
		if req.URL.String() != "http://example.com/test" {
			t.Fatalf("expected request URL to be http://example.com/test, got %s", req.URL.String())
		}
	})

	t.Run("bad marshal data", func(t *testing.T) {
		exampleConfig := &Config{APIEndpoint: "http://example.com"}
		exampleConfig.init()

		invalidData := struct {
			Data float64
		}{
			Data: math.NaN(),
		}

		_, err := req(exampleConfig, http.MethodPost, "/test", &invalidData)
		if err != io.ErrEncodeBody {
			t.Fatalf("expected ErrEncodeBody, got %v", err)
		}
	})
}

func randomEmail() string {
	randStr := strings.ToLower(rand.Text()[0:8])
	return fmt.Sprintf("%s@example.com", randStr)
}

func randomPort() int {
	port, err := rand.Int(rand.Reader, big.NewInt(65535-1024))
	if err != nil {
		panic(err)
	}
	return int(port.Int64()) + 1024
}
