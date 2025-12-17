# SHM API Reference

This document describes the REST API used by instances to report telemetry data to the SHM server.

## Overview

SHM uses Ed25519 cryptographic signatures for request authentication. Each instance generates a keypair on first run and registers its public key with the server.

### Base URL

```
https://your-shm-server.example.com
```

### Authentication Flow

1. **Register** - Instance sends its public key (unauthenticated)
2. **Activate** - Instance proves ownership by signing a request
3. **Snapshot** - Instance periodically sends signed metrics

## Endpoints

### POST /v1/register

Register a new instance with the server. This is the only unauthenticated endpoint.

**Request Body:**

```json
{
  "instance_id": "unique-uuid-v4",
  "public_key": "hex-encoded-ed25519-public-key",
  "app_name": "my-app",
  "app_version": "1.2.0",
  "deployment_mode": "docker",
  "environment": "production",
  "os_arch": "linux/amd64"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `instance_id` | string | Yes | Unique identifier (UUID v4 recommended) |
| `public_key` | string | Yes | Ed25519 public key, hex-encoded (64 chars) |
| `app_name` | string | Yes | Name of your application |
| `app_version` | string | Yes | Version string |
| `deployment_mode` | string | No | How the app is deployed (docker, binary, kubernetes...) |
| `environment` | string | No | Environment name (production, staging, dev...) |
| `os_arch` | string | No | OS and architecture (linux/amd64, darwin/arm64...) |

**Response:**

```json
{
  "status": "ok",
  "message": "Registered"
}
```

**Status Codes:**

| Code | Description |
|------|-------------|
| 201 | Instance registered successfully |
| 400 | Invalid JSON body |
| 405 | Method not allowed (use POST) |
| 500 | Server error |

---

### POST /v1/activate

Activate a registered instance. This proves the instance owns the private key corresponding to the registered public key.

**Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| `X-Instance-ID` | Yes | The instance_id used during registration |
| `X-Signature` | Yes | Ed25519 signature of the request body, hex-encoded |

**Request Body:**

The body can be empty (`{}`) or contain any valid JSON. The signature is computed over the exact body bytes.

```json
{}
```

**Response:**

```json
{
  "status": "active",
  "message": "Instance activated successfully"
}
```

**Status Codes:**

| Code | Description |
|------|-------------|
| 200 | Instance activated |
| 401 | Missing authentication headers |
| 403 | Invalid signature or unknown instance |
| 405 | Method not allowed |
| 500 | Server error |

---

### POST /v1/snapshot

Send a metrics snapshot. Must be called periodically (recommended: every 60 seconds).

**Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| `X-Instance-ID` | Yes | The instance_id |
| `X-Signature` | Yes | Ed25519 signature of the request body |

**Request Body:**

```json
{
  "instance_id": "unique-uuid-v4",
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": {
    "users_count": 150,
    "documents_count": 1234,
    "cpu_percent": 12.5,
    "memory_mb": 512,
    "custom_metric": 42
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `instance_id` | string | Yes | The instance_id |
| `timestamp` | string | Yes | ISO 8601 timestamp |
| `metrics` | object | Yes | Arbitrary key-value metrics (schema-agnostic) |

The `metrics` field accepts any JSON object. You define what metrics matter for your application.

**Response:**

```json
{
  "status": "ok",
  "message": "Snapshot received"
}
```

**Status Codes:**

| Code | Description |
|------|-------------|
| 202 | Snapshot accepted |
| 400 | Invalid JSON |
| 401 | Missing authentication headers |
| 403 | Invalid signature |
| 405 | Method not allowed |
| 500 | Server error |

---

## Cryptographic Signature

SHM uses Ed25519 for request signing. Here's how to implement it:

### Key Generation

Generate a 32-byte seed, then derive the keypair:

```
seed = random(32 bytes)
(publicKey, privateKey) = ed25519.GenerateKey(seed)
```

The public key is 32 bytes. Encode it as hex (64 characters) for the API.

### Signing a Request

1. Serialize the request body as JSON bytes
2. Sign the bytes with Ed25519: `signature = ed25519.Sign(privateKey, bodyBytes)`
3. Hex-encode the signature (128 characters)
4. Set `X-Signature` header to the hex-encoded signature

### Verification (Server-side)

The server reconstructs the signature verification:

```
pubKey = hex.Decode(instance.public_key)
signature = hex.Decode(request.Header["X-Signature"])
valid = ed25519.Verify(pubKey, request.Body, signature)
```

### Example (Go)

```go
import (
    "crypto/ed25519"
    "encoding/hex"
    "encoding/json"
)

// Generate keypair (once, on first run)
publicKey, privateKey, _ := ed25519.GenerateKey(nil)
publicKeyHex := hex.EncodeToString(publicKey)

// Sign a request
body := map[string]any{"instance_id": instanceID, "timestamp": time.Now()}
bodyBytes, _ := json.Marshal(body)
signature := ed25519.Sign(privateKey, bodyBytes)
signatureHex := hex.EncodeToString(signature)

// Set headers
req.Header.Set("X-Instance-ID", instanceID)
req.Header.Set("X-Signature", signatureHex)
```

### Example (Python)

```python
from nacl.signing import SigningKey
import json

# Generate keypair (once, on first run)
signing_key = SigningKey.generate()
public_key_hex = signing_key.verify_key.encode().hex()

# Sign a request
body = {"instance_id": instance_id, "timestamp": "2024-01-15T10:30:00Z"}
body_bytes = json.dumps(body, separators=(',', ':')).encode()
signature = signing_key.sign(body_bytes).signature.hex()

# Set headers
headers = {
    "X-Instance-ID": instance_id,
    "X-Signature": signature,
    "Content-Type": "application/json"
}
```

### Example (Node.js)

```javascript
import { generateKeyPairSync, sign } from 'crypto';

// Generate keypair (once, on first run)
const { publicKey, privateKey } = generateKeyPairSync('ed25519');
const publicKeyHex = publicKey.export({ type: 'spki', format: 'der' })
    .slice(-32).toString('hex');

// Sign a request
const body = JSON.stringify({ instance_id: instanceId, timestamp: new Date().toISOString() });
const signature = sign(null, Buffer.from(body), privateKey).toString('hex');

// Set headers
const headers = {
    'X-Instance-ID': instanceId,
    'X-Signature': signature,
    'Content-Type': 'application/json'
};
```

---

## Error Responses

All errors return a plain text message with an appropriate HTTP status code:

| Status | Message | Cause |
|--------|---------|-------|
| 400 | `JSON invalide` | Malformed request body |
| 401 | `Headers d'authentification manquants` | Missing X-Instance-ID or X-Signature |
| 403 | `Non autorisé` | Instance not found |
| 403 | `Signature invalide` | Signature verification failed |
| 405 | `Method not allowed` | Wrong HTTP method |
| 500 | `Erreur serveur` | Internal server error |

---

## Rate Limiting

The server does not enforce rate limits. However, we recommend:

- **Snapshots**: Every 60 seconds
- **Register**: Once per instance lifetime
- **Activate**: Once after registration

---

## SDK

An official Go SDK is available at [`sdk/golang/`](../sdk/golang/). It handles keypair generation, storage, registration, and periodic snapshot sending automatically.

```go
import "github.com/btouchard/shm/sdk/golang"

tracker := shm.NewTracker("https://shm.example.com", "my-app", "1.0.0")
tracker.Start(func() map[string]any {
    return map[string]any{
        "users": getUserCount(),
        "documents": getDocCount(),
    }
})
```
