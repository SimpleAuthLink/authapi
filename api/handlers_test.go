package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/notification/email"
	"github.com/simpleauthlink/authapi/notification/templates/login"
	"github.com/simpleauthlink/authapi/token"
)

type testCaseAPIHandler[ReqType, ResType any] struct {
	name     string
	method   string
	endpoint string
	header   http.Header
	request  *ReqType
	response *ResType
	err      *APIError
}

func (testCase testCaseAPIHandler[Rq, Rs]) url() string {
	return fmt.Sprintf("%s%s", testServerApiURL, testCase.endpoint)
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
		defer resp.Body.Close()
		switch {
		case testCase.err != nil:
			if resp.StatusCode != testCase.err.statusCode {
				t.Fatalf("expected status code: %d, got: %d", testCase.err.statusCode, resp.StatusCode)
			}
			err := new(APIError)
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
			ID: testApp.ID().String(),
		},
	}.Run(t)
	testCaseAPIHandler[AppIDRequest, AppIDResponse]{
		name:     "no request",
		method:   http.MethodPost,
		endpoint: AppsPath,
		request:  nil,
		err:      DecodeAppIDRequestErr,
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
		err: InvalidAppIDErr,
	}.Run(t)
}

func TestRequestTokenAndStatusHandler(t *testing.T) {
	testApp := &token.App{
		Name:            testAppName,
		RedirectURI:     testAppRedirectURL,
		SessionDuration: testAppSessionDuration,
	}
	testAppID := testApp.ID()
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
		err:      InvalidAppHeadersErr,
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
		err:      InvalidAppIDErr,
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
		err:      InvalidAppHeadersErr,
	}.Run(t)
	testCaseAPIHandler[TokenRequest, any]{
		name:     "no email provided",
		method:   http.MethodPost,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{testAppID.String()},
			AppSecretHeader: []string{testAppSecret},
		},
		request: &TokenRequest{
			Email: "",
		},
		response: nil,
		err:      GenerateTokenErr,
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
		err:      DecodeTokenRequestErr,
	}.Run(t)

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
		break
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
			Email: testUserEmail,
		},
		response: &TokenStatusResponse{
			Valid:      true,
			Expiration: testToken.Expiration().Time(),
		},
	}.Run(t)

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
			Email: testUserEmail,
		},
		response: nil,
		err:      InvalidAppIDErr,
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
			Email: testUserEmail,
		},
		response: nil,
		err:      InvalidAppHeadersErr,
	}.Run(t)

	testCaseAPIHandler[TokenStatusRequest, any]{
		name:     "no headers",
		method:   http.MethodPut,
		endpoint: TokensPath,
		request: &TokenStatusRequest{
			Token: testToken.String(),
			Email: testUserEmail,
		},
		response: nil,
		err:      InvalidAppHeadersErr,
	}.Run(t)

	testCaseAPIHandler[any, any]{
		name:     "no request",
		method:   http.MethodPut,
		endpoint: TokensPath,
		header: http.Header{
			AppIDHeader:     []string{testAppID.String()},
			AppSecretHeader: []string{testAppSecret},
		},
		err: DecodeTokenStatusRequestErr,
	}.Run(t)
}
