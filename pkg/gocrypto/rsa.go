package gocrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
)

const (
	pkcs1 = "PKCS#1"
	pkcs8 = "PKCS#8"
)

func RsaEncrypt(publicKey []byte, rawData []byte, opts ...RsaOption) ([]byte, error) {
	o := defaultRsaOptions()
	o.apply(opts...)

	return rsaEncryptWithPublicKey(publicKey, rawData)
}

func RsaDecrypt(privateKey []byte, cipherData []byte, opts ...RsaOption) ([]byte, error) {
	o := defaultRsaOptions()
	o.apply(opts...)

	return rsaDecryptWithPrivateKey(privateKey, cipherData, o.format)
}

func RsaEncryptHex(publicKey []byte, rawData []byte, opts ...RsaOption) (string, error) {
	o := defaultRsaOptions()
	o.apply(opts...)

	cipherData, err := rsaEncryptWithPublicKey(publicKey, rawData)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(cipherData), nil
}

func RsaDecryptHex(privateKey []byte, cipherHex string, opts ...RsaOption) (string, error) {
	o := defaultRsaOptions()
	o.apply(opts...)

	cipherData, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	rawData, err := rsaDecryptWithPrivateKey(privateKey, cipherData, o.format)
	if err != nil {
		return "", err
	}

	return string(rawData), nil
}

func RsaSign(privateKey []byte, rawData []byte, opts ...RsaOption) ([]byte, error) {
	o := defaultRsaOptions()
	o.apply(opts...)

	return rsaSignWithPrivateKey(privateKey, o.hashType, rawData, o.format)
}

func RsaVerify(publicKey []byte, rawData []byte, signData []byte, opts ...RsaOption) error {
	o := defaultRsaOptions()
	o.apply(opts...)

	return rsaVerifyWithPublicKey(publicKey, o.hashType, rawData, signData)
}

func RsaSignBase64(privateKey []byte, rawData []byte, opts ...RsaOption) (string, error) {
	o := defaultRsaOptions()
	o.apply(opts...)

	cipherData, err := rsaSignWithPrivateKey(privateKey, o.hashType, rawData, o.format)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipherData), nil
}

func RsaVerifyBase64(publicKey []byte, rawData []byte, signBase64 string, opts ...RsaOption) error {
	o := defaultRsaOptions()
	o.apply(opts...)

	signData, err := base64.StdEncoding.DecodeString(signBase64)
	if err != nil {
		return err
	}

	return rsaVerifyWithPublicKey(publicKey, o.hashType, rawData, signData)
}

func rsaEncryptWithPublicKey(publicKey []byte, rawData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key is not pem format")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	prk, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("it's not a public key")
	}

	return rsa.EncryptPKCS1v15(rand.Reader, prk, rawData)
}

func rsaDecryptWithPrivateKey(privateKey []byte, cipherData []byte, format string) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key is not pem format")
	}

	prk, err := getPrivateKey(block.Bytes, format)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, prk, cipherData)
}

func rsaSignWithPrivateKey(privateKey []byte, hash crypto.Hash, rawData []byte, format string) ([]byte, error) {
	if !hash.Available() {
		return nil, errors.New("not supported hash type")
	}

	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key is not pem format")
	}

	prk, err := getPrivateKey(block.Bytes, format)
	if err != nil {
		return nil, err
	}

	h := hash.New()
	_, err = h.Write(rawData)
	if err != nil {
		return nil, err
	}
	hashed := h.Sum(nil)

	return rsa.SignPKCS1v15(rand.Reader, prk, hash, hashed)
}

func rsaVerifyWithPublicKey(publicKey []byte, hash crypto.Hash, rawData []byte, signData []byte) (err error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return errors.New("public key is not pem format")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	prk, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("it's not a public key")
	}

	h := hash.New()
	_, err = h.Write(rawData)
	if err != nil {
		return err
	}
	hashed := h.Sum(nil)

	return rsa.VerifyPKCS1v15(prk, hash, hashed, signData)
}

func getPrivateKey(der []byte, format string) (*rsa.PrivateKey, error) {
	var prk *rsa.PrivateKey
	switch format {
	case pkcs1:
		var err error
		prk, err = x509.ParsePKCS1PrivateKey(der)
		if err != nil {
			return nil, err
		}

	case pkcs8:
		priv, err := x509.ParsePKCS8PrivateKey(der)
		if err != nil {
			return nil, err
		}
		var ok bool
		prk, ok = priv.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("it's not a private key")
		}

	default:
		return nil, errors.New("unknown format = " + format)
	}

	return prk, nil
}
