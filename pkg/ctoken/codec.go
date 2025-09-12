package ctoken

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"
)

type Encoder interface {
	Encode(ctx context.Context, userKey string) (token string, err error)
}

type Decoder interface {
	Decrypt(ctx context.Context, token string) (userKey string, err error)
}

type Codec interface {
	Encoder
	Decoder
}

type DefaultCodec struct {
	Delimiter  string
	EncryptKey []byte
}

func NewDefaultCodec(delimiter string, encryptKey []byte) *DefaultCodec {
	return &DefaultCodec{
		Delimiter:  delimiter,
		EncryptKey: encryptKey,
	}
}

func (c *DefaultCodec) Encode(ctx context.Context, userKey string) (token string, err error) {
	if userKey == "" {
		return "", errors.New(MsgErrUserKeyEmpty)
	}
	randStr, err := gmd5.Encrypt(grand.Letters(10))
	if err != nil {
		return "", err
	}
	encryptBeforeStr := userKey + c.Delimiter + randStr
	encryptByte, err := gaes.Encrypt([]byte(encryptBeforeStr), c.EncryptKey)
	if err != nil {
		return "", err
	}
	return gbase64.EncodeToString(encryptByte), nil
}

func (c *DefaultCodec) Decrypt(ctx context.Context, token string) (userKey string, err error) {
	if token == "" {
		return "", errors.New(MsgErrTokenEmpty)
	}
	token64, err := gbase64.Decode([]byte(token))
	if err != nil {
		return "", err
	}
	decryptStr, err := gaes.Decrypt(token64, c.EncryptKey)
	if err != nil {
		return "", err
	}
	decryptArray := gstr.Split(string(decryptStr), c.Delimiter)
	if len(decryptArray) < 2 {
		return "", errors.New(MsgErrTokenLen)
	}
	return decryptArray[0], nil
}
