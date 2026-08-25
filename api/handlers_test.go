package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/notification"
	"github.com/simpleauthlink/authapi/notification/email"
	"github.com/simpleauthlink/authapi/notification/email/templates/login"
	"github.com/simpleauthlink/authapi/token"
	xnet "go.k7z7z.cc/x/net"
	apiio "go.k7z7z.cc/x/net/http/io"
)

type testCaseAPIHandler[ReqType, ResType any] struct {
	name     string
	method   string
	endpoint string
	header   http.Header
	request  *ReqType
	response *ResType
	err      *apiio.APIError
}

func (testCase testCaseAPIHandler[Rq, Rs]) url() string {
	return fmt.Sprintf("%s%s", testServerAPIURL, testCase.endpoint)
}

func (testCase testCaseAPIHandler[Rq, Rs]) Run(t *testing.T) {
	t.Run(testCase.name, func(t *testing.T) {
		var reqBuffer io.Reader
		if testCase.request != nil {
			rawBody, err := json.Marshal(testCase.request)
			if err != nil {
				t.Fatalf("could not marshal request: %v", err)
			}
			reqBuffer = bytes.NewReader(rawBody)
		}
		req, err := http.NewRequest(testCase.method, testCase.url(), reqBuffer)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}
		req.Header = testCase.header
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("could not send request: %v", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		switch {
		case testCase.err != nil:
			if resp.StatusCode != testCase.err.StatusCode {
				t.Fatalf("expected status code: %d, got: %d", testCase.err.StatusCode, resp.StatusCode)
			}
			err := new(apiio.APIError)
			if err := json.NewDecoder(resp.Body).Decode(err); err != nil {
				t.Fatalf("could not decode error response: %v", err)
			}
			if err.Code != testCase.err.Code {
				t.Fatalf("expected error code: %d, got: %d", testCase.err.Code, err.Code)
			}
			return
		case testCase.response != nil:
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
			}
			expected, err := json.Marshal(testCase.response)
			if err != nil {
				t.Fatalf("could not marshal response: %v", err)
			}
			res, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}
			if !bytes.Equal(bytes.TrimSpace(expected), bytes.TrimSpace(res)) {
				t.Fatalf("expected response: %s, got: %s", expected, res)
			}
		default:
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
			}
		}
	})
}

func TestGenerateAppIDHandler(t *testing.T) {
	testApp := &token.App{
		Name:            testAppName,
		RedirectURI:     testAppRedirectURL,
		SessionDuration: testAppSessionDuration,
	}
	secret := new(token.Secret).SetParts([]byte(testServerSecret), []byte(testAppSecret))
	testApp.SetSecret(secret)
	testCaseAPIHandler[AppIDRequest, AppIDResponse]{
		name:     "valid request",
		method:   http.MethodPost,
		endpoint: AppsPath,
		request: &AppIDRequest{
			Name:        testApp.Name,
			RedirectURL: testApp.RedirectURI,
			Duration:    testApp.SessionDuration.String(),
			Secret:      testAppSecret,
		},
		response: &AppIDResponse{
			ID: testApp.ID(secret).String(),
		},
	}.Run(t)
	testCaseAPIHandler[AppIDRequest, AppIDResponse]{
		name:     "no request",
		method:   http.MethodPost,
		endpoint: AppsPath,
		request:  nil,
		err:      ErrDecodeAppIDRequest,
	}.Run(t)
	testCaseAPIHandler[AppIDRequest, AppIDResponse]{
		name:     "invalid request",
		method:   http.MethodPost,
		endpoint: AppsPath,
		request: &AppIDRequest{
			Name:        testAppName,
			RedirectURL: testAppRedirectURL,
			Duration:    time.Second.String(),
			Secret:      testAppSecret,
		},
		err: ErrInvalidAppID,
	}.Run(t)
	testCaseAPIHandler[AppIDRequest, AppIDResponse]{
		name:     "invalid duration string",
		method:   http.MethodPost,
		endpoint: AppsPath,
		request: &AppIDRequest{
			Name:        testAppName,
			RedirectURL: testAppRedirectURL,
			Duration:    "not-a-duration",
			Secret:      testAppSecret,
		},
		err: ErrInvalidAppID,
	}.Run(t)
}

func TestRequestTokenAndStatusHandler(t *testing.T) {
	testApp := &token.App{
		Name:            testAppName,
		RedirectURI:     testAppRedirectURL,
		SessionDuration: testAppSessionDuration,
	}
	secret := new(token.Secret).SetParts([]byte(testServerSecret), []byte(testAppSecret))
	testApp.SetSecret(secret)
	testAppID := testApp.ID(secret)
	testCaseAPIHandler[TokenRequest, any]{
		name:     "no appID request",
		method:   http.MethodPost,
		endpoint: TokensPath,
		header: http.Header{
			AppSecretHeader: []string{testAppSecret},
		},
		request: &TokenRequest{
			Email: testUserEmail,
		},
		response: nil,
		err:      ErrInvalidAppHeaders,
	}.Run(t)
	testCaseAPIHandler[TokenRequest, any]{
		name:     "invalid app id request",
		method:   http.MethodPost,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{"invalid"},
			AppSecretHeader: []string{testAppSecret},
		},
		request: &TokenRequest{
			Email: testUserEmail,
		},
		response: nil,
		err:      ErrInvalidAppID,
	}.Run(t)
	testCaseAPIHandler[TokenRequest, any]{
		name:     "no app secret request",
		method:   http.MethodPost,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader: []string{testAppID.String()},
		},
		request: &TokenRequest{
			Email: testUserEmail,
		},
		response: nil,
		err:      ErrInvalidAppHeaders,
	}.Run(t)
	invalid := []byte("invalid")
	testCaseAPIHandler[[]byte, any]{
		name:     "no request",
		method:   http.MethodPost,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{testAppID.String()},
			AppSecretHeader: []string{testAppSecret},
		},
		request:  &invalid,
		response: nil,
		err:      ErrDecodeTokenRequest,
	}.Run(t)

	originalTemplate := login.Template
	defer func() {
		login.Template = originalTemplate
	}()
	login.Template = email.EmailTemplate{
		HTML:  "",
		Plain: `\[{{.Token}}]`,
	}
	testCaseAPIHandler[TokenRequest, any]{
		name:     "valid request",
		method:   http.MethodPost,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{testAppID.String()},
			AppSecretHeader: []string{testAppSecret},
		},
		request: &TokenRequest{
			Email: testUserEmail,
		},
		response: nil,
	}.Run(t)

	var testToken *token.Token
	select {
	case receivedMsg := <-inboxChan:
		testToken = login.FindToken(testUserEmail, receivedMsg)
		if testToken == nil {
			t.Fatal("could not find token in email")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the email to be received")
	}

	testCaseAPIHandler[TokenStatusRequest, TokenStatusResponse]{
		name:     "valid token status request",
		method:   http.MethodPut,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{testAppID.String()},
			AppSecretHeader: []string{testAppSecret},
		},
		request: &TokenStatusRequest{
			Token: testToken.String(),
		},
		response: &TokenStatusResponse{
			Valid:      true,
			Expiration: testToken.Expiration().Time(),
		},
	}.Run(t)

	// invalid token tests
	testTokenStr := testToken.String()
	if !testToken.Valid() {
		t.Fatal("valid token expected")
	}
	parts := bytes.Split([]byte(testTokenStr), []byte("."))
	if len(parts) != 3 {
		t.Fatalf("expected 2 token parts, got %d", len(parts))
	}
	expPart, sigPart := string(parts[0]), string(parts[1])

	verifyToken := func(name, tokenStr string, expectValid bool) {
		t.Run(name, func(t *testing.T) {
			rawBody, _ := json.Marshal(&TokenStatusRequest{
				Token: tokenStr,
			})
			req, err := http.NewRequest(http.MethodPut,
				testServerAPIURL+TokensPath, bytes.NewReader(rawBody))
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			req.Header.Set(AppIDHeader, testAppID.String())
			req.Header.Set(AppSecretHeader, testAppSecret)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("could not send request: %v", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}
			var tr TokenStatusResponse
			if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
				t.Fatalf("could not decode response: %v", err)
			}
			if tr.Valid != expectValid {
				t.Errorf("expected Valid=%v, got Valid=%v", expectValid, tr.Valid)
			}
		})
	}

	// verify POST /tokens returns 200 for any body (dangerous if misused for verification)
	verifyPostNotVerify := func(name, body string) {
		t.Run("POST "+name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost,
				testServerAPIURL+TokensPath,
				bytes.NewReader([]byte(body)))
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			req.Header.Set(AppIDHeader, testAppID.String())
			req.Header.Set(AppSecretHeader, testAppSecret)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("could not send request: %v", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("POST with %s returned %d, expected 200", name, resp.StatusCode)
			}
		})
	}
	// If user accidentally sends token+email via POST, it returns 200,
	// looks like "verified"
	invalidTokenBody := fmt.Sprintf(`{"token":"%s","email":"%s"}`, testTokenStr, testUserEmail)
	verifyPostNotVerify("valid token via POST", invalidTokenBody)
	verifyPostNotVerify("modified exp via POST", fmt.Sprintf(`{"token":"%s","email":"%s"}`,
		expPart[:2]+"B"+expPart[3:]+"."+sigPart, testUserEmail))
	verifyPostNotVerify("garbage via POST", `{"token":"abc","email":"x@y.com"}`)

	verifyToken("valid original token", testTokenStr, true)

	if len(expPart) >= 3 {
		modifiedExp := expPart[:2] + "B" + expPart[3:]
		verifyToken("modified expiration (A→B)", modifiedExp+"."+sigPart, false)
	}

	if len(sigPart) >= 29 {
		modifiedSig := sigPart[:28] + "f" + sigPart[29:]
		verifyToken("modified signature (a→f)", expPart+"."+modifiedSig, false)
	}

	verifyToken("empty token", "", false)
	verifyToken("garbage token", "not.a.token", false)
	verifyToken("random token", "YWJj.eHl6", false)

	testCaseAPIHandler[TokenStatusRequest, any]{
		name:     "invalid app id",
		method:   http.MethodPut,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{"invalid"},
			AppSecretHeader: []string{testAppSecret},
		},
		request: &TokenStatusRequest{
			Token: testToken.String(),
		},
		response: nil,
		err:      ErrInvalidAppID,
	}.Run(t)

	testCaseAPIHandler[TokenStatusRequest, any]{
		name:     "no app secret",
		method:   http.MethodPut,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader: []string{testAppID.String()},
		},
		request: &TokenStatusRequest{
			Token: testToken.String(),
		},
		response: nil,
		err:      ErrInvalidAppHeaders,
	}.Run(t)

	testCaseAPIHandler[TokenStatusRequest, any]{
		name:     "no headers",
		method:   http.MethodPut,
		endpoint: TokensPath,
		request: &TokenStatusRequest{
			Token: testToken.String(),
		},
		response: nil,
		err:      ErrInvalidAppHeaders,
	}.Run(t)

	testCaseAPIHandler[any, any]{
		name:     "no request",
		method:   http.MethodPut,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{testAppID.String()},
			AppSecretHeader: []string{testAppSecret},
		},
		err: ErrDecodeTokenStatusRequest,
	}.Run(t)
}

func TestRequestTokenHandlerErrorBranches(t *testing.T) {
	// Build the same app/secret as TestMain
	app := &token.App{
		Name:            testAppName,
		RedirectURI:     testAppRedirectURL,
		SessionDuration: testAppSessionDuration,
	}
	secret := new(token.Secret).SetParts([]byte(testServerSecret), []byte(testAppSecret))
	app.SetSecret(secret)
	appID := app.ID(secret)

	t.Run("invalid secret provided", func(t *testing.T) {
		testCaseAPIHandler[TokenRequest, any]{
			name:     "whitespace-only secret",
			method:   http.MethodPost,
			endpoint: TokensPath,
			header: http.Header{
				AppIDHeader:     []string{appID.String()},
				AppSecretHeader: []string{"\u00a0"},
			},
			request: &TokenRequest{Email: testUserEmail},
			err:     ErrInvalidAppSecret,
		}.Run(t)
	})

	t.Run("invalid notification channel", func(t *testing.T) {
		testCaseAPIHandler[TokenRequest, any]{
			name:     "non-email hits default",
			method:   http.MethodPost,
			endpoint: TokensPath,
			header: http.Header{
				AppIDHeader:     []string{appID.String()},
				AppSecretHeader: []string{testAppSecret},
			},
			request: &TokenRequest{Email: "not-an-email"},
			err:     ErrInvalidNotificationChannel,
		}.Run(t)
	})

	t.Run("template compose failure", func(t *testing.T) {
		original := login.Template
		defer func() { login.Template = original }()
		login.Template = email.EmailTemplate{
			HTML:  "",
			Plain: `Hi {{.Email}}. Token: {{.Token}`, // syntax error: missing }
		}
		testCaseAPIHandler[TokenRequest, any]{
			name:     "broken template",
			method:   http.MethodPost,
			endpoint: TokensPath,
			header: http.Header{
				AppIDHeader:     []string{appID.String()},
				AppSecretHeader: []string{testAppSecret},
			},
			request: &TokenRequest{Email: testUserEmail},
			err:     ErrGenerateNotification,
		}.Run(t)
	})

	t.Run("queue push failure", func(t *testing.T) {
		port, err := xnet.SafeTestPort()
		if err != nil {
			t.Fatal(err)
		}
		// Zero-size queue: Push always returns ErrQueueFull
		eq, err := notification.NewQueue(t.Context(), 0)
		if err != nil {
			t.Fatal(err)
		}
		// no workers: push always returns ErrQueueFull on unbuffered channel
		eq.Start(0)
		defer eq.Stop()

		srv, err := New(t.Context(), &Config{
			Server:     testServerAddr,
			ServerPort: port,
			Secret:     testServerSecret,
		}, eq)
		if err != nil {
			t.Fatal(err)
		}
		go func() { _ = srv.Start() }()
		defer func() { _ = srv.Stop() }()

		// Wait for server to be ready
		for range 10 {
			if srv.Ping() {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		endpoint := fmt.Sprintf("http://%s:%d%s", testServerAddr, port, TokensPath)
		rawBody, _ := json.Marshal(&TokenRequest{Email: testUserEmail})
		req, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(rawBody))
		req.Header.Set(AppIDHeader, appID.String())
		req.Header.Set(AppSecretHeader, testAppSecret)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", resp.StatusCode)
		}
		var apiErr apiio.APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if apiErr.Code != ErrNotificationChannel.Code {
			t.Errorf("expected code %d, got %d", ErrNotificationChannel.Code, apiErr.Code)
		}
	})
}
