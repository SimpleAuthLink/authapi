package token

import (
	"regexp"
	"time"
)

const (
	appDataSeparator   = "|"
	appNameMinLen      = 3
	appNameMaxLen      = 20
	redirectURIPattern = `^https?://[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+(/[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;=]*)?$`
	redirectURIMaxLen  = 80
	minDuration        = 5 * time.Minute
	maxDuration        = 180 * 24 * time.Hour
)

var uriRegexp = regexp.MustCompile(redirectURIPattern)
