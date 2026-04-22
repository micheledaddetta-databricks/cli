---
title: "ucm diff"
description: "Detect which ucm resources changed between two git refs. Intended for CI matrix generation."
---

# ucm diff

Compares `ucm.yml` at `--base` (default `origin/main`) with `ucm.yml` at
`--head` (default `HEAD`, i.e. the working tree if clean) and prints the
set of resource keys that were added, removed, or modified. Intended for
CI matrix generation — pair with `-o json` and feed into a matrix strategy.

## Syntax

```bash
databricks ucm diff [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--base` | Git ref to diff against. Default: `origin/main`. |
| `--head` | Git ref (or `HEAD`) to diff from. `HEAD` reads the working tree. Default: `HEAD`. |
| `-o`, `--output` | `text` (default) or `json`. |

## Examples

```bash
databricks ucm diff                        # HEAD vs origin/main
databricks ucm diff --base main            # vs local main
databricks ucm diff --base v1.2.3 -o json  # JSON for CI
```

## Output

Text form lists changed resource keys, one per line. JSON form emits:

```json
{
  "changed_resources": [
    "resources.catalogs.sales",
    "resources.schemas.sales.raw"
  ]
}
```

## CI usage

```yaml
# GitHub Actions example
- id: diff
  run: databricks ucm diff --base origin/main -o json > diff.json
- id: matrix
  run: jq '.changed_resources' diff.json
```

## Note

Unlike [`plan`](plan.md) and [`drift`](drift.md), `diff` does not contact
the workspace. It is a pure git-level diff of `ucm.yml`.

## See also

- [plan](plan.md) — engine-level plan against live state.
- [drift](drift.md) — live UC vs persisted state.
