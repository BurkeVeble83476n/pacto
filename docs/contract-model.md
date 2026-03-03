---
title: Contract Model
layout: default
nav_order: 4
---

# Contract Model (v1.0)

A Pacto contract is a YAML file (`pacto.yaml`) that describes a service's operational interface. This page explains every section and field.

---

## Bundle structure

A Pacto bundle is a self-contained directory (or OCI artifact) with the following layout:

```
/
├── pacto.yaml
├── interfaces/
│   ├── openapi.yaml
│   ├── service.proto
│   └── events.yaml
└── configuration/
    └── schema.json
```

All files referenced by `pacto.yaml` must exist within the bundle.

---

## Full example

```yaml
pactoVersion: "1.0"

service:
  name: payments-api
  version: 2.1.0
  owner: team/payments
  image:
    ref: ghcr.io/acme/payments-api:2.1.0
    private: true

interfaces:
  - name: rest-api
    type: http
    port: 8080
    visibility: public
    contract: interfaces/openapi.yaml

  - name: grpc-api
    type: grpc
    port: 9090
    visibility: internal
    contract: interfaces/service.proto

  - name: order-events
    type: event
    visibility: internal
    contract: interfaces/events.yaml

configuration:
  schema: configuration/schema.json

dependencies:
  - ref: ghcr.io/acme/auth-pacto@sha256:abc123def456
    required: true
    compatibility: "^2.0.0"

  - ref: ghcr.io/acme/notifications-pacto:1.0.0
    required: false
    compatibility: "~1.0.0"

runtime:
  workload:
    type: service
    concurrency: long-lived

  network:
    defaultInterface: rest-api

  state:
    type: stateful
    persistence:
      scope: local
      durability: persistent
    dataCriticality: high

  lifecycle:
    upgradeStrategy: ordered
    gracefulShutdownSeconds: 30

  health:
    interface: rest-api
    path: /health
    initialDelaySeconds: 15

scaling:
  min: 2
  max: 10

metadata:
  team: payments
  tier: critical
```

---

## Sections

### `pactoVersion`

The contract specification version. Currently only `"1.0"` is supported.

```yaml
pactoVersion: "1.0"
```

---

### `service`

Identifies the service.

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | DNS-compatible name (`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`) |
| `version` | Yes | Semantic version (e.g., `2.1.0`) |
| `owner` | No | Team or individual owner identifier |
| `image` | No | Container image reference |

#### `service.image`

| Field | Required | Description |
|-------|----------|-------------|
| `ref` | Yes | OCI image reference (e.g., `ghcr.io/acme/api:2.1.0`) |
| `private` | No | Whether the image requires authentication |

---

### `interfaces`

Declares the service's communication boundaries. **At least one interface is required.**

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique interface name |
| `type` | Yes | `http`, `grpc`, or `event` |
| `port` | http, grpc | Port number (1-65535). Required for http and grpc |
| `visibility` | No | `public` or `internal` (default: `internal`) |
| `contract` | grpc, event | Path to interface contract file in the bundle. Required for grpc and event |

{: .note }
Interface names must be unique within a contract. The `contract` field for `http` interfaces is optional but recommended (typically an OpenAPI spec).

---

### `configuration`

Defines the service's configuration model.

| Field | Required | Description |
|-------|----------|-------------|
| `schema` | Yes | Path to a JSON Schema file within the bundle |

Required configuration keys are derived from the JSON Schema's `required` array.

---

### `dependencies`

Declares dependencies on other services via their Pacto contracts.

| Field | Required | Description |
|-------|----------|-------------|
| `ref` | Yes | OCI reference to the dependency's Pacto bundle |
| `required` | No | Whether the dependency is mandatory (default: `false`) |
| `compatibility` | Yes | Semver constraint (e.g., `^2.0.0`, `~1.0.0`, `>=1.2.3`) |

{: .tip }
Use digest-pinned references (`@sha256:...`) for production contracts. Tag-based references produce a validation warning.

---

### `runtime`

Describes how the service behaves at runtime. This is the most important section for platform engineers.

#### `runtime.workload`

| Field | Required | Values |
|-------|----------|--------|
| `type` | Yes | `service` — long-running process |
|  |  | `worker` — background processor |
|  |  | `job` — runs to completion |
|  |  | `scheduled` — runs on a schedule |
| `concurrency` | Yes | `long-lived` — persistent connections |
|  |  | `finite` — request-response |
|  |  | `event-driven` — reacts to events |

#### `runtime.network`

| Field | Required | Description |
|-------|----------|-------------|
| `defaultInterface` | No | Must reference a declared interface name |

#### `runtime.state`

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `stateless`, `stateful`, or `hybrid` |
| `persistence.scope` | Yes | `local` or `shared` |
| `persistence.durability` | Yes | `ephemeral` or `persistent` |
| `dataCriticality` | Yes | `low`, `medium`, or `high` |

**State invariants:**
- `stateless` **requires** `durability: ephemeral`
- `persistent` durability **requires** `stateful` or `hybrid`

These invariants are enforced by both the JSON Schema and cross-field validation.

#### `runtime.lifecycle`

Optional. Describes upgrade and shutdown behavior.

| Field | Required | Values |
|-------|----------|--------|
| `upgradeStrategy` | No | `rolling`, `recreate`, `ordered` |
| `gracefulShutdownSeconds` | No | Integer >= 0 |

#### `runtime.health`

| Field | Required | Description |
|-------|----------|-------------|
| `interface` | Yes | Must reference a declared http or grpc interface |
| `path` | http | Required when the health interface type is `http` |
| `initialDelaySeconds` | No | Integer >= 0 |

---

### `scaling`

Optional. Defines replica bounds.

| Field | Required | Description |
|-------|----------|-------------|
| `min` | Yes | Minimum replicas (>= 0) |
| `max` | Yes | Maximum replicas (>= 0, must be >= `min`) |

{: .warning }
Scaling must not be applied to `job` workloads.

---

### `metadata`

Optional. Free-form key-value pairs for organizational use. Not validated beyond type.

```yaml
metadata:
  team: payments
  tier: critical
  on-call: "#payments-oncall"
```
