---
title: "ucm deployment unbind"
description: "Drop the recorded binding for a ucm-declared resource."
---

# ucm deployment unbind

Drops the recorded state entry so the next deploy treats the resource as
newly declared — creating it if absent, adopting if present (engine-dependent).

## Syntax

```bash
databricks ucm deployment unbind <KEY> [flags]
```

## Arguments

| Arg | Description |
|---|---|
| `<KEY>` | The resource key declared in `ucm.yml` to unbind. |

## Flags

| Flag | Description |
|---|---|
| `--auto-approve` | Skip the interactive confirmation prompt. |
| `--target`, `-t` | Target to unbind within. |

## Examples

```bash
databricks ucm deployment unbind team_alpha
databricks ucm deployment unbind bronze --auto-approve
```

## Engine support

Currently supported only under the [direct engine](../engines.md). Running
under the terraform engine returns an error.

## Re-binding

To re-bind later:

```bash
databricks ucm deployment bind <KEY> <UC_NAME>
```

## See also

- [deployment bind](deployment-bind.md)
- [engines](../engines.md)
