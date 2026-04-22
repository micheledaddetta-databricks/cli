---
title: "Resources reference"
description: "Every resource kind you can declare under resources: in ucm.yml."
---

# Resources reference

Every kind you can declare under `resources:` in `ucm.yml`. Each page
documents fields, a minimal example, and engine-specific behavior.

## Kinds

| Kind | Scope | Purpose |
|---|---|---|
| [catalogs](catalogs.md) | workspace | UC catalogs. |
| [schemas](schemas.md) | workspace | UC schemas inside a catalog. |
| [grants](grants.md) | workspace | UC privileges on a securable. |
| [storage_credentials](storage-credentials.md) | workspace | Cloud storage auth (AWS/Azure/GCP). |
| [external_locations](external-locations.md) | workspace | URL + credential pair granting UC access to a storage prefix. |
| [volumes](volumes.md) | workspace | Managed or external UC volumes. |
| [connections](connections.md) | workspace | Foreign-catalog federation connections (MySQL, Snowflake, ...). |
| [tag_validation_rules](tag-validation-rules.md) | policy | ucm-native tag-policy rules. No server dependency. |

## See also

- [ucm.yml settings](../settings.md)
- [cross-resource references](../cross-resource-references.md)
- [load-time mutators](../load-time-mutators.md)
