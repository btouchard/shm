// SPDX-License-Identifier: MIT

import { describe, it, beforeEach, afterEach } from 'node:test';
import assert from 'node:assert';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { collectSystemMetricsFromEnv } from './client.js';

describe('collectSystemMetricsFromEnv', () => {
  const originalEnv = process.env['SHM_COLLECT_SYSTEM_METRICS'];

  afterEach(() => {
    if (originalEnv === undefined) {
      delete process.env['SHM_COLLECT_SYSTEM_METRICS'];
    } else {
      process.env['SHM_COLLECT_SYSTEM_METRICS'] = originalEnv;
    }
  });

  it('should return true when env is not set', () => {
    delete process.env['SHM_COLLECT_SYSTEM_METRICS'];
    assert.strictEqual(collectSystemMetricsFromEnv(), true);
  });

  it('should return true when env is "true"', () => {
    process.env['SHM_COLLECT_SYSTEM_METRICS'] = 'true';
    assert.strictEqual(collectSystemMetricsFromEnv(), true);
  });

  it('should return true when env is "TRUE"', () => {
    process.env['SHM_COLLECT_SYSTEM_METRICS'] = 'TRUE';
    assert.strictEqual(collectSystemMetricsFromEnv(), true);
  });

  it('should return true when env is "1"', () => {
    process.env['SHM_COLLECT_SYSTEM_METRICS'] = '1';
    assert.strictEqual(collectSystemMetricsFromEnv(), true);
  });

  it('should return false when env is "false"', () => {
    process.env['SHM_COLLECT_SYSTEM_METRICS'] = 'false';
    assert.strictEqual(collectSystemMetricsFromEnv(), false);
  });

  it('should return false when env is "FALSE"', () => {
    process.env['SHM_COLLECT_SYSTEM_METRICS'] = 'FALSE';
    assert.strictEqual(collectSystemMetricsFromEnv(), false);
  });

  it('should return false when env is "0"', () => {
    process.env['SHM_COLLECT_SYSTEM_METRICS'] = '0';
    assert.strictEqual(collectSystemMetricsFromEnv(), false);
  });

  it('should return true when env is any other value', () => {
    process.env['SHM_COLLECT_SYSTEM_METRICS'] = 'anything';
    assert.strictEqual(collectSystemMetricsFromEnv(), true);
  });
});

describe('SHMClient', () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), 'shm-client-test-'));
  });

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true });
  });

  // Dynamic import to avoid issues with ESM
  async function importClient() {
    const { SHMClient } = await import('./client.js');
    return SHMClient;
  }

  describe('collectSystemMetrics config', () => {
    it('should default to true when not specified and env not set', async () => {
      delete process.env['SHM_COLLECT_SYSTEM_METRICS'];
      const SHMClient = await importClient();

      const client = new SHMClient({
        serverUrl: 'http://localhost:8080',
        appName: 'test-app',
        appVersion: '1.0.0',
        dataDir: tmpDir,
      });

      // Access private config via any cast
      const config = (client as any).config;
      assert.strictEqual(config.collectSystemMetrics, true);
    });

    it('should respect explicit false config', async () => {
      const SHMClient = await importClient();

      const client = new SHMClient({
        serverUrl: 'http://localhost:8080',
        appName: 'test-app',
        appVersion: '1.0.0',
        dataDir: tmpDir,
        collectSystemMetrics: false,
      });

      const config = (client as any).config;
      assert.strictEqual(config.collectSystemMetrics, false);
    });

    it('should respect explicit true config', async () => {
      process.env['SHM_COLLECT_SYSTEM_METRICS'] = 'false';
      const SHMClient = await importClient();

      const client = new SHMClient({
        serverUrl: 'http://localhost:8080',
        appName: 'test-app',
        appVersion: '1.0.0',
        dataDir: tmpDir,
        collectSystemMetrics: true,
      });

      const config = (client as any).config;
      assert.strictEqual(config.collectSystemMetrics, true);

      delete process.env['SHM_COLLECT_SYSTEM_METRICS'];
    });
  });

  describe('DO_NOT_TRACK', () => {
    it('should disable client when DO_NOT_TRACK=true', async () => {
      process.env['DO_NOT_TRACK'] = 'true';
      const SHMClient = await importClient();

      const client = new SHMClient({
        serverUrl: 'http://localhost:8080',
        appName: 'test-app',
        appVersion: '1.0.0',
        dataDir: tmpDir,
        enabled: true, // explicitly enabled
      });

      const config = (client as any).config;
      assert.strictEqual(config.enabled, false);

      delete process.env['DO_NOT_TRACK'];
    });

    it('should disable client when DO_NOT_TRACK=1', async () => {
      process.env['DO_NOT_TRACK'] = '1';
      const SHMClient = await importClient();

      const client = new SHMClient({
        serverUrl: 'http://localhost:8080',
        appName: 'test-app',
        appVersion: '1.0.0',
        dataDir: tmpDir,
        enabled: true,
      });

      const config = (client as any).config;
      assert.strictEqual(config.enabled, false);

      delete process.env['DO_NOT_TRACK'];
    });

    it('should not disable client when DO_NOT_TRACK=false', async () => {
      process.env['DO_NOT_TRACK'] = 'false';
      const SHMClient = await importClient();

      const client = new SHMClient({
        serverUrl: 'http://localhost:8080',
        appName: 'test-app',
        appVersion: '1.0.0',
        dataDir: tmpDir,
        enabled: true,
      });

      const config = (client as any).config;
      assert.strictEqual(config.enabled, true);

      delete process.env['DO_NOT_TRACK'];
    });

    it('should not disable client when DO_NOT_TRACK is not set', async () => {
      delete process.env['DO_NOT_TRACK'];
      const SHMClient = await importClient();

      const client = new SHMClient({
        serverUrl: 'http://localhost:8080',
        appName: 'test-app',
        appVersion: '1.0.0',
        dataDir: tmpDir,
        enabled: true,
      });

      const config = (client as any).config;
      assert.strictEqual(config.enabled, true);
    });
  });
});