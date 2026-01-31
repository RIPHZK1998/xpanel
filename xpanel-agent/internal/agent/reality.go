package agent

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/curve25519"
)

// RealityKeys holds the generated x25519 keypair for Reality protocol.
type RealityKeys struct {
	PrivateKey string // Base64-encoded private key
	PublicKey  string // Base64-encoded public key
}

// GenerateRealityKeypair generates a new x25519 keypair for Reality protocol.
// Returns the Base64-encoded private and public keys.
func GenerateRealityKeypair() (*RealityKeys, error) {
	// Generate a random 32-byte private key
	var privateKey [32]byte
	if _, err := rand.Read(privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Clamp the private key for x25519 (standard curve25519 clamping)
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	// Compute the public key
	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &RealityKeys{
		PrivateKey: base64.RawURLEncoding.EncodeToString(privateKey[:]),
		PublicKey:  base64.RawURLEncoding.EncodeToString(publicKey[:]),
	}, nil
}

// LoadOrCreateRealityKeys loads existing keys from a file or creates new ones.
// The private key is stored in a file at the specified path.
// Returns the keypair (always includes both keys).
func LoadOrCreateRealityKeys(keyPath string) (*RealityKeys, bool, error) {
	// Ensure directory exists
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, false, fmt.Errorf("failed to create key directory: %w", err)
	}

	// Try to load existing key
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err == nil {
		// Key exists, decode and derive public key
		privateKeyStr := string(privateKeyBytes)
		privateKey, err := base64.RawURLEncoding.DecodeString(privateKeyStr)
		if err != nil || len(privateKey) != 32 {
			// Invalid key format, regenerate
			return generateAndSaveKeys(keyPath)
		}

		// Derive public key from private key
		var privKeyArray [32]byte
		copy(privKeyArray[:], privateKey)
		var publicKey [32]byte
		curve25519.ScalarBaseMult(&publicKey, &privKeyArray)

		return &RealityKeys{
			PrivateKey: privateKeyStr,
			PublicKey:  base64.RawURLEncoding.EncodeToString(publicKey[:]),
		}, false, nil
	}

	// Key doesn't exist, generate new one
	return generateAndSaveKeys(keyPath)
}

// generateAndSaveKeys generates a new keypair and saves the private key.
func generateAndSaveKeys(keyPath string) (*RealityKeys, bool, error) {
	keys, err := GenerateRealityKeypair()
	if err != nil {
		return nil, true, fmt.Errorf("failed to generate Reality keypair: %w", err)
	}

	// Save private key to file with restrictive permissions
	if err := os.WriteFile(keyPath, []byte(keys.PrivateKey), 0600); err != nil {
		return nil, true, fmt.Errorf("failed to save private key: %w", err)
	}

	return keys, true, nil
}
