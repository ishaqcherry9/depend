package gocrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"

	"github.com/ishaqcherry9/depend/pkg/gocrypto/wcipher"
)

func AesEncrypt(rawData []byte, opts ...AesOption) ([]byte, error) {
	o := defaultAesOptions()
	o.apply(opts...)

	return aesEncryptByMode(o.mode, rawData, o.aesKey)
}

func AesDecrypt(cipherData []byte, opts ...AesOption) ([]byte, error) {
	o := defaultAesOptions()
	o.apply(opts...)

	return aesDecryptByMode(o.mode, cipherData, o.aesKey)
}

func AesEncryptHex(rawData string, opts ...AesOption) (string, error) {
	o := defaultAesOptions()
	o.apply(opts...)

	cipherData, err := aesEncryptByMode(o.mode, []byte(rawData), o.aesKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(cipherData), nil
}

func AesDecryptHex(cipherStr string, opts ...AesOption) (string, error) {
	o := defaultAesOptions()
	o.apply(opts...)

	cipherData, err := hex.DecodeString(cipherStr)
	if err != nil {
		return "", err
	}

	rawData, err := aesDecryptByMode(o.mode, cipherData, o.aesKey)
	if err != nil {
		return "", err
	}

	return string(rawData), nil
}

func getCipherMode(mode string) (wcipher.CipherMode, error) {
	var cipherMode wcipher.CipherMode
	switch mode {
	case modeECB:
		cipherMode = wcipher.NewECBMode()
	case modeCBC:
		cipherMode = wcipher.NewCBCMode()
	case modeCFB:
		cipherMode = wcipher.NewCFBMode()
	case modeCTR:
		cipherMode = wcipher.NewCTRMode()
	default:
		return nil, errors.New("unknown mode = " + mode)
	}

	return cipherMode, nil
}

func aesEncryptByMode(mode string, rawData []byte, key []byte) ([]byte, error) {
	cipherMode, err := getCipherMode(mode)
	if err != nil {
		return nil, err
	}

	cip, err := wcipher.NewAESWith(key, cipherMode)
	if err != nil {
		return nil, err
	}

	return cip.Encrypt(rawData), nil
}

func aesDecryptByMode(mode string, cipherData []byte, key []byte) ([]byte, error) {
	cipherMode, err := getCipherMode(mode)
	if err != nil {
		return nil, err
	}

	cip, err := wcipher.NewAESWith(key, cipherMode)
	if err != nil {
		return nil, err
	}

	return cip.Decrypt(cipherData), nil
}

// AES-GCM 加密
func encryptAES(plaintext []byte, key []byte) (string, error) {
	// 随机生成12字节nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Encrypt the data
	// The additionalData parameter can be used to authenticate additional data that is not encrypted
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// 拼接nonce在密文后面返回
	result := append(ciphertext, nonce...)
	// 返回base64编码的密文
	return base64.StdEncoding.EncodeToString(result), nil

}

// AES-GCM 解密
func decryptAES(plaintextEncoded string, key []byte) ([]byte, error) {
	// 解码base64编码的密文
	ciphertext, err := base64.StdEncoding.DecodeString(plaintextEncoded)
	if err != nil {
		return nil, err
	}
	// 提取nonce
	nonce := ciphertext[len(ciphertext)-12:]
	ciphertext = ciphertext[:len(ciphertext)-12]
	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM cipher mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	// The same nonce must be used for decryption as for encryption
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
