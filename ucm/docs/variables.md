---
title: "Variables"
description: "Declare input variables with defaults, lookups, and per-target overrides. Reference them with ${var.NAME}."
---

# Variables

ucm supports input variables so the same config can drive multiple
environments without duplicating resource definitions. Variables are declared
under the root `variables:` block and referenced with `${var.NAME}`.

## Syntax

```yaml
variables:
  region:
    description: Cloud region for storage roots.
    default: us-east-1

  metastore_id:
    description: Metastore UUID looked up at runtime.
    lookup:
      metastore: primary

resources:
  catalogs:
    sales:
      name: sales_${var.region}
      storage_root: s3://acme-${var.region}-sales
```

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | no | Currently only `complex` is meaningful; required for struct/array/map defaults. |
| `default` | any | no | Default value. When present the variable is optional. |
| `description` | string | no | Surfaced in `ucm schema` output. |
| `lookup` | object | no | Runtime lookup (e.g. `metastore: <name>` resolves to the metastore UUID). |

## Resolution priority

Highest to lowest:

1. Command-line flag: `--var="name=value"`.
2. Environment variable: `DATABRICKS_UCM_VAR_<NAME>`.
3. Per-target default: `targets.<name>.variables.<name>.default`.
4. Root default: `variables.<name>.default`.
5. `lookup` (resolved against the live workspace).
6. Error — the variable is required but unset.

## Setting via CLI

```bash
databricks ucm deploy --var="region=eu-west-1" --var="owner=ops@example.com"
databricks ucm plan   --var="region=eu-west-1"
```

`--var` is a persistent flag — it is accepted on every verb that reads
config.

## Per-target override

```yaml
variables:
  region:
    default: us-east-1

targets:
  prod:
    variables:
      region:
        default: us-west-2
```

## Lookups

Lookups resolve a known UC entity name into its ID at runtime. Supported
kinds: `metastore`.

```yaml
variables:
  metastore_id:
    lookup:
      metastore: primary
```

## See also

- [targets](targets.md) — per-environment overrides share the same
  `variables:` block shape.
- [cross-resource references](cross-resource-references.md) — the parallel
  `${resources.X.Y.Z}` substitution for ucm-managed resources.
