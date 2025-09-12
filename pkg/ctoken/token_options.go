package ctoken

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
)

type Options struct {
	CacheMode        int8
	CachePreKey      string
	Timeout          int64
	MaxRefresh       int64
	MaxRefreshTimes  int
	TokenDelimiter   string
	EncryptKey       []byte
	MultiLogin       bool
	AuthExcludePaths g.SliceStr
}

func (o *Options) String() string {
	return fmt.Sprintf("Options{"+
		"CacheMode:%d, CachePreKey:%s, Timeout:%d, MaxRefresh:%d"+
		", TokenDelimiter:%s, MultiLogin:%v, AuthExcludePaths:%v"+
		"}", o.CacheMode, o.CachePreKey, o.Timeout, o.MaxRefresh,
		o.TokenDelimiter, o.MultiLogin, o.AuthExcludePaths)
}
