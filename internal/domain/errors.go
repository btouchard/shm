// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import "errors"

// Sentinel errors for the domain layer.
// Use errors.Is() to check for these errors.
// Wrap with fmt.Errorf("context: %w", ErrXxx) to add context.

var (
	// Instance errors
	ErrInstanceNotFound        = errors.New("instance not found")
	ErrInstanceRevoked         = errors.New("instance is revoked")
	ErrInvalidInstanceID       = errors.New("invalid instance ID")
	ErrInvalidPublicKey        = errors.New("invalid public key")
	ErrInvalidInstance         = errors.New("invalid instance")
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// Snapshot errors
	ErrInvalidSnapshot = errors.New("invalid snapshot")
	ErrInvalidMetrics  = errors.New("invalid metrics")

	// Authentication errors
	ErrInvalidSignature = errors.New("invalid signature")
	ErrMissingSignature = errors.New("missing signature")
)
