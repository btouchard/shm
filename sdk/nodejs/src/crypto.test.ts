// SPDX-License-Identifier: MIT

import { describe, it } from 'node:test';
import assert from 'node:assert';
import { generateKeypair, signMessage, verifySignature } from './crypto.js';

describe('crypto', () => {
  describe('generateKeypair', () => {
    it('should generate valid keypair', () => {
      const { publicKey, privateKey } = generateKeypair();

      assert.strictEqual(publicKey.length, 64, 'public key should be 32 bytes (64 hex chars)');
      assert.strictEqual(privateKey.length, 128, 'private key should be 64 bytes (128 hex chars)');

      // Verify both are valid hex
      assert.ok(/^[0-9a-f]+$/.test(publicKey), 'public key should be hex');
      assert.ok(/^[0-9a-f]+$/.test(privateKey), 'private key should be hex');
    });

    it('should generate unique keypairs', () => {
      const kp1 = generateKeypair();
      const kp2 = generateKeypair();

      assert.notStrictEqual(kp1.publicKey, kp2.publicKey, 'public keys should be unique');
      assert.notStrictEqual(kp1.privateKey, kp2.privateKey, 'private keys should be unique');
    });
  });

  describe('signMessage', () => {
    it('should produce 64-byte signature', () => {
      const { privateKey } = generateKeypair();
      const message = 'test message';

      const signature = signMessage(privateKey, message);

      assert.strictEqual(signature.length, 128, 'signature should be 64 bytes (128 hex chars)');
      assert.ok(/^[0-9a-f]+$/.test(signature), 'signature should be hex');
    });

    it('should produce different signatures for different messages', () => {
      const { privateKey } = generateKeypair();

      const sig1 = signMessage(privateKey, 'message 1');
      const sig2 = signMessage(privateKey, 'message 2');

      assert.notStrictEqual(sig1, sig2, 'signatures should differ for different messages');
    });

    it('should produce same signature for same message', () => {
      const { privateKey } = generateKeypair();
      const message = 'same message';

      const sig1 = signMessage(privateKey, message);
      const sig2 = signMessage(privateKey, message);

      assert.strictEqual(sig1, sig2, 'signatures should be identical for same message');
    });
  });

  describe('verifySignature', () => {
    it('should verify valid signature', () => {
      const { publicKey, privateKey } = generateKeypair();
      const message = 'test message';
      const signature = signMessage(privateKey, message);

      const isValid = verifySignature(publicKey, message, signature);

      assert.strictEqual(isValid, true, 'valid signature should verify');
    });

    it('should reject tampered message', () => {
      const { publicKey, privateKey } = generateKeypair();
      const message = 'original message';
      const signature = signMessage(privateKey, message);

      const isValid = verifySignature(publicKey, 'tampered message', signature);

      assert.strictEqual(isValid, false, 'tampered message should not verify');
    });

    it('should reject wrong public key', () => {
      const kp1 = generateKeypair();
      const kp2 = generateKeypair();
      const message = 'test message';
      const signature = signMessage(kp1.privateKey, message);

      const isValid = verifySignature(kp2.publicKey, message, signature);

      assert.strictEqual(isValid, false, 'wrong public key should not verify');
    });

    it('should reject invalid public key length', () => {
      const { privateKey } = generateKeypair();
      const message = 'test message';
      const signature = signMessage(privateKey, message);

      const isValid = verifySignature('abcd', message, signature);

      assert.strictEqual(isValid, false, 'invalid public key should not verify');
    });

    it('should reject invalid signature length', () => {
      const { publicKey } = generateKeypair();
      const message = 'test message';

      const isValid = verifySignature(publicKey, message, 'abcd');

      assert.strictEqual(isValid, false, 'invalid signature should not verify');
    });
  });

  describe('Buffer message support', () => {
    it('should sign Buffer messages', () => {
      const { publicKey, privateKey } = generateKeypair();
      const message = Buffer.from('test message');
      const signature = signMessage(privateKey, message);

      const isValid = verifySignature(publicKey, message, signature);

      assert.strictEqual(isValid, true, 'Buffer message should work');
    });

    it('should produce same signature for string and Buffer', () => {
      const { privateKey } = generateKeypair();
      const strMessage = 'test message';
      const bufMessage = Buffer.from(strMessage);

      const sig1 = signMessage(privateKey, strMessage);
      const sig2 = signMessage(privateKey, bufMessage);

      assert.strictEqual(sig1, sig2, 'string and Buffer should produce same signature');
    });
  });
});
