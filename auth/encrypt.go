package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"strings"
	"sync"

	"testing_go/models"
)

var (
	encryptionKey  []byte
	encKeyOnce     sync.Once
)

func getEncryptionKey() []byte {
	encKeyOnce.Do(func() {
		key := os.Getenv("ENCRYPTION_KEY")
		if key == "" {
			key = os.Getenv("JWT_SECRET")
		}
		if key == "" {
			key = "tierlog-dev-secret"
		}
		hash := hashSHA256([]byte(key))
		encryptionKey = hash
	})
	return encryptionKey
}

func hashSHA256(data []byte) []byte {
	h := [32]byte{}
	copy(h[:], data)
	return h[:]
}

func EncryptAPIKey(plaintext string) string {
	if plaintext == "" {
		return ""
	}
	if strings.HasPrefix(plaintext, "enc:v1:") {
		return plaintext
	}

	block, err := aes.NewCipher(getEncryptionKey())
	if err != nil {
		return plaintext
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plaintext
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return "enc:v1:" + base64.RawURLEncoding.EncodeToString(ciphertext)
}

func DecryptAPIKey(ciphertext string) string {
	if ciphertext == "" {
		return ""
	}
	if !strings.HasPrefix(ciphertext, "enc:v1:") {
		return ciphertext
	}

	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(ciphertext, "enc:v1:"))
	if err != nil {
		return ""
	}

	block, err := aes.NewCipher(getEncryptionKey())
	if err != nil {
		return ""
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	nonceSize := aesGCM.NonceSize()
	if len(raw) < nonceSize {
		return ""
	}

	nonce, ciphertextBytes := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return ""
	}

	return string(plaintext)
}

func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func IsKeySet(key string) bool {
	return strings.TrimSpace(key) != ""
}

// ResolveKeyForProvider decrypts the correct API key for a given provider string.
// Returns empty string if provider is unknown or user has no key for that provider.
func ResolveKeyForProvider(user *models.User, provider string) string {
	switch strings.ToLower(provider) {
	case "nvidia":
		return DecryptAPIKey(user.NvidiaKey)
	case "openai":
		return DecryptAPIKey(user.OpenAIKey)
	case "anthropic":
		return DecryptAPIKey(user.AnthropicKey)
	case "gemini":
		return DecryptAPIKey(user.GeminiKey)
	case "groq":
		return DecryptAPIKey(user.GroqKey)
	default:
		return ""
	}
}
