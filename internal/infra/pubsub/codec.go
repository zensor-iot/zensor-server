package pubsub

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Codec interface {
	Encode(value interface{}) (data []byte, err error)
	Decode(data []byte) (value interface{}, err error)
}

func newJSONCodec(prototype any) *JSONCodec {
	return &JSONCodec{prototype}
}

var _ Codec = &JSONCodec{}

type JSONCodec struct {
	prototype any
}

func (c *JSONCodec) Encode(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshaling data: %w", err)
	}

	return data, nil
}

func (c *JSONCodec) Decode(data []byte) (any, error) {
	pt := reflect.TypeOf(c.prototype)
	instance := reflect.New(pt).Interface()
	err := json.Unmarshal(data, &instance)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling data: %w", err)
	}

	return instance, nil
}
