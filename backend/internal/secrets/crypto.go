package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const encryptionVersion = "aes-256-gcm-v1"

type protector struct {
	aead cipher.AEAD
	key  []byte
}

// newProtector validates the master key and initializes authenticated encryption.
func newProtector(key []byte) (*protector, error) {
	if len(key) != 32 {
		return nil, errors.New("secret encryption key must contain exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create secret cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create secret AEAD: %w", err)
	}
	return &protector{aead: aead, key: append([]byte(nil), key...)}, nil
}

// protect encrypts one value and returns storage-only metadata.
func (p *protector) protect(id uuid.UUID, scope, value string) ([]byte, []byte, string, error) {
	nonce := make([]byte, p.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, "", fmt.Errorf("generate secret nonce: %w", err)
	}
	mac := hmac.New(sha256.New, p.key)
	_, _ = mac.Write([]byte("kubecoder:secret:fingerprint:" + value))
	fingerprint := hex.EncodeToString(mac.Sum(nil)[:12])
	ciphertext := p.aead.Seal(nil, nonce, []byte(value), []byte("kubecoder:secret:"+scope+":"+id.String()))
	return ciphertext, nonce, fingerprint, nil
}

// open decrypts one value for internal resolution and tests only.
func (p *protector) open(secret Secret) ([]byte, error) {
	if secret.EncryptionVersion != encryptionVersion {
		return nil, errors.New("unsupported secret encryption version")
	}
	value, err := p.aead.Open(nil, secret.Nonce, secret.EncryptedValue, []byte("kubecoder:secret:"+secret.Scope+":"+secret.ID.String()))
	if err != nil {
		return nil, errors.New("decrypt secret")
	}
	return value, nil
}
