package email

import (
	"bufio"
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// domainRgx is the regular expression used to validate a domain.
var domainRgx = regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)

// LoadRemoteDisposableDomains loads a list of disposable domains from a remote
// source url. It reads the content of the source url line by line and parses
// each line as a domain. It returns a list of disposable domains or an error if
// something fails.
func LoadRemoteDisposableDomains(ctx context.Context, disposableSrc string) ([]string, error) {
	internalCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// prepare the request
	req, err := http.NewRequestWithContext(internalCtx, http.MethodGet, disposableSrc, nil)
	if err != nil {
		return nil, errors.Join(ErrLoadingDisposableDomains, err)
	}
	// perform the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Join(ErrLoadingDisposableDomains, err)
	}
	// read the response body line by line
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	var domains []string
	for scanner.Scan() {
		domain := scanner.Text()
		if domainRgx.MatchString(domain) {
			domains = append(domains, domain)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Join(ErrLoadingDisposableDomains, err)
	}
	return domains, nil
}

// CheckEmail checks if the email address is valid. It compares the domain with
// a list of disallowed domains. It returns true if the email address is valid,
// otherwise it returns false.
func CheckEmail(disallowedDomains []string, email string) bool {
	if len(disallowedDomains) == 0 {
		return true
	}
	// split the email address
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	// check the domain
	for _, domain := range disallowedDomains {
		if domain == parts[1] {
			return false
		}
	}
	return true
}
