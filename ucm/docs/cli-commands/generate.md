---
title: "ucm generate"
description: "Scan an existing account + metastore + workspace and emit ucm.yml + seed state."
---

# ucm generate

Points at a workspace and writes a starter `ucm.yml` plus seed direct-engine
state so a subsequent `ucm deploy` is a no-op instead of trying to recreate
everything.

## Syntax

```bash
databricks ucm generate [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--output-dir` | Directory to write `ucm.yml` + seed state into. Default: `.`. |
| `--kinds` | Comma-separated resource kinds to scan. Default: `catalog,schema,storage_credential,external_location,volume,connection`. |
| `--name` | Value for the `ucm.name` field. Default: sanitized host-derived label (e.g. `acme-prod`). |
| `--force`, `-f` | Overwrite an existing `ucm.yml` in `--output-dir`. |

## Examples

```bash
databricks ucm generate --name prod
databricks ucm generate --output-dir ./bootstrap --kinds catalog,schema
databricks ucm generate --force
```

## What is scanned

Default kinds: `catalog`, `schema`, `storage_credential`,
`external_location`, `volume`, `connection`.

Grants are intentionally excluded — they reconcile per-securable at deploy
time, so seeding them adds noise without improving idempotency.

## Known limitations

- **Credentials with secret material** (`azure_service_principal.client_secret`)
  cannot round-trip — UC does not echo the secret. The generated YAML leaves
  a placeholder; fill it in before the next deploy.
- **Grants** are skipped. Re-emit them with `ucm deployment bind` (or by
  authoring them manually) after generation.
- **Catalog→schema tag inheritance** is not reconstructed; tags are emitted
  as-declared on each resource. Clean up by hand if desired.

## See also

- [init](init.md)
- [import](import.md)
- [deployment bind](deployment-bind.md)
