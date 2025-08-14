package client

import (
	"net/http"
	"net/mail"
	"net/url"
	"time"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/api/io"
	"github.com/simpleauthlink/authapi/token"
)

const (
	// DefaultClientTimeout is the default timeout for the client requests. It
	// can be overridden in the Config struct.
	DefaultClientTimeout = 30 * time.Second
	// DefaultAPIEndpoint is the default API endpoint for the client. It can be
	// overridden in the Config struct.
	DefaultAPIEndpoint       = "https://api.simpleauth.link"
	DefaultAuthTokenHeader   = "X-Auth-Token"
	DefaultAuthEmailHeader   = "X-Auth-User"
	DefaultAuthTokenURLParam = "token"
	DefaultAuthEmailURLParam = "user"
)

// Config holds the configuration for the client. It includes the API endpoint,
// application ID, application secret, and the timeout for requests. All fields
// are optional, but some of them are required for certain operations. If no
// API endpoint or timeout is provided, the default values will be used.
type Config struct {
	APIEndpoint string
	AppID       *token.AppID
	AppSecret   string
	Timeout     time.Duration
}

// init method initializes the Config struct with default values if they are
// not set.
func (c *Config) init() {
	if c.APIEndpoint == "" {
		c.APIEndpoint = DefaultAPIEndpoint
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultClientTimeout
	}
}

// Validate method checks if the configuration is valid. It returns an error if
// the configuration is invalid. It receives a boolean parameter to indicate if
// it should validate or not the application ID and secret. If the parameter is
// true, it will check if the AppID is not nil and the AppSecret is not empty.
func (c *Config) Validate(requireApp bool) error {
	if c.APIEndpoint == "" {
		return ErrInvalidAPIEndpoint
	}
	if _, err := url.ParseRequestURI(c.APIEndpoint); err != nil {
		return ErrInvalidAPIEndpoint
	}
	if c.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	if requireApp {
		if c.AppID == nil {
			return ErrInvalidAppID
		}
		if c.AppSecret == "" {
			return ErrInvalidAppSecret
		}
	}
	return nil
}

// Client is the main struct for the API client. It holds the configuration and
// allows to make requests to the API. It is initialized with a Config struct
// that contains the API endpoint, application ID, application secret, and the
// timeout for requests. The client can be used to create new application IDs,
// request tokens, and verify tokens. It also checks if the API is reachable
// during initialization.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// New method creates a new Client instance with the provided configuration.
// It initializes the configuration with default values if not set, checks if
// the API is reachable, and returns a new Client instance. If the API is not
// reachable, it returns an error.
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	// initialize the configuration with default values if not set
	cfg.init()
	// create a new HTTP client
	httpClient := &http.Client{Timeout: cfg.Timeout}
	// check if the api is reachable
	r, err := req[any](cfg, http.MethodGet, api.HealthCheckPath, nil)
	if err != nil {
		return nil, err
	}
	// make a request to the health check endpoint
	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, ErrAPIUnavailable
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	// check if the response is OK
	if resp.StatusCode != http.StatusOK {
		return nil, ErrAPIUnavailable
	}
	// if everything is fine, return a new client instance
	return &Client{
		config:     cfg,
		httpClient: httpClient,
	}, nil
}

// Default method creates a new Client instance with the default configuration
// and the provided application ID and secret. Both inputs are required. The
// application ID is expected as a string to prevent the user from converting
// it.
func Default(strAppID, appSecret string) (*Client, error) {
	if appSecret == "" {
		return nil, ErrInvalidAppSecret
	}
	if strAppID == "" {
		return nil, ErrInvalidAppID
	}
	appID := new(token.AppID).SetString(strAppID)
	if !new(token.App).SetID(appID).Valid(nil) {
		return nil, ErrInvalidAppID
	}
	return New(&Config{
		APIEndpoint: DefaultAPIEndpoint,
		AppID:       appID,
		AppSecret:   appSecret,
		Timeout:     DefaultClientTimeout,
	})
}

// NewAppID method creates a new AppID with the provided name, redirect URI,
// secret, and session duration. It requires the client configuration to be
// valid, including the application ID and secret. The name, redirect URI,
// secret, and session duration must be valid. If the configuration is not
// valid, it returns an error. If the request is successful, it returns a new
// AppID instance with the generated ID. If the request fails, it returns an
// error.
func (c *Client) NewAppID(name, redirectURI, secret string, sessionDuration time.Duration) (*token.AppID, error) {
	// validate the configuration before making the request, for this operation
	// the app ID and secret are not required
	if err := c.config.Validate(false); err != nil {
		return nil, err
	}
	// check if the parameters are valid
	if name == "" {
		return nil, ErrInvalidAppName
	} else if redirectURI == "" {
		return nil, ErrInvalidAppRedirectURI
	} else if secret == "" {
		return nil, ErrInvalidAppSecret
	} else if sessionDuration <= 0 {
		return nil, ErrInvalidAppSessionDuration
	}
	// create the request data
	data := &api.AppIDRequest{
		Name:        name,
		RedirectURL: redirectURI,
		Secret:      secret,
		Duration:    sessionDuration.String(),
	}
	// generate the request
	req, err := req(c.config, http.MethodPost, api.AppsPath, data)
	if err != nil {
		return nil, err
	}
	// make the request to the API
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrRequestAppID
	}
	// read the response data, it also checks if the response is OK and closes
	// the response body
	resData, err := new(io.Response[api.AppIDResponse]).Read(res)
	if err != nil {
		return nil, err
	}
	// parse and return the AppID
	return new(token.AppID).SetString(resData.ID), nil
}

// SetupAppID sets the application ID and secret in the client configuration.
func (c *Client) SetupAppID(appID *token.AppID, appSecret string) {
	c.config.AppID = appID
	c.config.AppSecret = appSecret
}

// RequestToken method requests a new token for the given email address. It
// requires the client configuration to be valid, including the application
// ID and secret. The email address must be a valid email format.
func (c *Client) RequestToken(email string) error {
	// validate the configuration before making the request, for this operation
	// the app ID and secret are required
	if err := c.config.Validate(true); err != nil {
		return err
	}
	// check if the email is valid
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmailAddress
	}
	// create the request with the email data
	data := &api.TokenRequest{Email: email}
	req, err := req(c.config, http.MethodPost, api.TokensPath, data)
	if err != nil {
		return err
	}
	// make the request to the API
	res, err := c.httpClient.Do(req)
	if err != nil {
		return ErrAPIUnavailable
	}
	defer func() {
		_ = res.Body.Close()
	}()
	// check if the response is OK
	if res.StatusCode != http.StatusOK {
		return ErrRequestToken
	}
	return nil
}

// VerifyToken method verifies the provided token for the given email address.
// It requires the client configuration to be valid, including the application
// ID and secret. The token must be valid, and the email address must be a
// valid email format. If the token is valid, it returns true and the
// expiration time. If the token is invalid or the email address is not valid,
// it returns false. If the request fails, it also returns an error.
func (c *Client) VerifyToken(token *token.Token, email string) (bool, time.Time, error) {
	// validate the configuration before making the request, for this operation
	// the app ID and secret are required
	if err := c.config.Validate(true); err != nil {
		return false, time.Time{}, err
	}
	// check if the token is valid
	if token == nil {
		return false, time.Time{}, ErrInvalidToken
	}
	// check if the email is valid
	if _, err := mail.ParseAddress(email); err != nil {
		return false, time.Time{}, ErrInvalidEmailAddress
	}
	// create the request with the token and email data
	data := &api.TokenStatusRequest{
		Token: token.String(),
		Email: email,
	}
	req, err := req(c.config, http.MethodPut, api.TokensPath, data)
	if err != nil {
		return false, time.Time{}, err
	}
	// make the request to the API
	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, time.Time{}, ErrAPIUnavailable
	}
	// decode the response data, it also checks if the response is OK and
	// closes the response body
	resData, err := new(io.Response[api.TokenStatusResponse]).Read(res)
	if err != nil {
		return false, time.Time{}, err
	}
	// return the verification result and expiration time
	return resData.Valid, resData.Expiration, nil
}

// AuthorizedRequestURLParams method checks if the request is authorized by
// verifying the token and email from the URL params of the request using the
// current client instance. It receives the request to check, and returns the
// user email in the request and a boolean that indicates if the token in the
// request is currently valid. If the client configuration is invalid, the
// request does not contains the required information or the validation fails,
// it returns an error.
func (c *Client) AuthorizedRequestURLParams(r *http.Request) (string, bool, error) {
	if err := c.config.Validate(true); err != nil {
		return "", false, err
	}
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		// map it to your sentinel
		return "", false, ErrMissingAuthURLParams
	}
	// get the strToken and email from the request headers
	strToken, _ := url.QueryUnescape(params.Get(DefaultAuthTokenURLParam))
	userEmail, _ := url.QueryUnescape(params.Get(DefaultAuthEmailURLParam))
	// check if the token and email are valid
	if strToken == "" || userEmail == "" {
		return "", false, ErrMissingAuthURLParams
	}
	userToken := new(token.Token).SetString(strToken)
	valid, _, err := c.VerifyToken(userToken, userEmail)
	if err != nil {
		return "", false, err
	}
	return userEmail, valid, nil
}

// AuthorizedRequestHeaders method checks if the request is authorized by
// verifying the token and email from the headers of the request using the
// current client instance. It receives the request to check, and returns the
// user email in the request and a boolean that indicates if the token in the
// request is currently valid. If the client configuration is invalid, the
// request does not contains the required information or the validation fails,
// it returns an error.
func (c *Client) AuthorizedRequestHeaders(r *http.Request) (string, bool, error) {
	if err := c.config.Validate(true); err != nil {
		return "", false, err
	}
	// get the strToken and email from the request headers
	strToken := r.Header.Get(DefaultAuthTokenHeader)
	userEmail := r.Header.Get(DefaultAuthEmailHeader)
	// check if the token and email are valid
	if strToken == "" || userEmail == "" {
		return "", false, ErrMissingAuthHeaders
	}
	userToken := new(token.Token).SetString(strToken)
	valid, _, err := c.VerifyToken(userToken, userEmail)
	if err != nil {
		return "", false, err
	}
	return userEmail, valid, nil
}

// req internal method creates a new HTTP request with the provided config,
// method, path, and data. It validates the configuration and sets the
// necessary headers. If the data is provided, it writes the JSON data to the
// request body. It returns the created request or an error if something goes
// wrong. This method is used internally by the client to create requests for
// the API.
func req[T any](config *Config, method, path string, data *T) (*http.Request, error) {
	if err := config.Validate(false); err != nil {
		return nil, err
	}
	// create the request URL
	finalEndpoint, err := url.JoinPath(config.APIEndpoint, path)
	if err != nil {
		return nil, ErrInvalidAPIEndpoint
	}
	// create the request with the method and URL
	req, err := http.NewRequest(method, finalEndpoint, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	// set the AppID and AppSecret headers if the configuration contains them
	if err := config.Validate(true); err == nil {
		req.Header.Set(api.AppIDHeader, config.AppID.String())
		req.Header.Set(api.AppSecretHeader, config.AppSecret)
	}
	// set the body and content type if data is provided
	if data != nil {
		// write json data if provided
		if err := io.RequestWith(&data).WriteJSON(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}
