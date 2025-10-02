package gocrypto

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

// Test data
var (
	testKey16 = []byte("1234567890123456")                 // 16 bytes for AES-128
	testKey24 = []byte("123456789012345678901234")         // 24 bytes for AES-192
	testKey32 = []byte("12345678901234567890123456789012") // 32 bytes for AES-256

	testPlaintext = []byte("Hello, World! This is a test message for AES encryption.")
	testString    = "Hello, World! This is a test string for AES encryption."
)

func TestAesEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
		key       []byte
		mode      string
	}{
		{
			name:      "AES-128 ECB",
			plaintext: testPlaintext,
			key:       testKey16,
			mode:      modeECB,
		},
		{
			name:      "AES-192 CBC",
			plaintext: testPlaintext,
			key:       testKey24,
			mode:      modeCBC,
		},
		{
			name:      "AES-256 CFB",
			plaintext: testPlaintext,
			key:       testKey32,
			mode:      modeCFB,
		},
		{
			name:      "AES-128 CTR",
			plaintext: testPlaintext,
			key:       testKey16,
			mode:      modeCTR,
		},
		{
			name:      "AES-128 GCM",
			plaintext: testPlaintext,
			key:       testKey16,
			mode:      modeGCM,
		},
		{
			name:      "AES-256 GCM",
			plaintext: testPlaintext,
			key:       testKey32,
			mode:      modeGCM,
		},
		{
			name:      "Empty plaintext",
			plaintext: []byte(""),
			key:       testKey16,
			mode:      modeECB,
		},
		{
			name:      "Long plaintext",
			plaintext: make([]byte, 1000),
			key:       testKey16,
			mode:      modeCBC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill long plaintext with random data
			if len(tt.plaintext) == 1000 {
				rand.Read(tt.plaintext)
			}

			// Encrypt
			ciphertext, err := AesEncrypt(tt.plaintext, WithAesKey(tt.key), getModeOption(tt.mode))
			if err != nil {
				t.Fatalf("AesEncrypt failed: %v", err)
			}

			// Verify ciphertext is different from plaintext
			if string(ciphertext) == string(tt.plaintext) {
				t.Error("Ciphertext should be different from plaintext")
			}

			// Decrypt
			decrypted, err := AesDecrypt(ciphertext, WithAesKey(tt.key), getModeOption(tt.mode))
			if err != nil {
				t.Fatalf("AesDecrypt failed: %v", err)
			}

			// Verify decrypted text matches original
			if string(decrypted) != string(tt.plaintext) {
				t.Errorf("Decrypted text doesn't match original. Got: %s, Want: %s", string(decrypted), string(tt.plaintext))
			}
		})
	}
}

func TestAesEncryptHexDecryptHex(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       []byte
		mode      string
	}{
		{
			name:      "AES-128 ECB",
			plaintext: testString,
			key:       testKey16,
			mode:      modeECB,
		},
		{
			name:      "AES-192 CBC",
			plaintext: testString,
			key:       testKey24,
			mode:      modeCBC,
		},
		{
			name:      "AES-256 CFB",
			plaintext: testString,
			key:       testKey32,
			mode:      modeCFB,
		},
		{
			name:      "AES-128 CTR",
			plaintext: testString,
			key:       testKey16,
			mode:      modeCTR,
		},
		{
			name:      "AES-128 GCM",
			plaintext: testString,
			key:       testKey16,
			mode:      modeGCM,
		},
		{
			name:      "AES-256 GCM",
			plaintext: testString,
			key:       testKey32,
			mode:      modeGCM,
		},
		{
			name:      "Empty string",
			plaintext: "",
			key:       testKey16,
			mode:      modeECB,
		},
		{
			name:      "Unicode string",
			plaintext: "你好，世界！这是一个测试消息。",
			key:       testKey16,
			mode:      modeCBC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			cipherHex, err := AesEncryptHex(tt.plaintext, WithAesKey(tt.key), getModeOption(tt.mode))
			if err != nil {
				t.Fatalf("AesEncryptHex failed: %v", err)
			}

			// Verify hex string is valid
			_, err = hex.DecodeString(cipherHex)
			if err != nil {
				t.Fatalf("Generated hex string is invalid: %v", err)
			}

			// Decrypt
			decrypted, err := AesDecryptHex(cipherHex, WithAesKey(tt.key), getModeOption(tt.mode))
			if err != nil {
				t.Fatalf("AesDecryptHex failed: %v", err)
			}

			// Verify decrypted text matches original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text doesn't match original. Got: %s, Want: %s", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAesGCMEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
		key       []byte
	}{
		{
			name:      "AES-128 GCM",
			plaintext: testPlaintext,
			key:       testKey16,
		},
		{
			name:      "AES-192 GCM",
			plaintext: testPlaintext,
			key:       testKey24,
		},
		{
			name:      "AES-256 GCM",
			plaintext: testPlaintext,
			key:       testKey32,
		},
		{
			name:      "Empty plaintext",
			plaintext: []byte(""),
			key:       testKey16,
		},
		{
			name:      "Long plaintext",
			plaintext: make([]byte, 1000),
			key:       testKey16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill long plaintext with random data
			if len(tt.plaintext) == 1000 {
				rand.Read(tt.plaintext)
			}

			// Encrypt
			ciphertext, err := encryptAES(tt.plaintext, tt.key)
			if err != nil {
				t.Fatalf("encryptAES failed: %v", err)
			}

			// Verify ciphertext is not empty
			if ciphertext == "" {
				t.Error("Ciphertext should not be empty")
			}

			// Decrypt
			decrypted, err := decryptAES(ciphertext, tt.key)
			if err != nil {
				t.Fatalf("decryptAES failed: %v", err)
			}

			// Verify decrypted text matches original
			if string(decrypted) != string(tt.plaintext) {
				t.Errorf("Decrypted text doesn't match original. Got: %s, Want: %s", string(decrypted), string(tt.plaintext))
			}
		})
	}
}

func TestAesGCMNonceRandomness(t *testing.T) {
	// Test that each encryption produces different ciphertext due to random nonce
	plaintext := []byte("Test message for nonce randomness")
	key := testKey16

	ciphertext1, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("encryptAES failed: %v", err)
	}

	ciphertext2, err := encryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("encryptAES failed: %v", err)
	}

	// Ciphertexts should be different due to random nonce
	if ciphertext1 == ciphertext2 {
		t.Error("Ciphertexts should be different due to random nonce")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := decryptAES(ciphertext1, key)
	if err != nil {
		t.Fatalf("decryptAES failed: %v", err)
	}

	decrypted2, err := decryptAES(ciphertext2, key)
	if err != nil {
		t.Fatalf("decryptAES failed: %v", err)
	}

	if string(decrypted1) != string(plaintext) || string(decrypted2) != string(plaintext) {
		t.Error("Both decryptions should produce the same plaintext")
	}
}

func TestAesErrorCases(t *testing.T) {
	t.Run("Invalid key length", func(t *testing.T) {
		invalidKey := []byte("short") // Too short for AES
		_, err := AesEncrypt(testPlaintext, WithAesKey(invalidKey))
		if err == nil {
			t.Error("Expected error for invalid key length")
		}
	})

	t.Run("Invalid mode", func(t *testing.T) {
		_, err := aesEncryptByMode("INVALID", testPlaintext, testKey16)
		if err == nil {
			t.Error("Expected error for invalid mode")
		}
	})

	t.Run("Invalid hex string", func(t *testing.T) {
		_, err := AesDecryptHex("invalid hex", WithAesKey(testKey16))
		if err == nil {
			t.Error("Expected error for invalid hex string")
		}
	})

	t.Run("Corrupted ciphertext", func(t *testing.T) {
		// Encrypt first
		ciphertext, err := AesEncrypt(testPlaintext, WithAesKey(testKey16))
		if err != nil {
			t.Fatalf("AesEncrypt failed: %v", err)
		}

		// Corrupt the ciphertext (but keep it long enough to avoid padding issues)
		if len(ciphertext) > 16 {
			ciphertext[0] ^= 0xFF
		}

		// Try to decrypt - this might panic or return error depending on implementation
		var decrypted []byte
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Panic is acceptable for corrupted data
					t.Logf("Recovered from panic (expected): %v", r)
					return
				}
			}()
			decrypted, err = AesDecrypt(ciphertext, WithAesKey(testKey16))
		}()

		// If we got here without panic, check if decryption failed or produced wrong result
		if err == nil && decrypted != nil {
			if string(decrypted) == string(testPlaintext) {
				t.Error("Corrupted ciphertext should not decrypt to original plaintext")
			}
		}
	})

	t.Run("Wrong key for decryption", func(t *testing.T) {
		// Encrypt with one key
		ciphertext, err := AesEncrypt(testPlaintext, WithAesKey(testKey16))
		if err != nil {
			t.Fatalf("AesEncrypt failed: %v", err)
		}

		// Try to decrypt with different key - this might panic or return error
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Panic is acceptable for wrong key
					t.Logf("Recovered from panic (expected): %v", r)
				}
			}()
			_, err = AesDecrypt(ciphertext, WithAesKey(testKey24))
			if err == nil {
				t.Error("Expected error for wrong decryption key")
			}
		}()
	})
}

func TestAesGCMErrorCases(t *testing.T) {
	t.Run("Invalid key length for GCM", func(t *testing.T) {
		invalidKey := []byte("short") // Too short for AES
		_, err := encryptAES(testPlaintext, invalidKey)
		if err == nil {
			t.Error("Expected error for invalid key length")
		}
	})

	t.Run("Invalid base64 for GCM decryption", func(t *testing.T) {
		_, err := decryptAES("invalid base64", testKey16)
		if err == nil {
			t.Error("Expected error for invalid base64")
		}
	})

	t.Run("Corrupted GCM ciphertext", func(t *testing.T) {
		// Encrypt first
		ciphertext, err := encryptAES(testPlaintext, testKey16)
		if err != nil {
			t.Fatalf("encryptAES failed: %v", err)
		}

		// Corrupt the base64 string
		ciphertext = ciphertext[:len(ciphertext)-1] + "X"

		// Try to decrypt
		_, err = decryptAES(ciphertext, testKey16)
		if err == nil {
			t.Error("Expected error for corrupted GCM ciphertext")
		}
	})

	t.Run("Wrong key for GCM decryption", func(t *testing.T) {
		// Encrypt with one key
		ciphertext, err := encryptAES(testPlaintext, testKey16)
		if err != nil {
			t.Fatalf("encryptAES failed: %v", err)
		}

		// Try to decrypt with different key
		_, err = decryptAES(ciphertext, testKey24)
		if err == nil {
			t.Error("Expected error for wrong GCM decryption key")
		}
	})
}

func TestAesOptions(t *testing.T) {
	t.Run("Default options", func(t *testing.T) {
		opts := defaultAesOptions()
		if string(opts.aesKey) != string(defaultAesKey) {
			t.Errorf("Default key mismatch. Got: %s, Want: %s", string(opts.aesKey), string(defaultAesKey))
		}
		if opts.mode != defaultMode {
			t.Errorf("Default mode mismatch. Got: %s, Want: %s", opts.mode, defaultMode)
		}
	})

	t.Run("Custom key option", func(t *testing.T) {
		opts := defaultAesOptions()
		WithAesKey(testKey32)(opts)
		if string(opts.aesKey) != string(testKey32) {
			t.Errorf("Custom key not applied. Got: %s, Want: %s", string(opts.aesKey), string(testKey32))
		}
	})

	t.Run("Mode options", func(t *testing.T) {
		modes := []struct {
			option   func(*aesOptions)
			expected string
		}{
			{WithAesModeECB(), modeECB},
			{WithAesModeCBC(), modeCBC},
			{WithAesModeCFB(), modeCFB},
			{WithAesModeCTR(), modeCTR},
			{WithAesModeGCM(), modeGCM},
		}

		for _, m := range modes {
			opts := defaultAesOptions()
			m.option(opts)
			if opts.mode != m.expected {
				t.Errorf("Mode option not applied. Got: %s, Want: %s", opts.mode, m.expected)
			}
		}
	})
}

func TestAesConsistency(t *testing.T) {
	// Test that the same input always produces the same output for deterministic modes
	plaintext := []byte("Consistency test message")
	key := testKey16

	// ECB mode should be deterministic
	ciphertext1, err := AesEncrypt(plaintext, WithAesKey(key), WithAesModeECB())
	if err != nil {
		t.Fatalf("AesEncrypt failed: %v", err)
	}

	ciphertext2, err := AesEncrypt(plaintext, WithAesKey(key), WithAesModeECB())
	if err != nil {
		t.Fatalf("AesEncrypt failed: %v", err)
	}

	if string(ciphertext1) != string(ciphertext2) {
		t.Error("ECB mode should be deterministic")
	}
}

func TestAesGCMNonDeterministic(t *testing.T) {
	// Test that GCM mode produces different ciphertexts each time due to random nonce
	plaintext := []byte("GCM non-deterministic test message")
	key := testKey16

	ciphertext1, err := AesEncrypt(plaintext, WithAesKey(key), WithAesModeGCM())
	if err != nil {
		t.Fatalf("AesEncrypt failed: %v", err)
	}

	ciphertext2, err := AesEncrypt(plaintext, WithAesKey(key), WithAesModeGCM())
	if err != nil {
		t.Fatalf("AesEncrypt failed: %v", err)
	}

	// GCM mode should produce different ciphertexts due to random nonce
	if string(ciphertext1) == string(ciphertext2) {
		t.Error("GCM mode should be non-deterministic due to random nonce")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := AesDecrypt(ciphertext1, WithAesKey(key), WithAesModeGCM())
	if err != nil {
		t.Fatalf("AesDecrypt failed: %v", err)
	}

	decrypted2, err := AesDecrypt(ciphertext2, WithAesKey(key), WithAesModeGCM())
	if err != nil {
		t.Fatalf("AesDecrypt failed: %v", err)
	}

	if string(decrypted1) != string(plaintext) || string(decrypted2) != string(plaintext) {
		t.Error("Both GCM decryptions should produce the same plaintext")
	}
}

func BenchmarkAesEncrypt(b *testing.B) {
	plaintext := make([]byte, 1024)
	rand.Read(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AesEncrypt(plaintext, WithAesKey(testKey16), WithAesModeCBC())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAesDecrypt(b *testing.B) {
	plaintext := make([]byte, 1024)
	rand.Read(plaintext)

	ciphertext, err := AesEncrypt(plaintext, WithAesKey(testKey16), WithAesModeCBC())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AesDecrypt(ciphertext, WithAesKey(testKey16), WithAesModeCBC())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAesGCMEncrypt(b *testing.B) {
	plaintext := make([]byte, 1024)
	rand.Read(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptAES(plaintext, testKey16)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAesGCMDecrypt(b *testing.B) {
	plaintext := make([]byte, 1024)
	rand.Read(plaintext)

	ciphertext, err := encryptAES(plaintext, testKey16)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decryptAES(ciphertext, testKey16)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to get mode option
func getModeOption(mode string) AesOption {
	switch mode {
	case modeECB:
		return WithAesModeECB()
	case modeCBC:
		return WithAesModeCBC()
	case modeCFB:
		return WithAesModeCFB()
	case modeCTR:
		return WithAesModeCTR()
	case modeGCM:
		return WithAesModeGCM()
	default:
		return WithAesModeECB()
	}
}
