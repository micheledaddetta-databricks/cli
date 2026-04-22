---
title: "ucm plan"
description: "Preview the changes ucm deploy would make."
---

# ucm plan

Runs the initialize → build → terraform init → terraform plan sequence and
prints a DAB-style action list plus the add/change/delete/unchanged tally.
No state is mutated and no remote resources are touched.

## Syntax

```bash
databricks ucm plan [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--target`, `-t` | Target from `targets:` to plan against. |
| `-o`, `--output` | `text` (default) or `json`. JSON emits the structured plan. |
| `--var="name=value"` | Set a declared variable. Repeatable. |

## Examples

```bash
databricks ucm plan                     # Plan against the default target
databricks ucm plan --target prod       # Plan a specific target
databricks ucm plan -o json             # Emit the structured plan as JSON
```

## Output format

Text output matches `databricks bundle plan` byte-for-byte for the resources
ucm models. Example:

```
  + resources.catalogs.sales
  ~ resources.schemas.raw

Plan: 1 to add, 1 to change, 0 to delete, 3 unchanged
```

Action glyphs:

| Glyph | Action |
|---|---|
| `+` | Create |
| `~` | Update / UpdateWithID / Resize |
| `-` | Delete |
| `-/+` | Recreate |

JSON output emits the full `deployplan.Plan` structure — useful for CI
matrix generation or downstream tooling.

## See also

- [deploy](deploy.md)
- [diff](diff.md) — related: git-level diff, not engine plan.
- [engines](../engines.md)
