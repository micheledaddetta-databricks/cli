---
title: "ucm deployment bind"
description: "Bind a ucm-declared resource to an existing Unity Catalog object."
---

# ucm deployment bind

Records a state entry so subsequent deploys update — rather than recreate —
the existing UC object. `ucm.yml` is never modified.

## Syntax

```bash
databricks ucm deployment bind <KEY> <UC_NAME> [flags]
```

## Arguments

| Arg | Description |
|---|---|
| `<KEY>` | The resource key declared in `ucm.yml` (e.g. `team_alpha`). |
| `<UC_NAME>` | The name or full name of the existing UC object (e.g. `sales_prod` for a catalog, `team_alpha.bronze` for a schema, `team_alpha.bronze.landing` for a volume). |

## Flags

| Flag | Description |
|---|---|
| `--auto-approve` | Skip the interactive confirmation prompt. |
| `--target`, `-t` | Target to bind within. |

## Examples

```bash
# Bind a catalog declaration to an existing UC catalog
databricks ucm deployment bind team_alpha team_alpha

# Bind a schema (UC_NAME must be the schema's full name)
databricks ucm deployment bind bronze team_alpha.bronze

# Bind with automatic approval (CI/CD)
databricks ucm deployment bind my_vol team_alpha.bronze.landing --auto-approve
```

## Supported kinds

Catalogs, schemas, storage_credentials, external_locations, volumes,
connections. Grants are not bindable — they reconcile per securable.

## Engine support

Currently supported only under the [direct engine](../engines.md). Running
under the terraform engine returns an error.

## Warnings

After binding, the UC object is managed by ucm. Manual changes made outside
ucm may be overwritten on the next deploy.

## See also

- [deployment unbind](deployment-unbind.md)
- [import](import.md)
- [engines](../engines.md)
