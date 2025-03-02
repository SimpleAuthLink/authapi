package token

import (
	"regexp"
	"time"
)

const (
	appDataSeparator   = "|"
	appNameMinLen      = 3
	appNameMaxLen      = 20
	redirectURIPattern = `^https?://(?:localhost|[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)+)(?::\d+)?(/[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;=]*)?$`
	redirectURIMaxLen  = 80
	minDuration        = 30 * time.Second
	maxDuration        = 180 * 24 * time.Hour
	tokenSeparator     = '.'
)

var uriRegexp = regexp.MustCompile(redirectURIPattern)
