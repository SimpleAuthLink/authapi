package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/helpers"
)

// Client struct represents the client to interact with the API server. It
// contains the configuration of the client. The configuration includes the
// secret of the app and the API endpoint. The API endpoint is optional and if
// it is empty, it uses the default API endpoint. The client provides two
// methods to interact with the API server, RequestToken and ValidateToken.
type Client struct {
	config *ClientConfig
}

// New function creates a new client based on the provided configuration. It
// returns the client and an error if the configuration is invalid. The
// configuration must include, at least, the secret of your app. If the API
// endpoint is empty, it uses the default API endpoint. It validates the config
// and returns an error if the configuration is nil, the secret is empty or the
// API endpoint is invalid.
func New(config *ClientConfig) (*Client, error) {
	if err := config.check(); err != nil {
		return nil, err
	}
	return &Client{config: config}, nil
}

// RequestToken function requests a token for the user based on the provided
// email. It returns an error if the email is empty. It receives the context
// and the token request. The token request includes the email of the user, the
// redirect URL and the session duration. The session duration is optional and
// if it is zero, it uses the default session duration. It creates a new URL
// based on the API endpoint, encodes the request, creates the request, sets
// the secret in the header, sets the content type and makes the request. It
// checks the status code and returns an error if the status code is different
// from 200, if so returns an error trying to decode the body of the response.
func (cli *Client) RequestToken(ctx context.Context, req *api.TokenRequest) error {
	if req == nil || req.Email == "" {
		return fmt.Errorf("email is required to request a token")
	}
	// create a new URL based on the API endpoint
	url := new(url.URL)
	*url = *cli.config.url
	// set the path
	url.Path = helpers.UserEndpointPath
	// encode the request
	encodedReq, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error encoding request: %w", err)
	}
	// create the request
	buf := bytes.NewBuffer(encodedReq)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	// set the secret in the header
	httpReq.Header.Set(helpers.AppSecretHeader, cli.config.Secret)
	// set the content type
	httpReq.Header.Set("Content-Type", "application/json")
	// make the request
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()
	// check the status code and return an error if the status code is different
	// from 200, if so return an error trying to decode the body of the response
	if res.StatusCode != http.StatusOK {
		// decode body and return error
		msg, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code: %d", res.StatusCode)
		}
		return fmt.Errorf("unexpected response: [%d] %s", res.StatusCode, string(msg))
	}
	return nil
}

// ValidateToken function validates the token provided using the API server. It
// returns true if the token is valid, false if the token is invalid, or an
// error if something goes wrong during the process. It receives the context,
// the token and the client configuration. The configuration must include, at
// least, the secret of your app. If the API endpoint is empty, it uses the
// default API endpoint. It validates the config and returns an error if the
// configuration is nil, the secret is empty or the API endpoint is invalid.
func (cli *Client) ValidateToken(ctx context.Context, token string) (bool, error) {
	// create a new URL based on the API endpoint
	url := new(url.URL)
	*url = *cli.config.url
	// add token to the query
	query := url.Query()
	query.Set(helpers.TokenQueryParam, token)
	// set the path and query
	url.Path = helpers.UserEndpointPath
	url.RawQuery = query.Encode()
	// create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	// set the secret in the header
	req.Header.Set(helpers.AppSecretHeader, cli.config.Secret)
	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	// check the status code, return true if the status code is 200 or false if
	// the status code is 401, otherwise return an error trying to decode the
	// body of the response
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized:
		return false, nil
	default:
		// decode body and return error
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return false, fmt.Errorf("unexpected response: [%d] %s", resp.StatusCode, string(msg))
	}
}
