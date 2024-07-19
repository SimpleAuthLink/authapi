package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/simpleauthlink/authapi/internal"
)

// ValidToken function validates the token provided using the API server. It
// returns true if the token is valid, false if the token is invalid, or an
// error if something goes wrong during the process. It receives the context,
// the token and the client configuration. The configuration must include, at
// least, the secret of your app. If the API endpoint is empty, it uses the
// default API endpoint. It validates the config and returns an error if the
// configuration is nil, the secret is empty or the API endpoint is invalid.
func ValidToken(ctx context.Context, token string, config *ClientConfig) (bool, error) {
	if err := config.check(); err != nil {
		return false, err
	}
	// add token to the query
	query := config.url.Query()
	query.Set(TokenQueryParam, token)
	// set the path and query
	config.url.Path = ValidateTokenPath
	config.url.RawQuery = query.Encode()
	// create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.url.String(), nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	// set the secret in the header
	req.Header.Set(AppSecretHeader, config.Secret)
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

// EncodeUserToken function encodes the user information into a token and
// returns it. It receives the app id and the email of the user and returns the
// token and the user id. If the app id or the email are empty, it returns an
// error. The token is composed of three parts separated by a token separator.
// The first part is a random sequence of 8 bytes encoded as a hexadecimal
// string. The second part is the app id and the third part is the user id. The
// user id is generated hashing the email with a length of 4 bytes. The token
// is returned following the token format:
//
//	[appId(8)]-[userId(8)]-[randomPart(16)]
func EncodeUserToken(appId, email string) (string, string, error) {
	// check if the app id and email are not empty
	if len(appId) == 0 || len(email) == 0 {
		return "", "", fmt.Errorf("appId and email are required")
	}
	bToken := internal.RandBytes(8)
	hexToken := hex.EncodeToString(bToken)
	// hash email
	userId, err := internal.Hash(email, 4)
	if err != nil {
		return "", "", err
	}
	return strings.Join([]string{appId, userId, hexToken}, TokenSeparator), userId, nil
}

// DecodeUserToken function decodes the user information from the token provided
// and returns the app id and the user id. If the token is invalid, it returns
// an error. It splits the provided token by the token separator and returns the
// second and third parts, which are the app id and the user id respectively,
// following the token format:
//
//	[appId(8)]-[userId(8)]-[randomPart(16)]
func DecodeUserToken(token string) (string, string, error) {
	tokenParts := strings.Split(token, TokenSeparator)
	if len(tokenParts) != 3 {
		return "", "", fmt.Errorf("invalid token")
	}
	return tokenParts[0], tokenParts[1], nil
}
