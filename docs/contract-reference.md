---
title: Contract Reference
layout: default
nav_order: 4
---

# Contract Reference (v1.0)
{: .no_toc }

A Pacto contract is a YAML file (`pacto.yaml`) that describes a service's operational interface — interfaces, dependencies, runtime behavior, configuration, and scaling. This page covers every section, field, validation rule, and change classification rule.

---

<details open markdown="block">
  <summary>Table of contents</summary>
- TOC
{:toc}
</details>

The canonical JSON Schema is available at [`schema/pacto-v1.0.schema.json`](https://github.com/TrianaLab/pacto/blob/main/internal/validation/schema/pacto-v1.0.schema.json).

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
├── configuration/
│   └── schema.json
├── docs/                    ← optional
│   ├── README.md
│   ├── architecture.md
│   ├── runbook.md
│   └── integration.md
└── sbom/                    ← optional
    └── sbom.spdx.json
```

All files referenced by `pacto.yaml` must exist within the bundle. Validation enforces this.

When you run `pacto push`, the bundle is packaged as an OCI artifact — versioned, content-addressed, and distributable through any OCI registry. This is how contracts travel between teams, services, and environments.

### `docs/` — Optional documentation

The `docs/` directory is an optional convention for including human-readable documentation alongside the contract. Its contents are treated as **informational metadata** — they travel with the contract but have no effect on contract semantics, validation, or diff classification.

**Key properties:**

- **Optional.** Bundles without `docs/` are fully valid. No existing workflow breaks.
- **No contract semantics.** Documentation is not part of the contract. It does not influence validation, diffing, or compatibility checks.
- **Self-contained.** Documentation lives inside the OCI artifact, so it is versioned, distributed, and cached alongside the contract it describes.
- **Format-flexible.** Markdown is the natural default, but there are no restrictions on file formats inside `docs/`.

**What documentation could include:**

- Service overview — what the service does, its purpose within the platform
- Architecture notes — internal design, data flow, key dependencies
- Operational runbooks — incident response, scaling procedures, known failure modes
- Integration guides — how consumers should interact with the service's interfaces

### `sbom/` — Optional Software Bill of Materials

The `sbom/` directory is an optional convention for including a Software Bill of Materials alongside the contract. Like `docs/`, its contents are treated as **informational metadata** — they travel with the contract but have no effect on contract semantics or validation.

Unlike `docs/`, SBOM files **are included in diff output**. When both the old and new bundles contain an SBOM, `pacto diff` reports package-level changes (added, removed, version or license modified). These changes are informational only — they never affect the overall diff classification.

**Supported formats:**

| Format | File extension | Spec version |
|--------|---------------|--------------|
| [SPDX](https://spdx.dev/) | `.spdx.json` | 2.3 |
| [CycloneDX](https://cyclonedx.org/) | `.cdx.json` | 1.5 |

Pacto auto-detects the format by file extension. Place one or more SBOM files directly in the `sbom/` directory (subdirectories are not scanned).

**Key properties:**

- **Optional.** Bundles without `sbom/` are fully valid.
- **Convention-based.** No contract-level field references the SBOM — Pacto discovers it automatically, just like `docs/`.
- **Informational diffing.** `pacto diff` reports SBOM package changes but they don't affect breaking/non-breaking classification.
- **Self-contained.** The SBOM lives inside the OCI artifact, versioned and distributed alongside the contract.

**Generating an SBOM:**

Use any standard SBOM generator. Recommended tools:

- [Syft](https://github.com/anchore/syft) — `syft . -o spdx-json=sbom/sbom.spdx.json`
- [Trivy](https://github.com/aquasecurity/trivy) — `trivy fs . --format spdx-json -o sbom/sbom.spdx.json`
- [cdxgen](https://github.com/CycloneDX/cdxgen) — `cdxgen -o sbom/bom.cdx.json`

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
  chart:
    ref: oci://ghcr.io/acme/payments-chart
    version: 2.1.0

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
  - ref: oci://ghcr.io/acme/auth-pacto@sha256:abc123def456
    required: true
    compatibility: "^2.0.0"

  - ref: oci://ghcr.io/acme/notifications-pacto:1.0.0
    required: false
    compatibility: "~1.0.0"

runtime:
  workload: service

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

  metrics:
    interface: rest-api
    path: /metrics

scaling:
  min: 2
  max: 10

metadata:
  team: payments
  tier: critical
```

---

## Minimal contract

Only `pactoVersion` and `service` are required. All other sections — `interfaces`, `runtime`, `configuration`, `dependencies`, `scaling`, and `metadata` — are optional:

```yaml
pactoVersion: "1.0"

service:
  name: my-library
  version: 1.0.0
```

This is useful for lightweight dependency declarations, shared libraries, or contracts where runtime semantics are managed externally.

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

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | Yes | Pattern: `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$` |
| `version` | string | Yes | Valid semver (e.g., `2.1.0`) |
| `owner` | string | No | |
| `image` | [Image](#image) | No | |
| `chart` | [Chart](#chart) | No | |

#### Image

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `ref` | string | Yes | Non-empty. Valid OCI image reference |
| `private` | boolean | No | |

#### Chart

Optional Helm chart reference for deploying the service.

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `ref` | string | Yes | Non-empty. Local path (e.g. `./charts/my-chart`) or OCI reference (e.g. `oci://ghcr.io/org/chart`) |
| `version` | string | Yes | Non-empty. Valid semver |

{: .warning }
Local chart references are only allowed during development. `pacto push` rejects contracts with local chart references — use an OCI reference before publishing.

---

### `interfaces`

Declares the service's communication boundaries. Optional — a service with no network interfaces (e.g. a batch job or shared library) may omit this section entirely.

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | Yes | Non-empty. Must be unique across interfaces |
| `type` | string | Yes | Enum: `http`, `grpc`, `event` |
| `port` | integer | Conditional | Range: 1-65535. Required for `http` and `grpc` |
| `visibility` | string | No | Enum: `public`, `internal`. Default: `internal` |
| `contract` | string | Conditional | Non-empty. Required for `grpc` and `event` |

#### Conditional requirements

| Interface type | Required fields |
|---|---|
| `http` | `port` |
| `grpc` | `port`, `contract` |
| `event` | `contract` |

{: .note }
Interface names must be unique within a contract. The `contract` field for `http` interfaces is optional but recommended (typically an OpenAPI spec).

---

### `configuration`

Defines the service's configuration model.

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `schema` | string | Yes | Non-empty. Must reference a file in the bundle |
| `values` | object | No | Must conform to the schema defined in `schema` |

Required configuration keys are derived from the JSON Schema's `required` array.

The optional `values` field provides default configuration values that are validated against the referenced JSON Schema. This is useful for documenting expected defaults or providing environment-specific overrides via the `--set` and `--values` flags (see [Contract overrides](#contract-overrides)).

---

### `dependencies`

Declares dependencies on other services via their Pacto contracts.

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `ref` | string | Yes | Non-empty. OCI reference (`oci://...`) or local path (`file://...` or bare path) |
| `required` | boolean | No | Default: `false` |
| `compatibility` | string | Yes | Non-empty. Valid semver constraint |

#### Dependency reference schemes

| Scheme | Example | Description |
|--------|---------|-------------|
| `oci://` | `oci://ghcr.io/acme/auth-pacto:1.0.0` | OCI registry reference (required for `pacto push`) |
| `oci://` (no tag) | `oci://ghcr.io/acme/auth-pacto` | Resolved to the highest semver tag satisfying `compatibility` |
| `file://` | `file://../shared-db` | Local filesystem path |
| *(bare path)* | `../shared-db` | Local filesystem path (shorthand for `file://`) |

When an `oci://` reference omits the tag, pacto queries the registry for available tags and selects the highest semver version that satisfies the `compatibility` constraint. For example, with `compatibility: "^2.0.0"` and available tags `1.0.0`, `2.0.0`, `2.3.0`, `3.0.0`, pacto resolves to `2.3.0`. Tag listings are cached in memory for the duration of the command, so multiple dependencies pointing to the same repository only trigger a single registry query.

#### Compatibility constraint examples

Pacto uses [Masterminds/semver](https://github.com/Masterminds/semver#checking-version-constraints) constraint syntax:

| Constraint | Matches | Use case |
|------------|---------|----------|
| `^2.0.0` | `>= 2.0.0`, `< 3.0.0` | Accept patches and minors within a major version |
| `~2.1.0` | `>= 2.1.0`, `< 2.2.0` | Accept only patches within a minor version |
| `>= 2.0.0` | `2.0.0` and above (including `3.x`, `4.x`, …) | Track the latest version above a floor |
| `>= 2.0.0, < 4.0.0` | `2.x` and `3.x` only | Constrain to a range of major versions |
| `*` | Any version | Always resolve to the absolute latest |

{: .warning }
Local dependency references (`file://` and bare paths) are only allowed during development. `pacto push` rejects contracts with local dependencies — all refs must use `oci://` before publishing.

{: .tip }
Use digest-pinned references (`oci://...@sha256:...`) for production contracts. Tag-based references produce a validation warning.

{: .tip }
If your service depends on a cloud-managed resource (e.g. GCP Cloud SQL, AWS SNS, Azure Service Bus), create a lightweight Pacto contract representing that resource and reference it as a dependency. This makes cloud dependencies explicit and version-tracked alongside your service contracts.

---

### `runtime`

Describes how the service behaves at runtime. This section is what lets platforms make informed deployment decisions without guessing. Optional — a minimal contract (e.g. a lightweight dependency declaration) may omit it entirely.

| Field | Type | Required |
|-------|------|----------|
| `workload` | string | Yes |
| `state` | [State](#state) | Yes |
| `lifecycle` | [Lifecycle](#lifecycle) | No |
| `health` | [Health](#health) | No |
| `metrics` | [Metrics](#metrics) | No |

#### `runtime.workload`

A plain string describing the workload type. Enum: `service`, `job`, `scheduled`.

| Value | Description |
|-------|-------------|
| `service` | A long-running process that serves requests continuously |
| `job` | A one-shot task that runs to completion and then exits |
| `scheduled` | A task that runs on a recurring schedule (e.g. cron) |

#### `runtime.state`

This is one of Pacto's most distinctive features. Instead of platforms guessing whether a service needs persistent storage, stable network identity, or special upgrade procedures, the contract declares it explicitly.

| Field | Type | Required | Enum values |
|-------|------|----------|-------------|
| `type` | string | Yes | `stateless`, `stateful`, `hybrid` |
| `persistence` | [Persistence](#persistence) | Yes | |
| `dataCriticality` | string | Yes | `low`, `medium`, `high` |

**State types:**

| Value | What it means | Example services |
|-------|---------------|------------------|
| `stateless` | No data retained between requests. Any instance can handle any request. Instances are interchangeable. | REST APIs, reverse proxies, API gateways |
| `stateful` | Retains data between requests. Requires stable storage or instance affinity. | Databases, message brokers, distributed caches |
| `hybrid` | Handles requests statelessly but keeps selective in-memory or local state that enriches behavior. Loss of that state degrades but doesn't break the service. | APIs with local caches, services with in-memory session stores |

**How platforms interpret state:**

The combination of `state.type`, `persistence.scope`, and `persistence.durability` tells a platform exactly what infrastructure a service needs:

| State | Persistence | Platform reasoning |
|-------|-------------|--------------------|
| `stateless` + `local/ephemeral` | No persistent storage needed. Horizontally scalable. Use a Deployment with HPA. |
| `stateful` + `local/persistent` | Needs stable identity and local durable storage. Use a StatefulSet with PVCs. |
| `stateful` + `shared/persistent` | Needs durable storage shared across instances. Provision network-attached or shared storage. |
| `hybrid` + `local/ephemeral` | Tolerates instance loss. Can use a Deployment, but consider warm-up time if caches are large. |
| `hybrid` + `local/persistent` | Wants persistent local state but survives without it. StatefulSet with PVC, but can fall back to emptyDir if needed. |

These aren't Kubernetes prescriptions — they're platform-agnostic signals. Whether you deploy to Kubernetes, Nomad, ECS, or a custom platform, the reasoning is the same.

**Data criticality:**

| Value | What it means |
|-------|---------------|
| `low` | Loss of data has minimal impact. Can be regenerated or is non-essential. |
| `medium` | Loss has moderate impact. May require manual recovery. |
| `high` | Loss has severe business impact. Must be prevented. Implies backups, replication, stricter disruption budgets. |

##### Persistence

| Field | Type | Required | Enum values |
|-------|------|----------|-------------|
| `scope` | string | Yes | `local`, `shared` |
| `durability` | string | Yes | `ephemeral`, `persistent` |

- **`local`** — data is confined to a single instance. Not shared across replicas.
- **`shared`** — data is shared across all instances via a common store.
- **`ephemeral`** — data can be lost on restart without impact. Caches, temp files, reconstructible state.
- **`persistent`** — data must survive restarts. Requires durable storage.

##### State invariants

| Condition | Constraint |
|---|---|
| `type: stateless` | `durability` must be `ephemeral` |
| `durability: persistent` | `type` must be `stateful` or `hybrid` |

These invariants are enforced by both the JSON Schema and cross-field validation. A stateless service with persistent storage is a contradiction — validation catches it.

#### `runtime.lifecycle`

Optional. Describes upgrade and shutdown behavior.

| Field | Type | Required | Enum values / Constraints |
|-------|------|----------|-------------|
| `upgradeStrategy` | string | No | `rolling`, `recreate`, `ordered` |
| `gracefulShutdownSeconds` | integer | No | Minimum: 0 |

#### `runtime.health`

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `interface` | string | Yes | Must reference a declared `http` or `grpc` interface |
| `path` | string | Conditional | Required when health interface is `http` |
| `initialDelaySeconds` | integer | No | Minimum: 0 |

#### `runtime.metrics`

Optional. Tells the platform where to scrape metrics from the service (e.g. Prometheus endpoint).

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `interface` | string | Yes | Must reference a declared `http` or `grpc` interface |
| `path` | string | Conditional | Required when metrics interface is `http` |

```yaml
runtime:
  metrics:
    interface: api
    path: /metrics
```

---

### `scaling`

Optional. Defines replica count as either an exact number or a min/max range. Uses one of two mutually exclusive forms.

#### Fixed replica count

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `replicas` | integer | Yes | Minimum: 0. Mutually exclusive with `min`/`max` |

```yaml
scaling:
  replicas: 3
```

#### Auto-scaling range

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `min` | integer | Yes | Minimum: 0. Mutually exclusive with `replicas` |
| `max` | integer | Yes | Minimum: 0. Must be >= `min`. Mutually exclusive with `replicas` |

```yaml
scaling:
  min: 2
  max: 10
```

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

`additionalProperties: false` — no extra fields allowed at any level (except inside `metadata`).

---

## Validation layers

Pacto validates contracts through three successive layers. Each layer short-circuits — if it fails, subsequent layers are skipped.

### Layer 1: Structural (JSON Schema)

Validates against the embedded JSON Schema:
- Field types match
- Required fields are present
- Enum values are valid
- Conditional requirements are met (`http` needs `port`, etc.)
- State invariants are enforced (`stateless` needs `ephemeral`)

### Layer 2: Cross-field

Validates semantic references and consistency:

| Rule | Code |
|---|---|
| `service.version` is valid semver | `INVALID_SEMVER` |
| Interface names are unique | `DUPLICATE_INTERFACE_NAME` |
| `http`/`grpc` interfaces have `port` | `PORT_REQUIRED` |
| `event` interfaces with `port` set | `PORT_IGNORED` (warning) |
| `grpc`/`event` interfaces have `contract` | `CONTRACT_REQUIRED` |
| `health.interface` matches a declared interface | `HEALTH_INTERFACE_NOT_FOUND` |
| Health interface is not `event` type | `HEALTH_INTERFACE_INVALID` |
| `health.path` required for `http` health interface | `HEALTH_PATH_REQUIRED` |
| `health.path` on `grpc` health interface is ignored | `HEALTH_PATH_IGNORED` (warning) |
| `metrics.interface` matches a declared interface | `METRICS_INTERFACE_NOT_FOUND` |
| Metrics interface is not `event` type | `METRICS_INTERFACE_INVALID` |
| `metrics.path` required for `http` metrics interface | `METRICS_PATH_REQUIRED` |
| `metrics.path` on `grpc` metrics interface is ignored | `METRICS_PATH_IGNORED` (warning) |
| Referenced files exist in the bundle | `FILE_NOT_FOUND` |
| OCI dependency refs (`oci://`) are valid OCI references | `INVALID_OCI_REF` |
| Compatibility ranges are valid semver constraints | `INVALID_COMPATIBILITY` |
| Compatibility range is empty | `EMPTY_COMPATIBILITY` |
| OCI dependency uses tag instead of digest | `TAG_NOT_DIGEST` (warning) |
| `image.ref` is a valid OCI reference | `INVALID_IMAGE_REF` |
| `chart.ref` (OCI) is a valid OCI reference | `INVALID_CHART_REF` |
| `chart.version` is valid semver | `INVALID_CHART_VERSION` |
| `configuration.values` without a `configuration.schema` | `VALUES_WITHOUT_SCHEMA` |
| `configuration.values` don't match the schema | `CONFIG_VALUES_VALIDATION_FAILED` |
| `scaling.min` <= `scaling.max` | `SCALING_MIN_EXCEEDS_MAX` |
| Job workloads cannot have scaling | `JOB_SCALING_NOT_ALLOWED` |
| Stateless + persistent is invalid | `STATELESS_PERSISTENT_CONFLICT` |

### Layer 3: Semantic

Validates cross-concern consistency:

| Rule | Code | Type |
|---|---|---|
| `ordered` upgrade strategy with `stateless` state | `UPGRADE_STRATEGY_STATE_MISMATCH` | Warning |

---

## Contract overrides

Pacto supports a Helm-style override system that lets you modify contract values without editing `pacto.yaml` directly. Overrides are available on all commands that take a contract reference (`validate`, `explain`, `diff`, `doc`, `generate`, `graph`, `pack`, `push`).

### Override flags

| Flag | Short | Description |
|------|-------|-------------|
| `--values <file>` | `-f` | Merge a YAML values file into the contract |
| `--set <key=value>` | | Set an individual value using dot-notation |

{: .note }
The `-f` shorthand is not available on `pacto push` because `-f` is already used for `--force`.

### Precedence

Overrides follow the same precedence logic as Helm (lowest to highest):

1. Base `pacto.yaml`
2. Values files (`-f`), in the order they are specified
3. `--set` values

Later sources override earlier ones. Multiple `-f` flags are applied in order — the last file wins for any conflicting key.

### Examples

```bash
# Override the service version
pacto validate my-service --set service.version=2.0.0

# Use a values file for environment-specific overrides
pacto validate my-service -f staging-values.yaml

# Combine both (--set takes precedence over -f)
pacto validate my-service -f staging-values.yaml --set service.version=3.0.0

# Set configuration values validated against the config schema
pacto validate my-service --set configuration.values.DB_HOST=localhost

# Set array elements using bracket notation
pacto validate my-service --set interfaces[0].port=9090
```

### Diff overrides

The `diff` command takes two contract references. To override each independently, use prefixed flags:

| Flag | Description |
|------|-------------|
| `--old-values <file>` | Values file for the old (first) contract |
| `--old-set <key=value>` | Set value for the old contract |
| `--new-values <file>` | Values file for the new (second) contract |
| `--new-set <key=value>` | Set value for the new contract |

```bash
# Override only the new contract
pacto diff oci://ghcr.io/acme/svc-pacto:1.0.0 my-service --new-set service.version=2.0.0

# Override both contracts
pacto diff old-service new-service \
  --old-values old-env.yaml \
  --new-values new-env.yaml \
  --new-set service.version=3.0.0
```

### Type inference

`--set` values are automatically parsed into their most specific type:

| Input | Parsed as |
|-------|-----------|
| `42` | integer |
| `3.14` | float |
| `true` / `false` | boolean |
| `hello` | string |
| `1.0.0` | string (not a float — contains multiple dots) |

### Schema validation

All overrides are validated against the Pacto JSON Schema after they are applied. This means invalid enum values, unknown fields, and type mismatches are rejected by every command — not just `validate` and `push`.

```bash
# Rejected: "invalid" is not a valid enum value for runtime.state.type
pacto explain my-service --set runtime.state.type=invalid

# Rejected: "unknownField" is not defined in the schema
pacto doc my-service --set service.unknownField=value
```

### Environment-specific values files

A common pattern is to maintain per-environment values files that override configuration and scaling for each deployment target. The base `pacto.yaml` defines defaults, and each values file layers environment-specific settings on top.

```
my-service/
├── pacto.yaml
├── values/
│   ├── dev.yaml
│   ├── staging.yaml
│   └── production.yaml
└── configuration/
    └── schema.json
```

```yaml
# values/dev.yaml
configuration:
  values:
    DB_HOST: localhost
    DB_PORT: 5432
    LOG_LEVEL: debug
scaling:
  replicas: 1
```

```yaml
# values/staging.yaml
configuration:
  values:
    DB_HOST: staging-db.internal
    DB_PORT: 5432
    LOG_LEVEL: info
scaling:
  min: 2
  max: 4
```

```yaml
# values/production.yaml
configuration:
  values:
    DB_HOST: prod-db.internal
    DB_PORT: 5432
    LOG_LEVEL: warn
scaling:
  min: 3
  max: 10
```

Validate each environment independently:

```bash
pacto validate my-service -f values/dev.yaml
pacto validate my-service -f values/staging.yaml
pacto validate my-service -f values/production.yaml
```

Compare what changes between environments:

```bash
pacto diff my-service my-service \
  --old-values values/staging.yaml \
  --new-values values/production.yaml
```

Push a release with production configuration:

```bash
pacto push my-service -f values/production.yaml --set service.version=2.1.0
```

### Configuration values validation

Overrides can set `configuration.values` fields. These values are validated against the JSON Schema referenced by `configuration.schema`. If a value has the wrong type or is not defined in the schema, validation fails.

```yaml
# pacto.yaml
configuration:
  schema: configuration/schema.json
```

```bash
# This will fail if DB_PORT expects an integer but receives a string
pacto validate my-service --set configuration.values.DB_PORT=not-a-number
```

---

## Change classification rules

`pacto diff` classifies every detected change using a deterministic rule table. This is what powers breaking change detection in CI pipelines.

Each change is classified as:

- **`BREAKING`** — a change that will break consumers or platforms relying on the previous contract
- **`POTENTIAL_BREAKING`** — a change that *may* break consumers depending on how they use the field
- **`NON_BREAKING`** — a safe change that doesn't affect compatibility

### Service identity

| Field | Change | Classification |
|-------|--------|----------------|
| `service.name` | Modified | **BREAKING** |
| `service.version` | Modified | NON_BREAKING |
| `service.owner` | Added / Modified / Removed | NON_BREAKING |
| `service.image` | Added / Modified / Removed | NON_BREAKING |
| `service.chart` | Added / Modified / Removed | NON_BREAKING |

### Interfaces

| Field | Change | Classification |
|-------|--------|----------------|
| `interfaces` | Added | NON_BREAKING |
| `interfaces` | Removed | **BREAKING** |
| `interfaces.type` | Modified | **BREAKING** |
| `interfaces.port` | Modified | **BREAKING** |
| `interfaces.port` | Added | POTENTIAL_BREAKING |
| `interfaces.port` | Removed | **BREAKING** |
| `interfaces.visibility` | Modified | POTENTIAL_BREAKING |
| `interfaces.contract` | Modified | POTENTIAL_BREAKING |

### Configuration

| Field | Change | Classification |
|-------|--------|----------------|
| `configuration` | Added | NON_BREAKING |
| `configuration` | Removed | **BREAKING** |
| `configuration.schema` | Added | NON_BREAKING |
| `configuration.schema` | Modified | POTENTIAL_BREAKING |
| `configuration.schema` | Removed | **BREAKING** |

### Runtime

| Field | Change | Classification |
|-------|--------|----------------|
| `runtime.workload` | Modified | **BREAKING** |
| `runtime.state.type` | Modified | **BREAKING** |
| `runtime.state.persistence.scope` | Modified | **BREAKING** |
| `runtime.state.persistence.durability` | Modified | **BREAKING** |
| `runtime.state.dataCriticality` | Modified | POTENTIAL_BREAKING |
| `runtime.lifecycle.upgradeStrategy` | Added | NON_BREAKING |
| `runtime.lifecycle.upgradeStrategy` | Modified | POTENTIAL_BREAKING |
| `runtime.lifecycle.upgradeStrategy` | Removed | POTENTIAL_BREAKING |
| `runtime.lifecycle.gracefulShutdownSeconds` | Modified | NON_BREAKING |
| `runtime.health.interface` | Modified | POTENTIAL_BREAKING |
| `runtime.health.path` | Modified | POTENTIAL_BREAKING |
| `runtime.health.initialDelaySeconds` | Modified | NON_BREAKING |
| `runtime.metrics.interface` | Modified | POTENTIAL_BREAKING |
| `runtime.metrics.path` | Modified | POTENTIAL_BREAKING |

### Scaling

| Field | Change | Classification |
|-------|--------|----------------|
| `scaling` | Added | NON_BREAKING |
| `scaling` | Removed | POTENTIAL_BREAKING |
| `scaling.replicas` | Modified | POTENTIAL_BREAKING |
| `scaling.min` | Modified | POTENTIAL_BREAKING |
| `scaling.max` | Modified | NON_BREAKING |

### Dependencies

| Field | Change | Classification |
|-------|--------|----------------|
| `dependencies` | Added | NON_BREAKING |
| `dependencies` | Removed | **BREAKING** |
| `dependencies.compatibility` | Modified | POTENTIAL_BREAKING |
| `dependencies.required` | Modified | POTENTIAL_BREAKING |

### OpenAPI

`pacto diff` performs deep comparison of referenced OpenAPI specs, detecting changes at the path, method, parameter, request body, and response level.

#### Paths

| Field | Change | Classification |
|-------|--------|----------------|
| `openapi.paths` | Added | NON_BREAKING |
| `openapi.paths` | Removed | **BREAKING** |

#### Methods

| Field | Change | Classification |
|-------|--------|----------------|
| `openapi.methods` | Added | NON_BREAKING |
| `openapi.methods` | Removed | **BREAKING** |

#### Parameters

| Field | Change | Classification |
|-------|--------|----------------|
| `openapi.parameters` | Added | POTENTIAL_BREAKING |
| `openapi.parameters` | Removed | **BREAKING** |
| `openapi.parameters` | Modified | POTENTIAL_BREAKING |

Parameters are identified by `name` + `in` (location: query, path, header, cookie). A parameter renamed or moved to a different location is treated as a removal + addition.

#### Request body

| Field | Change | Classification |
|-------|--------|----------------|
| `openapi.request-body` | Added | POTENTIAL_BREAKING |
| `openapi.request-body` | Removed | POTENTIAL_BREAKING |
| `openapi.request-body` | Modified | POTENTIAL_BREAKING |

#### Responses

| Field | Change | Classification |
|-------|--------|----------------|
| `openapi.responses` | Added | NON_BREAKING |
| `openapi.responses` | Removed | **BREAKING** |
| `openapi.responses` | Modified | POTENTIAL_BREAKING |

Change paths use a hierarchical format that pinpoints the exact location, for example:

```
openapi.paths[/users].methods[GET].parameters[filter:query]
openapi.paths[/users].methods[POST].request-body
openapi.paths[/users].methods[GET].responses[200]
```

### JSON Schema properties

| Field | Change | Classification |
|-------|--------|----------------|
| `schema.properties` | Added | NON_BREAKING |
| `schema.properties` | Removed | **BREAKING** |

### SBOM

SBOM changes are reported separately from contract changes. They are **informational only** and never affect the overall diff classification.

| Change | Description |
|--------|-------------|
| Package added | A new package appears in the SBOM |
| Package removed | A package no longer appears in the SBOM |
| Package version modified | A package's version changed |
| Package license modified | A package's license changed |
| Package supplier modified | A package's supplier changed |

Unknown changes default to **POTENTIAL_BREAKING**.
