package sshkeys

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/thomas-illiet/KubeCoder/backend/internal/models"
	"golang.org/x/crypto/ssh"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestGenerateEncryptsValidEd25519PrivateKey verifies both OpenSSH formats and encryption.
func TestGenerateEncryptsValidEd25519PrivateKey(t *testing.T) {
	t.Parallel()
	service, err := NewService(nil, bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	organizationID := uuid.New()
	key, err := service.Generate(organizationID)
	if err != nil {
		t.Fatal(err)
	}
	if key.KeyAlgorithm != "ssh-ed25519" || key.EncryptionVersion != encryptionVersion {
		t.Fatalf("unexpected algorithms: %#v", key)
	}
	if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key.PublicKey)); err != nil {
		t.Fatalf("parse public key: %v", err)
	}
	privatePEM, err := service.DecryptPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ssh.ParseRawPrivateKey(privatePEM); err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	if bytes.Contains(key.EncryptedPrivateKey, privatePEM) {
		t.Fatal("encrypted value contains plaintext private key")
	}
}

// TestDecryptRejectsWrongMasterKey verifies authenticated encryption rejects another master key.
func TestDecryptRejectsWrongMasterKey(t *testing.T) {
	t.Parallel()
	owner, _ := NewService(nil, bytes.Repeat([]byte{1}, 32))
	other, _ := NewService(nil, bytes.Repeat([]byte{2}, 32))
	key, err := owner.Generate(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.DecryptPrivateKey(key); err == nil {
		t.Fatal("expected decryption with another master key to fail")
	}
}

// TestOrganizationSSHKeyJSONExcludesPrivateMaterial verifies the model's public wire contract.
func TestOrganizationSSHKeyJSONExcludesPrivateMaterial(t *testing.T) {
	t.Parallel()
	payload, err := json.Marshal(models.OrganizationSSHKey{
		PublicKey: "ssh-ed25519 public", EncryptedPrivateKey: []byte("ciphertext"), Nonce: []byte("nonce"),
		EncryptionVersion: encryptionVersion, KeyAlgorithm: keyAlgorithm, Fingerprint: "SHA256:test",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range [][]byte{[]byte("ciphertext"), []byte("nonce"), []byte("encryption"), []byte("organization_id")} {
		if bytes.Contains(payload, forbidden) {
			t.Fatalf("response leaks protected material %q: %s", forbidden, payload)
		}
	}
}

// TestNewServiceRequiresAES256Key verifies weaker AES key sizes are rejected.
func TestNewServiceRequiresAES256Key(t *testing.T) {
	t.Parallel()
	if _, err := NewService(nil, make([]byte, 16)); err == nil {
		t.Fatal("expected non-256-bit key to fail")
	}
}

// TestRegenerateCreatesMissingKey verifies existing organizations can bootstrap through rotation.
func TestRegenerateCreatesMissingKey(t *testing.T) {
	t.Parallel()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE organization_ssh_keys (organization_id uuid PRIMARY KEY, public_key text NOT NULL, encrypted_private_key blob NOT NULL, nonce blob NOT NULL, key_algorithm text NOT NULL, encryption_version text NOT NULL, fingerprint text NOT NULL, created_at datetime NOT NULL, updated_at datetime NOT NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	service, err := NewService(db, bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	organizationID := uuid.New()
	key, err := service.Regenerate(t.Context(), organizationID)
	if err != nil {
		t.Fatal(err)
	}
	if key.OrganizationID != organizationID {
		t.Fatalf("organization ID = %s", key.OrganizationID)
	}
}
