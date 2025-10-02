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
// Returns ciphertext with nonce appended: ciphertext + nonce
func (g *gcmCipher) Encrypt(plaintext []byte) []byte {
	// Generate random 12-byte nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil
	}

	// Encrypt the data (GCM can handle empty plaintext)
	ciphertext := g.aesgcm.Seal(nil, nonce, plaintext, nil)

	// Append nonce to ciphertext
	result := append(ciphertext, nonce...)

	return result
}

// Decrypt decrypts the ciphertext using AES-GCM
// Expects format: ciphertext + nonce
func (g *gcmCipher) Decrypt(ciphertext []byte) []byte {
	// Extract nonce (last 12 bytes)
	if len(ciphertext) < 12 {
		// Return nil to indicate failure (different from empty slice)
		return nil
	}
	nonce := ciphertext[len(ciphertext)-12:]
	ciphertextOnly := ciphertext[:len(ciphertext)-12]

	// Decrypt the data
	plaintext, err := g.aesgcm.Open(nil, nonce, ciphertextOnly, nil)
	if err != nil {
		// Return nil to indicate authentication failure (different from empty slice)
		return nil
	}

	// Return the plaintext (which can be empty for empty input)
	// If plaintext is nil (from aesgcm.Open), return empty slice instead
	if plaintext == nil {
		return []byte{}
	}
	return plaintext
}
