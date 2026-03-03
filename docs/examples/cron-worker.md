---
title: Cron Worker
layout: default
parent: Examples
nav_order: 5
---

# Cron Worker

A Pacto contract for a scheduled batch job — a stateless worker that runs on a cron schedule.

```yaml
pactoVersion: "1.0"

service:
  name: report-generator
  version: 1.2.0
  owner: team/analytics
  image:
    ref: ghcr.io/acme/report-generator:1.2.0
    private: true

interfaces:
  - name: health
    type: http
    port: 8080
    visibility: internal

runtime:
  workload:
    type: scheduled
    concurrency: finite

  state:
    type: stateless
    persistence:
      scope: local
      durability: ephemeral
    dataCriticality: low

  health:
    interface: health
    path: /ready

metadata:
  schedule: "0 2 * * *"
  timeout: 3600
  team: analytics
```

### Key decisions

- **`workload.type: scheduled`** — runs on a cron schedule, not continuously
- **`concurrency: finite`** — each execution starts, processes, and exits
- **No `scaling` section** — job workloads don't scale horizontally (enforced by validation)
- **No `lifecycle` section** — upgrade strategy doesn't apply to jobs
- **Schedule in `metadata`** — the cron expression is platform-specific, so it belongs in metadata rather than in the contract's core fields

### Variant: One-shot job

For a job that runs once (e.g., a database migration):

```yaml
runtime:
  workload:
    type: job
    concurrency: finite
  state:
    type: stateless
    persistence:
      scope: local
      durability: ephemeral
    dataCriticality: low
  health:
    interface: health
    path: /ready
```
