// SPDX-License-Identifier: MIT

import {
  generateKeyPairSync,
  sign,
  verify,
  createPrivateKey,
  createPublicKey,
} from 'node:crypto';

/**
 * Generates an Ed25519 keypair.
 * @returns Object containing hex-encoded public and private keys.
 */
export function generateKeypair(): { publicKey: string; privateKey: string } {
  const { publicKey, privateKey } = generateKeyPairSync('ed25519');

  // Export raw key bytes and convert to hex
  const pubKeyBuffer = publicKey.export({ type: 'spki', format: 'der' });
  const privKeyBuffer = privateKey.export({ type: 'pkcs8', format: 'der' });

  // Ed25519 SPKI format: 12-byte header + 32-byte key
  const pubKeyRaw = pubKeyBuffer.subarray(12);
  // Ed25519 PKCS8 format: 16-byte header + 32-byte key
  const privKeyRaw = privKeyBuffer.subarray(16);

  // The full private key for signing is seed (32 bytes) + public key (32 bytes)
  const fullPrivateKey = Buffer.concat([privKeyRaw, pubKeyRaw]);

  return {
    publicKey: pubKeyRaw.toString('hex'),
    privateKey: fullPrivateKey.toString('hex'),
  };
}

/**
 * Signs a message using Ed25519 private key.
 * @param privateKeyHex - Hex-encoded private key (64 bytes = seed + public key).
 * @param message - Message to sign as Buffer or string.
 * @returns Hex-encoded signature.
 */
export function signMessage(privateKeyHex: string, message: Buffer | string): string {
  const privKeyBytes = Buffer.from(privateKeyHex, 'hex');

  // Node.js expects PKCS8 format for Ed25519 private keys
  // We need to wrap our raw key in the PKCS8 structure
  const pkcs8Header = Buffer.from([
    0x30, 0x2e, // SEQUENCE, length 46
    0x02, 0x01, 0x00, // INTEGER, version 0
    0x30, 0x05, // SEQUENCE, length 5
    0x06, 0x03, 0x2b, 0x65, 0x70, // OID 1.3.101.112 (Ed25519)
    0x04, 0x22, // OCTET STRING, length 34
    0x04, 0x20, // OCTET STRING, length 32
  ]);

  // Use only the seed portion (first 32 bytes) for PKCS8 format
  const seed = privKeyBytes.subarray(0, 32);
  const pkcs8Key = Buffer.concat([pkcs8Header, seed]);

  const keyObject = createPrivateKey({
    key: pkcs8Key,
    format: 'der',
    type: 'pkcs8',
  });

  const messageBuffer = typeof message === 'string' ? Buffer.from(message) : message;
  const signature = sign(null, messageBuffer, keyObject);

  return signature.toString('hex');
}

/**
 * Verifies an Ed25519 signature.
 * @param publicKeyHex - Hex-encoded public key (32 bytes).
 * @param message - Original message as Buffer or string.
 * @param signatureHex - Hex-encoded signature.
 * @returns True if signature is valid.
 */
export function verifySignature(
  publicKeyHex: string,
  message: Buffer | string,
  signatureHex: string
): boolean {
  try {
    const pubKeyBytes = Buffer.from(publicKeyHex, 'hex');
    if (pubKeyBytes.length !== 32) {
      return false;
    }

    const sigBytes = Buffer.from(signatureHex, 'hex');
    if (sigBytes.length !== 64) {
      return false;
    }

    // Wrap raw public key in SPKI format
    const spkiHeader = Buffer.from([
      0x30, 0x2a, // SEQUENCE, length 42
      0x30, 0x05, // SEQUENCE, length 5
      0x06, 0x03, 0x2b, 0x65, 0x70, // OID 1.3.101.112 (Ed25519)
      0x03, 0x21, 0x00, // BIT STRING, length 33, no unused bits
    ]);
    const spkiKey = Buffer.concat([spkiHeader, pubKeyBytes]);

    const keyObject = createPublicKey({
      key: spkiKey,
      format: 'der',
      type: 'spki',
    });

    const messageBuffer = typeof message === 'string' ? Buffer.from(message) : message;

    return verify(null, messageBuffer, keyObject, sigBytes);
  } catch {
    return false;
  }
}
