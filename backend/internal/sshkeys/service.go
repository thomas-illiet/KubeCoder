// Package sshkeys generates and protects organization SSH identities.
package sshkeys

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	keyAlgorithm      = "ssh-ed25519"
	encryptionVersion = "aes-256-gcm-v1"
)

var ErrNotFound = errors.New("organization SSH key not found")

// Service generates, encrypts, stores, and retrieves organization SSH keys.
type Service struct {
	db   *gorm.DB
	aead cipher.AEAD
}

// NewService validates the 256-bit master key and creates an SSH key service.
func NewService(db *gorm.DB, masterKey []byte) (*Service, error) {
	if len(masterKey) != 32 {
		return nil, errors.New("organization SSH key master key must contain exactly 32 bytes")
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("create organization SSH key cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create organization SSH key AEAD: %w", err)
	}
	return &Service{db: db, aead: aead}, nil
}

// Generate creates a new Ed25519 identity encrypted for the organization.
func (s *Service) Generate(organizationID uuid.UUID) (models.OrganizationSSHKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return models.OrganizationSSHKey{}, fmt.Errorf("generate Ed25519 key: %w", err)
	}
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return models.OrganizationSSHKey{}, fmt.Errorf("encode SSH public key: %w", err)
	}
	privateBlock, err := ssh.MarshalPrivateKey(privateKey, "kubecoder organization "+organizationID.String())
	if err != nil {
		return models.OrganizationSSHKey{}, fmt.Errorf("encode SSH private key: %w", err)
	}
	privatePEM := pem.EncodeToMemory(privateBlock)
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return models.OrganizationSSHKey{}, fmt.Errorf("generate SSH key nonce: %w", err)
	}
	return models.OrganizationSSHKey{
		OrganizationID:      organizationID,
		PublicKey:           string(ssh.MarshalAuthorizedKey(sshPublicKey)),
		EncryptedPrivateKey: s.aead.Seal(nil, nonce, privatePEM, additionalData(organizationID)),
		Nonce:               nonce,
		KeyAlgorithm:        keyAlgorithm,
		EncryptionVersion:   encryptionVersion,
		Fingerprint:         ssh.FingerprintSHA256(sshPublicKey),
	}, nil
}

// Get returns the public metadata for an organization SSH key.
func (s *Service) Get(ctx context.Context, organizationID uuid.UUID) (models.OrganizationSSHKey, error) {
	var result models.OrganizationSSHKey
	err := s.db.WithContext(ctx).First(&result, "organization_id = ?", organizationID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.OrganizationSSHKey{}, ErrNotFound
	}
	return result, err
}

// Regenerate atomically creates a missing identity or replaces an existing one.
func (s *Service) Regenerate(ctx context.Context, organizationID uuid.UUID) (models.OrganizationSSHKey, error) {
	key, err := s.Generate(organizationID)
	if err != nil {
		return models.OrganizationSSHKey{}, err
	}
	err = s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"public_key", "encrypted_private_key", "nonce", "key_algorithm", "encryption_version", "fingerprint", "updated_at",
		}),
	}).Create(&key).Error
	if err != nil {
		return models.OrganizationSSHKey{}, fmt.Errorf("persist organization SSH key: %w", err)
	}
	return key, nil
}

// DecryptPrivateKey returns the protected private key for future internal clone workers.
func (s *Service) DecryptPrivateKey(key models.OrganizationSSHKey) ([]byte, error) {
	if key.EncryptionVersion != encryptionVersion {
		return nil, fmt.Errorf("unsupported SSH key encryption version %q", key.EncryptionVersion)
	}
	plaintext, err := s.aead.Open(nil, key.Nonce, key.EncryptedPrivateKey, additionalData(key.OrganizationID))
	if err != nil {
		return nil, errors.New("decrypt organization SSH key")
	}
	return plaintext, nil
}

// additionalData cryptographically binds ciphertext to its organization and purpose.
func additionalData(organizationID uuid.UUID) []byte {
	return []byte("kubecoder:organization-ssh-key:" + organizationID.String())
}
