package mapping

import (
	"github.com/ishaqcherry9/depend/pkg/jsonx"
	"io"
)

const jsonTagKey = "json"

var jsonUnmarshaler = NewUnmarshaler(jsonTagKey)

func UnmarshalJsonBytes(content []byte, v any, opts ...UnmarshalOption) error {
	return unmarshalJsonBytes(content, v, getJsonUnmarshaler(opts...))
}

func UnmarshalJsonMap(m map[string]any, v any, opts ...UnmarshalOption) error {
	return getJsonUnmarshaler(opts...).Unmarshal(m, v)
}

func UnmarshalJsonReader(reader io.Reader, v any, opts ...UnmarshalOption) error {
	return unmarshalJsonReader(reader, v, getJsonUnmarshaler(opts...))
}

func getJsonUnmarshaler(opts ...UnmarshalOption) *Unmarshaler {
	if len(opts) > 0 {
		return NewUnmarshaler(jsonTagKey, opts...)
	}

	return jsonUnmarshaler
}

func unmarshalJsonBytes(content []byte, v any, unmarshaler *Unmarshaler) error {
	var m any
	if err := jsonx.Unmarshal(content, &m); err != nil {
		return err
	}

	return unmarshaler.Unmarshal(m, v)
}

func unmarshalJsonReader(reader io.Reader, v any, unmarshaler *Unmarshaler) error {
	var m any
	if err := jsonx.UnmarshalFromReader(reader, &m); err != nil {
		return err
	}

	return unmarshaler.Unmarshal(m, v)
}
