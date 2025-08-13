package encoding

import (
	"encoding"
	"errors"
	"reflect"
	"strings"
)

var (
	ErrNotAPointer = errors.New("v argument must be a pointer")
)

type Codec interface {
	Marshal(v interface{}) ([]byte, error)

	Unmarshal(data []byte, v interface{}) error

	Name() string
}

var registeredCodecs = make(map[string]Codec)

func RegisterCodec(codec Codec) {
	if codec == nil {
		panic("cannot register a nil Codec")
	}
	if codec.Name() == "" {
		panic("cannot register Codec with empty string result for Name()")
	}
	contentSubtype := strings.ToLower(codec.Name())
	registeredCodecs[contentSubtype] = codec
}

func GetCodec(contentSubtype string) Codec {
	return registeredCodecs[contentSubtype]
}

type Encoding interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

func Marshal(e Encoding, v interface{}) (data []byte, err error) {
	if !isPointer(v) {
		return data, ErrNotAPointer
	}
	bm, ok := v.(encoding.BinaryMarshaler)
	if ok && e == nil {
		return bm.MarshalBinary()
	}

	data, err = e.Marshal(v)
	if err == nil {
		return data, err
	}
	if ok {
		data, err = bm.MarshalBinary()
	}

	return data, err
}

func Unmarshal(e Encoding, data []byte, v interface{}) (err error) {
	if !isPointer(v) {
		return ErrNotAPointer
	}
	bm, ok := v.(encoding.BinaryUnmarshaler)
	if ok && e == nil {
		err = bm.UnmarshalBinary(data)
		return err
	}
	err = e.Unmarshal(data, v)
	if err == nil {
		return err
	}
	if ok {
		return bm.UnmarshalBinary(data)
	}
	return err
}

func isPointer(data interface{}) bool {
	switch reflect.ValueOf(data).Kind() {
	case reflect.Ptr, reflect.Interface:
		return true
	default:
		return false
	}
}
