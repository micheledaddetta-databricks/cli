---
title: "Targets"
description: "Declare per-environment overrides (dev/staging/prod) and select them with --target."
---

# Targets

Targets are per-environment overrides of the root tree. A ucm deployment can
declare any number of targets; one is selected at verb time and its
overrides are merged into the root before any engine runs.

## Syntax

```yaml
targets:
  dev:
    default: true                 # selected when --target is omitted
    workspace:
      host: https://dev.cloud.databricks.com
    resources:
      catalogs:
        sales:
          name: sales_dev         # override catalog name

  prod:
    workspace:
      host: https://prod.cloud.databricks.com
    resources:
      catalogs:
        sales:
          name: sales_prod
```

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `default` | bool | no | When `true`, the target is selected when `--target` is omitted. |
| `workspace` | object | no | Override the root `workspace:` block. |
| `account` | object | no | Override the root `account:` block. |
| `resources` | object | no | Merged with the root `resources:` tree. |
| `variables` | object | no | Override variable defaults. See [variables](variables.md). |

## Default target

If no target is marked `default: true` and `--target` is omitted, the
`DefineDefaultTarget` mutator adds an empty target named `default` so that
verb wiring is uniform.

## Selecting at runtime

```bash
databricks ucm plan                     # use default target
databricks ucm plan --target prod       # use prod target
databricks ucm deploy -t prod           # short form
```

The `--target` (alias `-t`) flag is inherited by every verb.

## Merge semantics

For each of `workspace`, `account`, and `resources`, the target value is
deep-merged with the root value. Scalar overrides win; map entries are
unioned; list entries are replaced (not appended).

## See also

- [ucm.yml settings](settings.md)
- [load-time mutators](load-time-mutators.md) — the exact merge order.
- [variables](variables.md)
