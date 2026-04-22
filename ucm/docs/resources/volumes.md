---
title: "volumes"
description: "Declare managed or external Unity Catalog volumes."
---

# volumes

A UC volume, either `MANAGED` (UC provisions the underlying storage) or
`EXTERNAL` (points at a cloud path under an external_location).

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Volume name. |
| `catalog_name` | string | yes | Parent catalog. Literal, or `${resources.catalogs.<key>.name}`. |
| `schema_name` | string | yes | Parent schema. Literal, or `${resources.schemas.<key>.name}`. |
| `volume_type` | string | yes | `MANAGED` or `EXTERNAL`. |
| `storage_location` | string | required for EXTERNAL | Cloud URL under a registered external_location. Unset for MANAGED. |
| `comment` | string | no | Human-readable description. |

## Example (managed)

```yaml
resources:
  volumes:
    docs:
      name: docs
      catalog_name: ${resources.catalogs.sales.name}
      schema_name: ${resources.schemas.raw.name}
      volume_type: MANAGED
      comment: sales documentation attachments
```

## Example (external)

```yaml
resources:
  external_locations:
    sales_raw:
      name: sales_raw
      url: s3://acme-sales/raw
      credential_name: ${resources.storage_credentials.sales_cred.name}

  volumes:
    docs_ext:
      name: docs_ext
      catalog_name: sales_prod
      schema_name: raw
      volume_type: EXTERNAL
      storage_location: s3://acme-sales/raw/docs
```

## Engines

| Engine | Behavior |
|---|---|
| terraform | `databricks_volume.<key>` with `depends_on` on the parent schema (and, for EXTERNAL, the external_location). |
| direct | `w.Volumes.Create` / `.Update` / `.Delete`, applied after catalogs and schemas. |

## See also

- [catalogs](catalogs.md)
- [schemas](schemas.md)
- [external_locations](external-locations.md)
- [grants](grants.md) — volumes can be granted with `securable.type: volume`.
