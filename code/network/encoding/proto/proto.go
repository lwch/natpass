package proto

import (
	"fmt"

	"github.com/lwch/natpass/code/network/encoding"
	"google.golang.org/protobuf/proto"
)

type codec struct{}

// New create protobuf codec
func New() encoding.Codec {
	return &codec{}
}

// Marshal protobuf marshal
func (*codec) Marshal(v interface{}) ([]byte, error) {
	vv, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("invalid value type, want proto.Message, got %T", v)
	}
	return proto.Marshal(vv)
}

// Unmarshal protobuf unmarshal
func (*codec) Unmarshal(data []byte, v interface{}) error {
	vv, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("invalid value type, want proto.Message, got %T", v)
	}
	return proto.Unmarshal(data, vv)
}
