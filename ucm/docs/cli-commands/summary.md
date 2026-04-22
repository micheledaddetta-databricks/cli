---
title: "ucm summary"
description: "Summarize deployed resources by type."
---

# ucm summary

Reads the local terraform state cached under `.databricks/ucm/<target>/`
and prints a table of resource type + count. Run `ucm deploy` (or at least
`ucm plan`) first; without a local state the table is empty.

## Syntax

```bash
databricks ucm summary [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--target`, `-t` | Target to summarize. |
| `--var="name=value"` | Set a declared variable (passed through for config load). |

## Example

```bash
$ databricks ucm summary --target prod
TYPE                              COUNT
databricks_catalog                4
databricks_external_location      2
databricks_grants                 12
databricks_schema                 18
databricks_storage_credential     2
databricks_volume                 3
```

## Notes

- Reads local state only; does not contact the remote workspace.
- If no state exists, prints `No deployed resources found. Run 'ucm deploy' first.`
- Resource `TYPE` values are the terraform type names, not ucm kind names.

## See also

- [deploy](deploy.md)
- [debug](debug.md) — lists the paths of all state files ucm maintains.
