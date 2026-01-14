package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSessionToken generates a cryptographically secure random session token
// SECURITY: Uses crypto/rand (not math/rand) to prevent predictability attacks
// Returns 64-character hex string (32 bytes = 256 bits of entropy)
func GenerateSessionToken() string {
	b := make([]byte, 32) // 256 bits of entropy
	_, err := rand.Read(b)
	if err != nil {
		// CRITICAL: If random generation fails, panic instead of returning empty string
		// Returning empty string would cause silent failures and potential security issues
		panic("FATAL: failed to generate secure random session token: " + err.Error())
	}
	return hex.EncodeToString(b)
}
