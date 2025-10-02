package wcipher

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// gcmCipher implements the Cipher interface for AES-GCM mode
type gcmCipher struct {
	aesgcm cipher.AEAD
}

// NewGCMCipher creates a new GCM cipher
func NewGCMCipher(block cipher.Block) (Cipher, error) {
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &gcmCipher{aesgcm: aesgcm}, nil
}

// Encrypt encrypts the plaintext using AES-GCM
// Returns base64 encoded string with format: base64(ciphertext + nonce)
func (g *gcmCipher) Encrypt(plaintext []byte) []byte {
	// Generate random 12-byte nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		// If random generation fails, we can't proceed
		return nil
	}

	// Encrypt the data
	ciphertext := g.aesgcm.Seal(nil, nonce, plaintext, nil)

	// Append nonce to ciphertext
	result := append(ciphertext, nonce...)

	// // Return base64 encoded result
	// encoded := make([]byte, base64.StdEncoding.EncodedLen(len(result)))
	// base64.StdEncoding.Encode(encoded, result)
	return result
}

// Decrypt decrypts the base64 encoded ciphertext using AES-GCM
// Expects format: base64(ciphertext + nonce)
func (g *gcmCipher) Decrypt(ciphertext []byte) []byte {
	// // Decode base64
	// decoded := make([]byte, base64.StdEncoding.DecodedLen(len(ciphertext)))
	// n, err := base64.StdEncoding.Decode(decoded, ciphertext)
	// if err != nil {
	// 	return nil
	// }
	// decoded = decoded[:n]

	// Extract nonce (last 12 bytes)
	if len(ciphertext) < 12 {
		return nil
	}
	nonce := ciphertext[len(ciphertext)-12:]
	ciphertextOnly := ciphertext[:len(ciphertext)-12]

	// Decrypt the data
	plaintext, err := g.aesgcm.Open(nil, nonce, ciphertextOnly, nil)
	if err != nil {
		return nil
	}

	return plaintext
}
