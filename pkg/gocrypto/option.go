package gocrypto

import "crypto"

const (
	modeECB = "ECB"
	modeCBC = "CBC"
	modeCFB = "CFB"
	modeCTR = "CTR"
)

var (
	defaultAesKey = []byte("mKoF_pL,NjI9=I;w")
	defaultDesKey = []byte("VgY7*uHb")
	defaultMode   = "ECB"

	defaultRsaFormat   = "PKCS#1"
	defaultRsaHashType = crypto.SHA1
)

type aesOptions struct {
	aesKey []byte

	mode string
}

type AesOption func(*aesOptions)

func (o *aesOptions) apply(opts ...AesOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultAesOptions() *aesOptions {
	return &aesOptions{
		aesKey: defaultAesKey,
		mode:   defaultMode,
	}
}

func WithAesKey(key []byte) AesOption {
	return func(o *aesOptions) {
		o.aesKey = key
	}
}

func WithAesModeCBC() AesOption {
	return func(o *aesOptions) {
		o.mode = modeCBC
	}
}

func WithAesModeECB() AesOption {
	return func(o *aesOptions) {
		o.mode = modeECB
	}
}

func WithAesModeCFB() AesOption {
	return func(o *aesOptions) {
		o.mode = modeCFB
	}
}

func WithAesModeCTR() AesOption {
	return func(o *aesOptions) {
		o.mode = modeCTR
	}
}

type desOptions struct {
	desKey []byte
	mode   string
}

type DesOption func(*desOptions)

func (o *desOptions) apply(opts ...DesOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultDesOptions() *desOptions {
	return &desOptions{
		desKey: defaultDesKey,
		mode:   defaultMode,
	}
}

func WithDesKey(key []byte) DesOption {
	return func(o *desOptions) {
		o.desKey = key
	}
}

func WithDesModeCBC() DesOption {
	return func(o *desOptions) {
		o.mode = modeCBC
	}
}

func WithDesModeECB() DesOption {
	return func(o *desOptions) {
		o.mode = modeECB
	}
}

func WithDesModeCFB() DesOption {
	return func(o *desOptions) {
		o.mode = modeCFB
	}
}

func WithDesModeCTR() DesOption {
	return func(o *desOptions) {
		o.mode = modeCTR
	}
}

type rsaOptions struct {
	format string

	hashType crypto.Hash
}

type RsaOption func(*rsaOptions)

func (o *rsaOptions) apply(opts ...RsaOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultRsaOptions() *rsaOptions {
	return &rsaOptions{
		format:   defaultRsaFormat,
		hashType: defaultRsaHashType,
	}
}

func WithRsaFormatPKCS1() RsaOption {
	return func(o *rsaOptions) {
		o.format = pkcs1
	}
}

func WithRsaFormatPKCS8() RsaOption {
	return func(o *rsaOptions) {
		o.format = pkcs8
	}
}

func WithRsaHashTypeMd5() RsaOption {
	return func(o *rsaOptions) {
		o.hashType = crypto.MD5
	}
}

func WithRsaHashTypeSha1() RsaOption {
	return func(o *rsaOptions) {
		o.hashType = crypto.SHA1
	}
}

func WithRsaHashTypeSha256() RsaOption {
	return func(o *rsaOptions) {
		o.hashType = crypto.SHA256
	}
}

func WithRsaHashTypeSha512() RsaOption {
	return func(o *rsaOptions) {
		o.hashType = crypto.SHA512
	}
}

func WithRsaHashType(hash crypto.Hash) RsaOption {
	return func(o *rsaOptions) {
		o.hashType = hash
	}
}
