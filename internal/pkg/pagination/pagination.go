package pagination

import (
	"bytes"
	"encoding/gob"

	"google.golang.org/protobuf/proto"
)

// Pagination is a helper struct for implementing cursor-based pagination.
type Pagination struct {
	Cursor      string
	ListRequest proto.Message
}

// ToToken returns a next page token.
func ToToken(p Pagination) (string, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(p); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// FromToken returns a Pagination struct from a next page token.
func FromToken(token string) (Pagination, error) {
	var p Pagination
	if err := gob.NewDecoder(bytes.NewBufferString(token)).Decode(&p); err != nil {
		return Pagination{}, err
	}
	return p, nil
}
