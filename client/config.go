package client

import (
	"fmt"
	"net/url"

	"github.com/simpleauthlink/authapi/helpers"
)

// ClientConfig struct represents the configuration needed to use the client.
type ClientConfig struct {
	// APIEndpoint is the API hostname.
	APIEndpoint string
	url         *url.URL
	// Secret is the app secret on the API server.
	Secret string
}

// check function validates the configuration and returns an error if the
// configuration is invalid. It checks if the configuration is nil, if the
// secret is empty, and if the API endpoint is invalid. If the API endpoint is
// empty, it uses the default API endpoint. It returns an error if the
// configuration is nil, the secret is empty or the API endpoint is invalid.
func (conf *ClientConfig) check() error {
	if conf == nil {
		return fmt.Errorf("config is required")
	}
	if conf.APIEndpoint == "" {
		conf.APIEndpoint = helpers.DefaultAPIEndpoint
	}
	if conf.Secret == "" {
		return fmt.Errorf("secret is required")
	}
	var err error
	conf.url, err = url.Parse(conf.APIEndpoint)
	if err != nil {
		return fmt.Errorf("invalid API endpoint: %w", err)
	}
	return nil
}
