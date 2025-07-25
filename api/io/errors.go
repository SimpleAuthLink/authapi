package io

import "fmt"

var (
	ErrNilRequest            = fmt.Errorf("nil request")
	ErrNilRequestBody        = fmt.Errorf("nil request body")
	ErrNotInitializedRequest = fmt.Errorf("request not initialized")
	ErrEmptyRequestBody      = fmt.Errorf("empty request body")
	ErrReadBody              = fmt.Errorf("failed to read request body")
	ErrEncodeBody            = fmt.Errorf("failed to encode request body")
	ErrDecodeBody            = fmt.Errorf("failed to decode request body")
	ErrHTTPStatusNotOk       = fmt.Errorf("no HTTP status OK")
)
