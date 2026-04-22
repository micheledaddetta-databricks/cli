---
title: "ucm drift"
description: "Detect out-of-band Unity Catalog changes by comparing persisted state with live UC reads."
---

# ucm drift

For every resource recorded in the direct-engine state file, `drift`
fetches the live UC object through the Databricks SDK and reports
per-field mismatches. Use this periodically or from CI to catch manual
UI/API edits that did not go through ucm.

## Syntax

```bash
databricks ucm drift [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--target`, `-t` | Target to check. |
| `-o`, `--output` | `text` (default) or `json`. |
| `--var="name=value"` | Set a declared variable. Repeatable. |

## Examples

```bash
databricks ucm drift                   # Check the default target
databricks ucm drift --target prod     # Check a specific target
databricks ucm drift -o json           # Emit structured JSON for tooling
```

## Exit codes

| Exit | Meaning |
|---|---|
| 0 | No drift detected. |
| 1 | At least one resource has drifted. |

## Engine support

Currently operates on direct-engine state only. Terraform-engine drift
requires parsing generic attribute maps from `tfstate` and is a follow-up
item. Live reads are routed through the SDK regardless of engine.

## See also

- [plan](plan.md)
- [deployment bind](deployment-bind.md) — rebind drifted resources after
  reconciling the source of truth.
- [engines](../engines.md)
