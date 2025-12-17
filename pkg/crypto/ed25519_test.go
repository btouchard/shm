// SPDX-License-Identifier: AGPL-3.0-or-later

package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"strings"
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY TESTS
// =============================================================================

func TestGenerateKeypair(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair() error = %v", err)
	}

	if len(pub) != ed25519.PublicKeySize {
		t.Errorf("public key size = %d, want %d", len(pub), ed25519.PublicKeySize)
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Errorf("private key size = %d, want %d", len(priv), ed25519.PrivateKeySize)
	}
}

func TestGenerateKeypair_Uniqueness(t *testing.T) {
	pub1, _, _ := GenerateKeypair()
	pub2, _, _ := GenerateKeypair()

	if hex.EncodeToString(pub1) == hex.EncodeToString(pub2) {
		t.Error("two generated keypairs should not be identical")
	}
}

func TestSignAndVerify_RoundTrip(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)
	message := []byte("test message for signing")

	signature := Sign(priv, message)

	if signature == "" {
		t.Fatal("Sign() returned empty signature")
	}

	if !Verify(pubHex, message, signature) {
		t.Error("Verify() should return true for valid signature")
	}
}

func TestSign_DeterministicForSameInput(t *testing.T) {
	_, priv, _ := GenerateKeypair()
	message := []byte("same message")

	sig1 := Sign(priv, message)
	sig2 := Sign(priv, message)

	// Ed25519 is deterministic - same key + message = same signature
	if sig1 != sig2 {
		t.Error("Ed25519 signatures should be deterministic")
	}
}

func TestSign_DifferentMessagesProduceDifferentSignatures(t *testing.T) {
	_, priv, _ := GenerateKeypair()

	sig1 := Sign(priv, []byte("message one"))
	sig2 := Sign(priv, []byte("message two"))

	if sig1 == sig2 {
		t.Error("different messages should produce different signatures")
	}
}

// =============================================================================
// SECURITY TESTS - Signature Verification
// =============================================================================

func TestVerify_WrongPublicKey(t *testing.T) {
	pub1, priv1, _ := GenerateKeypair()
	pub2, _, _ := GenerateKeypair()

	message := []byte("secret message")
	signature := Sign(priv1, message)

	// Verify with correct key should pass
	if !Verify(hex.EncodeToString(pub1), message, signature) {
		t.Error("verification with correct key should pass")
	}

	// Verify with wrong key should fail
	if Verify(hex.EncodeToString(pub2), message, signature) {
		t.Error("verification with wrong public key should fail")
	}
}

func TestVerify_TamperedMessage(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)

	original := []byte("original message")
	signature := Sign(priv, original)

	// Original message verifies
	if !Verify(pubHex, original, signature) {
		t.Fatal("original message should verify")
	}

	// Tampered message should NOT verify
	tampered := []byte("tampered message")
	if Verify(pubHex, tampered, signature) {
		t.Error("tampered message should NOT verify with original signature")
	}

	// Even slight modification should fail
	slightlyModified := []byte("original messagE") // capital E
	if Verify(pubHex, slightlyModified, signature) {
		t.Error("even slightly modified message should NOT verify")
	}
}

func TestVerify_TamperedSignature(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)
	message := []byte("message")

	signature := Sign(priv, message)

	// Tamper with signature (flip one character)
	sigBytes, _ := hex.DecodeString(signature)
	sigBytes[0] ^= 0xFF // flip bits
	tamperedSig := hex.EncodeToString(sigBytes)

	if Verify(pubHex, message, tamperedSig) {
		t.Error("tampered signature should NOT verify")
	}
}

func TestVerify_SignatureNotReplayable(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)

	message1 := []byte(`{"action": "activate", "nonce": "abc123"}`)
	message2 := []byte(`{"action": "activate", "nonce": "def456"}`)

	sig1 := Sign(priv, message1)

	// Signature for message1 should NOT work for message2
	if Verify(pubHex, message2, sig1) {
		t.Error("signature should not be replayable on different message")
	}
}

// =============================================================================
// SECURITY TESTS - Malformed Input Handling
// =============================================================================

func TestVerify_MalformedPublicKeyHex(t *testing.T) {
	_, priv, _ := GenerateKeypair()
	message := []byte("test")
	signature := Sign(priv, message)

	tests := []struct {
		name   string
		pubKey string
	}{
		{"empty string", ""},
		{"not hex", "not-valid-hex!@#$"},
		{"odd length hex", "abc"},
		{"too short", "abcd1234"},
		{"too long", strings.Repeat("ab", 64)}, // 128 chars = 64 bytes, should be 32
		{"null bytes in hex", "00000000000000000000000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should return false, NOT panic
			result := Verify(tt.pubKey, message, signature)
			if result {
				t.Errorf("Verify() with malformed pubKey %q should return false", tt.name)
			}
		})
	}
}

func TestVerify_MalformedSignatureHex(t *testing.T) {
	pub, _, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)
	message := []byte("test")

	tests := []struct {
		name      string
		signature string
	}{
		{"empty string", ""},
		{"not hex", "not-valid-hex!@#$"},
		{"odd length hex", "abc"},
		{"too short", "abcd1234"},
		{"too long", strings.Repeat("ab", 128)}, // 256 chars = 128 bytes, should be 64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should return false, NOT panic
			result := Verify(pubHex, message, tt.signature)
			if result {
				t.Errorf("Verify() with malformed signature %q should return false", tt.name)
			}
		})
	}
}

func TestVerify_EmptyMessage(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)

	// Empty message is valid for Ed25519
	emptyMessage := []byte{}
	signature := Sign(priv, emptyMessage)

	if !Verify(pubHex, emptyMessage, signature) {
		t.Error("empty message should be signable and verifiable")
	}

	// But signature for empty should not verify non-empty
	if Verify(pubHex, []byte("not empty"), signature) {
		t.Error("signature for empty message should not verify non-empty message")
	}
}

func TestVerify_NilMessage(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)

	// nil is treated as empty slice in Go
	signature := Sign(priv, nil)

	if !Verify(pubHex, nil, signature) {
		t.Error("nil message should be signable and verifiable")
	}
}

func TestVerify_LargeMessage(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	pubHex := hex.EncodeToString(pub)

	// 1MB message
	largeMessage := make([]byte, 1024*1024)
	for i := range largeMessage {
		largeMessage[i] = byte(i % 256)
	}

	signature := Sign(priv, largeMessage)

	if !Verify(pubHex, largeMessage, signature) {
		t.Error("large message should be signable and verifiable")
	}
}

// =============================================================================
// EDGE CASES
// =============================================================================

func TestVerify_CaseSensitiveHex(t *testing.T) {
	pub, priv, _ := GenerateKeypair()
	message := []byte("test")
	signature := Sign(priv, message)

	pubHexLower := strings.ToLower(hex.EncodeToString(pub))
	pubHexUpper := strings.ToUpper(hex.EncodeToString(pub))

	// Both should work (hex.DecodeString is case-insensitive)
	if !Verify(pubHexLower, message, signature) {
		t.Error("lowercase hex pubkey should work")
	}
	if !Verify(pubHexUpper, message, signature) {
		t.Error("uppercase hex pubkey should work")
	}
}

func TestSign_OutputIsValidHex(t *testing.T) {
	_, priv, _ := GenerateKeypair()
	signature := Sign(priv, []byte("test"))

	// Should be valid hex
	decoded, err := hex.DecodeString(signature)
	if err != nil {
		t.Errorf("Sign() output is not valid hex: %v", err)
	}

	// Ed25519 signatures are 64 bytes
	if len(decoded) != ed25519.SignatureSize {
		t.Errorf("signature size = %d bytes, want %d", len(decoded), ed25519.SignatureSize)
	}
}
