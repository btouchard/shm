// SPDX-License-Identifier: MIT

import { existsSync, readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname } from 'node:path';
import { randomUUID } from 'node:crypto';
import type { Identity } from './types.js';
import { generateKeypair } from './crypto.js';

/**
 * Loads an existing identity from file or generates a new one.
 * The identity file is created with restrictive permissions (0600).
 *
 * @param filePath - Path to the identity JSON file.
 * @returns The loaded or newly generated identity.
 * @throws Error if the file cannot be read or written.
 */
export function loadOrGenerateIdentity(filePath: string): Identity {
  // Try to load existing identity
  if (existsSync(filePath)) {
    try {
      const data = readFileSync(filePath, 'utf-8');
      const parsed = JSON.parse(data);

      // Validate the parsed identity
      if (
        typeof parsed.instanceId === 'string' &&
        typeof parsed.privateKey === 'string' &&
        typeof parsed.publicKey === 'string'
      ) {
        return parsed as Identity;
      }
    } catch {
      // File exists but is corrupted, regenerate
    }
  }

  // Generate new identity
  const { publicKey, privateKey } = generateKeypair();

  const identity: Identity = {
    instanceId: randomUUID(),
    privateKey,
    publicKey,
  };

  // Ensure directory exists
  const dir = dirname(filePath);
  if (dir && dir !== '.' && !existsSync(dir)) {
    mkdirSync(dir, { recursive: true, mode: 0o755 });
  }

  // Write identity file with restrictive permissions
  writeFileSync(filePath, JSON.stringify(identity, null, 2), {
    mode: 0o600,
    encoding: 'utf-8',
  });

  return identity;
}

/**
 * Converts a string to a URL-safe slug.
 * Handles accented characters and special characters.
 *
 * @param str - Input string to slugify.
 * @returns Lowercase slug with hyphens.
 */
export function slug(str: string): string {
  const replacements: Record<string, string> = {
    'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a',
    'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e',
    'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i',
    'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
    'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u',
    'ý': 'y', 'ÿ': 'y',
    'ñ': 'n', 'ç': 'c',
  };

  let result = '';
  const lower = str.toLowerCase();

  for (const char of lower) {
    if (replacements[char]) {
      result += replacements[char];
    } else if (/[a-z0-9]/.test(char)) {
      result += char;
    } else if (char === ' ' || char === '-' || char === '_') {
      result += '-';
    }
    // Other characters are dropped
  }

  // Remove multiple consecutive hyphens
  result = result.replace(/-+/g, '-');

  // Trim hyphens from start and end
  result = result.replace(/^-+|-+$/g, '');

  return result || 'app';
}
