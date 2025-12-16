// SPDX-License-Identifier: AGPL-3.0-or-later

package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
)

// GenerateKeypair crée une nouvelle paire de clés
func GenerateKeypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// Sign signe un payload
func Sign(privateKey ed25519.PrivateKey, message []byte) string {
	sig := ed25519.Sign(privateKey, message)
	return hex.EncodeToString(sig)
}

// Verify vérifie une signature (utilisé par le serveur)
func Verify(pubKeyHex string, message []byte, signatureHex string) bool {
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil || len(pubKey) != ed25519.PublicKeySize {
		return false
	}
	sig, err := hex.DecodeString(signatureHex)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(pubKey, message, sig)
}
