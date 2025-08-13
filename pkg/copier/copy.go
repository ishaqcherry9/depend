package copier

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"
)

func Copy(dst interface{}, src interface{}) error {
	return copier.CopyWithOption(dst, src, Converter)
}

var Converter = copier.Option{
	DeepCopy: true,
	Converters: []copier.TypeConverter{
		{
			SrcType: time.Time{},
			DstType: copier.String,
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(time.Time)
				if !ok {
					return nil, fmt.Errorf("expected time.Time got %T", src)
				}
				return s.Format(time.RFC3339), nil
			},
		},

		{
			SrcType: &time.Time{},
			DstType: copier.String,
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(*time.Time)
				if !ok {
					return nil, fmt.Errorf("expected *time.Time got %T", src)
				}
				if s == nil {
					return "", nil
				}
				return s.Format(time.RFC3339), nil
			},
		},

		{
			SrcType: copier.String,
			DstType: time.Time{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(string)
				if !ok {
					return nil, fmt.Errorf("expected string got %T", src)
				}
				if s == "" {
					return time.Time{}, nil
				}
				return time.Parse(time.RFC3339, s)
			},
		},

		{
			SrcType: copier.String,
			DstType: &time.Time{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(string)
				if !ok {
					return nil, fmt.Errorf("expected string got %T", src)
				}
				if s == "" {
					return nil, nil
				}
				t, err := time.Parse(time.RFC3339, s)
				return &t, err
			},
		},
	},
}

func CopyDefault(dst interface{}, src interface{}) error {
	return copier.Copy(dst, src)
}

func CopyWithOption(dst interface{}, src interface{}, options copier.Option) error {
	return copier.CopyWithOption(dst, src, options)
}
