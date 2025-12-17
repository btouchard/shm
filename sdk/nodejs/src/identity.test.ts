// SPDX-License-Identifier: MIT

import { describe, it, beforeEach, afterEach } from 'node:test';
import assert from 'node:assert';
import { mkdtempSync, rmSync, writeFileSync, existsSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { loadOrGenerateIdentity, slug } from './identity.js';
import { signMessage, verifySignature } from './crypto.js';

describe('identity', () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), 'shm-test-'));
  });

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true });
  });

  describe('loadOrGenerateIdentity', () => {
    it('should generate new identity when file does not exist', () => {
      const idPath = join(tmpDir, 'test_identity.json');

      const identity = loadOrGenerateIdentity(idPath);

      assert.ok(identity.instanceId, 'instanceId should exist');
      assert.ok(identity.publicKey, 'publicKey should exist');
      assert.ok(identity.privateKey, 'privateKey should exist');
      assert.strictEqual(identity.publicKey.length, 64, 'publicKey should be 32 bytes hex');
      assert.strictEqual(identity.privateKey.length, 128, 'privateKey should be 64 bytes hex');
      assert.ok(existsSync(idPath), 'identity file should be created');
    });

    it('should load existing identity from file', () => {
      const idPath = join(tmpDir, 'test_identity.json');

      const id1 = loadOrGenerateIdentity(idPath);
      const id2 = loadOrGenerateIdentity(idPath);

      assert.strictEqual(id1.instanceId, id2.instanceId, 'instanceId should be same');
      assert.strictEqual(id1.publicKey, id2.publicKey, 'publicKey should be same');
      assert.strictEqual(id1.privateKey, id2.privateKey, 'privateKey should be same');
    });

    it('should regenerate identity if file is corrupted', () => {
      const idPath = join(tmpDir, 'corrupted_identity.json');
      writeFileSync(idPath, 'not valid json {{{');

      const identity = loadOrGenerateIdentity(idPath);

      assert.ok(identity.instanceId, 'should generate new identity');
      assert.ok(identity.publicKey, 'publicKey should exist');
    });

    it('should create nested directories', () => {
      const nestedPath = join(tmpDir, 'a', 'b', 'c', 'identity.json');

      const identity = loadOrGenerateIdentity(nestedPath);

      assert.ok(identity.instanceId, 'identity should be created');
      assert.ok(existsSync(nestedPath), 'file should exist in nested dir');
    });

    it('should generate valid signing keypair', () => {
      const idPath = join(tmpDir, 'signing_identity.json');
      const identity = loadOrGenerateIdentity(idPath);

      const message = 'test message';
      const signature = signMessage(identity.privateKey, message);
      const isValid = verifySignature(identity.publicKey, message, signature);

      assert.strictEqual(isValid, true, 'signature should verify');
    });
  });

  describe('slug', () => {
    const testCases: [string, string][] = [
      ['My App', 'my-app'],
      ['MyApp', 'myapp'],
      ['my_app', 'my-app'],
      ['My-App', 'my-app'],
      ['  spaced  ', 'spaced'],
      ['UPPERCASE', 'uppercase'],
      ['été', 'ete'],
      ['café', 'cafe'],
      ['niño', 'nino'],
      ['', 'app'],
      ['---', 'app'],
      ['123app', '123app'],
      ['app123', 'app123'],
      ['app@#$%name', 'appname'],
      ['Très Spécial Àpp', 'tres-special-app'],
      ['multiple   spaces', 'multiple-spaces'],
      ['a--b', 'a-b'],
    ];

    for (const [input, expected] of testCases) {
      it(`should convert "${input}" to "${expected}"`, () => {
        assert.strictEqual(slug(input), expected);
      });
    }
  });
});
