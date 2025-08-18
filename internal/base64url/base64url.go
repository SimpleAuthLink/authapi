// base64url package provides a way to encode and decode data using a modified
// base64 encoding scheme that is safe for URLs.
package base64url

import (
	"bytes"
	"encoding/base64"
)

// RawEncode function encodes the input byte slice into a base64 URL-safe
// encoded byte slice. It encodes the input using the raw standard base64
// encoding and then replaces the URL-unsafe characters.
func RawEncode(src []byte) []byte {
	rawBase64 := make([]byte, base64.RawStdEncoding.EncodedLen(len(src)))
	base64.RawStdEncoding.Encode(rawBase64, src)
	// replace the characters that are not URL safe (+, /)
	rawBase64 = bytes.ReplaceAll(rawBase64, []byte("+"), []byte("-"))
	rawBase64 = bytes.ReplaceAll(rawBase64, []byte("/"), []byte("_"))
	return rawBase64
}

// RawDecode function decodes the input byte slice from a base64 URL-safe
// encoded byte slice. It replaces the URL-unsafe characters with their
// standard base64 counterparts and then decodes the result using the raw
// standard base64 decoding.
func RawDecode(data []byte) ([]byte, error) {
	// recover the characters that are not URL safe (+, /)
	rawBase64 := bytes.ReplaceAll(data, []byte("-"), []byte("+"))
	rawBase64 = bytes.ReplaceAll(rawBase64, []byte("_"), []byte("/"))
	// return the decoded string from basic base64
	res := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	if _, err := base64.RawStdEncoding.Decode(res, rawBase64); err != nil {
		return nil, err
	}
	return res, nil
}
