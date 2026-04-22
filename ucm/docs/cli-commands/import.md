---
title: "ucm import"
description: "Bind state for an existing UC resource to a ucm.yml declaration."
---

# ucm import

Binds state for an existing UC resource to a `ucm.yml` declaration — no UC
object is created or modified. Use this to reconcile a drifted or imported
resource without running `ucm generate` on the entire metastore.

## Syntax

```bash
databricks ucm import <kind> <name> [flags]
```

## Arguments

| Arg | Description |
|---|---|
| `<kind>` | `catalog`, `schema`, `storage_credential`, `external_location`, `volume`, or `connection`. |
| `<name>` | The UC identifier. E.g. `sales_prod` for a catalog, `sales.raw` for a schema, `sales.raw.docs` for a volume. |

## Flags

| Flag | Description |
|---|---|
| `--key` | `ucm.yml` map key to bind under. Defaults to the UC name (or its last path component for schemas/volumes). |
| `--auto-approve` | Skip the interactive confirmation prompt. |
| `--target`, `-t` | Target to import into. |

## Examples

```bash
databricks ucm import catalog sales_prod
databricks ucm import schema sales_prod.raw --key raw_landing
databricks ucm import storage_credential my_cred --auto-approve
```

## Prerequisites

The resource must already be declared under `resources.<kind>.<key>` in
`ucm.yml`. Without a declaration, ucm has no target to bind the imported
state to.

## See also

- [generate](generate.md) — bulk version of `import` for an entire metastore.
- [deployment bind](deployment-bind.md) — similar, but scoped to re-binding ucm keys to already-discovered UC objects.
