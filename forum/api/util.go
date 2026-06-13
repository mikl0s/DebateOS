// Package api — shared utilities.
package api

import (
	"encoding/json"
	"io"
)

// jsonDecodeBody decodes the JSON body of an HTTP response into v.
// Limited to 1 MB to prevent unbounded memory allocation.
func jsonDecodeBody(body io.Reader, v interface{}) error {
	return json.NewDecoder(io.LimitReader(body, 1<<20)).Decode(v)
}
