---
title: "external_locations"
description: "Grant Unity Catalog access to a specific cloud-storage prefix via URL + storage credential."
---

# external_locations

A UC external location. A `url` + `credential_name` pair together grant UC
access to a specific cloud-storage prefix.

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | External location name in UC. |
| `url` | string | yes | Cloud storage URL (e.g. `s3://bucket/path`, `abfss://...`, `gs://...`). |
| `credential_name` | string | yes | Name of the storage credential UC should use. Literal, or `${resources.storage_credentials.<key>.name}`. |
| `comment` | string | no | Human-readable description. |
| `read_only` | bool | no | Location is usable only for read operations. Default `false`. |
| `skip_validation` | bool | no | Skip server-side validation on create. Default `false`. |
| `fallback` | bool | no | Whether the location is a fallback for nested-path access. Default `false`. |

## Example

```yaml
resources:
  storage_credentials:
    sales_cred:
      name: sales_cred
      aws_iam_role:
        role_arn: arn:aws:iam::111122223333:role/uc-sales

  external_locations:
    sales_raw:
      name: sales_raw
      url: s3://acme-sales/raw
      credential_name: ${resources.storage_credentials.sales_cred.name}
      comment: sales raw-zone landing
```

## Engines

| Engine | Behavior |
|---|---|
| terraform | `databricks_external_location.<key>`. Emitted after storage_credentials and before catalogs. |
| direct | `w.ExternalLocations.Create` / `.Update` / `.Delete`. Created after its credential; deleted before it. |

## See also

- [storage_credentials](storage-credentials.md)
- [volumes](volumes.md) — external volumes reference an external location's prefix.
- [catalogs](catalogs.md) — `storage_root` can reference `${resources.external_locations.*.url}`.
