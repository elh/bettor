package pagination

import (
	"bytes"
	"encoding/gob"

	"google.golang.org/protobuf/proto"
)

type Pagination struct {
	Cursor      string
	ListRequest proto.Message
}

func ToToken(p Pagination) (string, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(p); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func FromToken(token string) (Pagination, error) {
	var p Pagination
	if err := gob.NewDecoder(bytes.NewBufferString(token)).Decode(&p); err != nil {
		return Pagination{}, err
	}
	return p, nil
}
