---
title: "catalogs"
description: "Declare Unity Catalog catalogs with storage roots, tags, and nested schemas/grants."
---

# catalogs

A Unity Catalog catalog.

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Catalog name in UC. |
| `comment` | string | no | Human-readable description. |
| `storage_root` | string | no | Cloud URL or `${resources.storage_credentials.*.name}` / `${resources.external_locations.*.url}`. Required for non-metastore-default catalogs. |
| `tags` | map[string]string | no | Validated by `tag_validation_rules`. Inherited by child schemas unless `tag_inherit: false`. |
| `schemas` | map[string]*Schema | no | Nested form; flattened to the top-level `schemas` map at load time. |
| `grants` | map[string]*Grant | no | Nested form; flattened to the top-level `grants` map at load time. |

## Example

```yaml
resources:
  catalogs:
    sales:
      name: sales_prod
      comment: sales domain catalog
      storage_root: ${resources.storage_credentials.sales_cred.name}
      tags:
        cost_center: "1234"
        data_owner: sales
        classification: internal
```

## Engines

| Engine | Behavior |
|---|---|
| terraform | `databricks_catalog.<key>` with `force_destroy=true` and `properties=<tags>`. |
| direct | `w.Catalogs.Create` / `.Update` / `.Delete`. |

## Nested form

Schemas and grants can be declared under a catalog. `FlattenNestedResources`
unrolls them into the top-level maps at load time, injecting the parent
reference. These two forms are equivalent:

```yaml
# Nested
resources:
  catalogs:
    sales:
      name: sales_prod
      schemas:
        raw: { name: raw }
      grants:
        admins:
          principal: sales-admins
          privileges: [USE_CATALOG]
```

```yaml
# Flat
resources:
  catalogs:
    sales: { name: sales_prod }
  schemas:
    raw: { name: raw, catalog: sales }
  grants:
    admins:
      securable: { type: catalog, name: sales }
      principal: sales-admins
      privileges: [USE_CATALOG]
```

Collisions (same key declared both flat and nested) are a hard error.

## See also

- [schemas](schemas.md)
- [grants](grants.md)
- [storage_credentials](storage-credentials.md)
- [tag_validation_rules](tag-validation-rules.md)
