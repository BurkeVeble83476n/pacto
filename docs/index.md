---
title: Home
layout: home
nav_order: 1
---

# Pacto

**Pacto** (/ˈpak.to/ — from Spanish: *pact*, *agreement*) is an open, OCI-distributed contract standard for cloud-native services.

It provides a single, declarative source of truth that bridges the gap between **developers** who build services and **platform engineers** who run them. A Pacto contract captures everything a platform needs to know about a service — its interfaces, configuration, runtime semantics, dependencies, and scaling requirements — without assuming any specific infrastructure.

---

## The problem

Modern cloud systems suffer from a recurring misalignment:

- Developers describe APIs but not runtime behavior.
- Platform engineers describe infrastructure but lack service context.
- CI systems validate fragments, not the whole picture.
- Documentation drifts from reality.
- State and persistence assumptions are implicit.
- Dependencies are loosely defined and unversioned.

**There is no shared operational contract.**

## The solution

Pacto introduces a **formal service contract** — a versioned, machine-validated YAML file that:

- Is **declarative** — describes *what*, not *how*
- Is **machine-validated** — three-layer validation catches errors before deployment
- Is **OCI-distributed** — bundles are versioned artifacts in any OCI registry
- Encodes **state semantics explicitly** — stateless, stateful, or hybrid
- Enables **deterministic platform behavior** — no guessing about workload type or persistence
- Remains **infrastructure-agnostic** — no Kubernetes, no cloud provider, no platform assumptions

```yaml
pactoVersion: "1.0"

service:
  name: payments-api
  version: 2.1.0
  owner: team/payments

interfaces:
  - name: rest-api
    type: http
    port: 8080
    visibility: public
    contract: interfaces/openapi.yaml

runtime:
  workload:
    type: service
    concurrency: long-lived
  state:
    type: stateful
    persistence:
      scope: local
      durability: persistent
    dataCriticality: high
  health:
    interface: rest-api
    path: /health

dependencies:
  - ref: ghcr.io/acme/auth-pacto@sha256:abc123
    required: true
    compatibility: "^2.0.0"
```

---

## Key capabilities

```mermaid
graph LR
    A[pacto init] --> B[pacto validate]
    B --> C[pacto pack]
    C --> D[pacto push]
    D --> E[OCI Registry]
    E --> F[pacto pull]
    E --> G[pacto diff]
    E --> H[pacto graph]
    E --> I[pacto generate]
```

| Command | Purpose |
|---------|---------|
| `pacto init` | Scaffold a new service contract |
| `pacto validate` | Three-layer validation (structural, cross-field, semantic) |
| `pacto pack` | Create an OCI-ready tar.gz bundle |
| `pacto push` | Push bundle to any OCI registry |
| `pacto pull` | Pull bundle from an OCI registry |
| `pacto diff` | Compare two contracts and classify changes |
| `pacto graph` | Resolve and visualize the dependency graph |
| `pacto explain` | Human-readable contract summary |
| `pacto generate` | Generate deployment artifacts via plugins |
| `pacto login` | Authenticate with an OCI registry |
| `pacto version` | Print version information |

---

## Who is Pacto for?

### Developers
Define your service's operational interface alongside your code. Declare interfaces, configuration schema, health checks, and dependencies. Validate locally before pushing. [Learn more]({{ site.baseurl }}{% link developers.md %})

### Platform Engineers
Consume contracts to generate deployment manifests, enforce policies, detect breaking changes, and build dependency graphs. [Learn more]({{ site.baseurl }}{% link platform-engineers.md %})

---

## What Pacto is not

- **Not a deployment tool** — it describes *what* to deploy, not *how*
- **Not a registry** — it uses existing OCI registries (GHCR, ECR, ACR, Docker Hub)
- **Not a cloud abstraction** — no provider-specific constructs
- **Not a replacement for Helm or Terraform** — it complements them as input
- **Not a CI system** — it integrates into any CI pipeline

Pacto is a **contract standard**.
