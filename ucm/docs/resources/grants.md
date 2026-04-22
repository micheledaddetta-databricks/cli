---
title: "grants"
description: "Assign Unity Catalog privileges on a securable to a principal."
---

# grants

Assigns UC privileges on a securable to a principal.

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `securable.type` | string | yes | `catalog`, `schema`, `storage_credential`, `external_location`, `volume`, or `connection`. |
| `securable.name` | string | yes | UC name of the target object. Can reference a ucm-managed sibling by its key; ucm wires the dependency. |
| `principal` | string | yes | User, group, or service principal name (account-level). ucm does not create principals — pass a name that already exists. |
| `privileges` | list[string] | yes | UC privilege names (e.g., `USE_CATALOG`, `USE_SCHEMA`, `SELECT`, `MODIFY`). |

## Example

```yaml
resources:
  grants:
    sales_readers:
      securable: { type: schema, name: raw }
      principal: sales-readers
      privileges: [USE_SCHEMA, SELECT]
```

Or nested under the schema:

```yaml
resources:
  catalogs:
    sales:
      name: sales_prod
      schemas:
        raw:
          name: raw
          grants:
            sales_readers:
              principal: sales-readers
              privileges: [USE_SCHEMA, SELECT]
```

## Engines

| Engine | Behavior |
|---|---|
| terraform | `databricks_grants.<key>`. The `catalog` / `schema` field takes `${databricks_catalog.<k>.name}` or `${databricks_schema.<k>.id}` when the securable is ucm-managed; otherwise the literal name. Emits `depends_on` for ucm-managed parents. |
| direct | `w.Grants.Update` per securable, reconciled in a single pass after all catalogs/schemas land. |

## Grant reconciliation (direct engine)

Grants are not stored per-entry in the remote — they're a set on the
securable. The direct engine reconciles by securable: it reads the current
grants, computes the diff against the declared set, and applies the
combined Add/Remove payload in one call. Removing a grant from `ucm.yml`
unassigns it on the next apply.

## See also

- [catalogs](catalogs.md)
- [schemas](schemas.md)
- [volumes](volumes.md)
- [external_locations](external-locations.md)
