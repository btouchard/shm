// SPDX-License-Identifier: MIT

/**
 * Configuration for the SHM client.
 */
export interface Config {
  /** Base URL of the SHM server (e.g., "https://telemetry.example.com") */
  serverUrl: string;
  /** Name of your application */
  appName: string;
  /** Version of your application */
  appVersion: string;
  /** Directory where identity file will be stored (default: ".") */
  dataDir?: string;
  /** Environment identifier (e.g., "production", "staging") */
  environment?: string;
  /** Enable or disable telemetry (default: true) */
  enabled?: boolean;
  /** Interval between snapshots in milliseconds (default: 3600000 = 1 hour, minimum: 60000 = 1 minute) */
  reportIntervalMs?: number;
}

/**
 * Function that returns custom metrics to include in snapshots.
 */
export type MetricsProvider = () => Record<string, unknown> | Promise<Record<string, unknown>>;

/**
 * Identity stored locally for each instance.
 */
export interface Identity {
  instanceId: string;
  privateKey: string; // Hex encoded
  publicKey: string;  // Hex encoded
}

/**
 * Payload for instance registration (sent to /v1/register).
 */
export interface RegisterRequest {
  instance_id: string;
  public_key: string;
  app_name: string;
  app_version: string;
  deployment_mode?: string;
  environment?: string;
  os_arch: string;
}

/**
 * Payload for snapshot submission (sent to /v1/snapshot).
 */
export interface SnapshotRequest {
  instance_id: string;
  timestamp: string; // ISO 8601 format
  metrics: Record<string, unknown>;
}

/**
 * System metrics collected automatically.
 */
export interface SystemMetrics {
  sys_os: string;
  sys_arch: string;
  sys_cpu_cores: number;
  sys_node_version: string;
  sys_mode: string;
  app_mem_heap_mb: number;
  app_mem_rss_mb: number;
  app_uptime_h?: number;
}
