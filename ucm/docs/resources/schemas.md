---
title: "schemas"
description: "Declare UC schemas inside a catalog, with tag inheritance and nested grants."
---

# schemas

A UC schema (database) inside a catalog.

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Schema name within its catalog. |
| `catalog` | string | yes (flat form) | Parent catalog name. Injected automatically when declared nested under a catalog. |
| `comment` | string | no | Human-readable description. |
| `tags` | map[string]string | no | Validated by `tag_validation_rules`. Merged with parent-catalog tags unless `tag_inherit: false` (schema keys win on conflict). |
| `tag_inherit` | bool pointer | no | Default `true` (nil). Set to `false` to opt out of catalog-tag merging. |
| `grants` | map[string]*Grant | no | Nested form; flattened with `securable = {type: schema, name: <key>}`. |

## Example

```yaml
resources:
  schemas:
    raw:
      name: raw
      catalog: sales
      comment: landing zone
      tag_inherit: false            # don't pull in sales.tags
      tags:
        cost_center: "1234"
        data_owner: sales
        classification: internal
```

## Engines

| Engine | Behavior |
|---|---|
| terraform | `databricks_schema.<key>` with `catalog_name`, `force_destroy=true`, `properties=<tags>`, and `depends_on: [databricks_catalog.<catalog>]` when the parent is ucm-managed. |
| direct | `w.Schemas.Create` / `.Update` / `.Delete`, applied after the parent catalog. |

## See also

- [catalogs](catalogs.md)
- [grants](grants.md)
- [tag_validation_rules](tag-validation-rules.md)
- [../load-time-mutators.md](../load-time-mutators.md) — how `InheritCatalogTags` works.
