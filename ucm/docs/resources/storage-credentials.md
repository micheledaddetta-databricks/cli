---
title: "storage_credentials"
description: "Declare UC storage credentials for AWS, Azure, or Databricks-managed GCP identities."
---

# storage_credentials

A UC storage credential — the capability UC uses to authenticate to cloud
storage. Exactly one identity shape (AWS / Azure MI / Azure SP / Databricks
GCP SA) must be set.

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Credential name in UC. |
| `comment` | string | no | Human-readable description. |
| `aws_iam_role` | object | one-of | `{ role_arn: arn:aws:iam::...:role/... }` |
| `azure_managed_identity` | object | one-of | `{ access_connector_id: <ARM id>, managed_identity_id: <ARM id, optional> }` |
| `azure_service_principal` | object | one-of | `{ directory_id, application_id, client_secret }` |
| `databricks_gcp_service_account` | object | one-of | `{}` (empty; presence alone toggles the shape — the GCP SA is managed by Databricks) |
| `read_only` | bool | no | Credential is usable only for read operations. Default `false`. |
| `skip_validation` | bool | no | Skip server-side validation on create. Default `false`. Use sparingly — trades fast-fail for runtime surprises. |

Exactly-one-of on the identity fields is enforced by both the terraform
converter and the direct-engine input builder. Missing or multiple identity
fields fail before any API call fires.

## Example (AWS)

```yaml
resources:
  storage_credentials:
    sales_cred:
      name: sales_cred
      comment: sales domain storage access
      aws_iam_role:
        role_arn: arn:aws:iam::111122223333:role/uc-sales
```

## Example (Azure MI)

```yaml
resources:
  storage_credentials:
    shared_cred:
      name: shared_cred
      azure_managed_identity:
        access_connector_id: /subscriptions/s/resourceGroups/rg/providers/Microsoft.Databricks/accessConnectors/uc
```

## Example (Databricks-managed GCP SA)

```yaml
resources:
  storage_credentials:
    data_cred:
      name: data_cred
      databricks_gcp_service_account: {}
```

## Engines

| Engine | Behavior |
|---|---|
| terraform | `databricks_storage_credential.<key>`. Emitted ahead of catalogs in the rendered tree. |
| direct | `w.StorageCredentials.Create` / `.Update` / `.Delete`. Created before any catalog that references it; deleted after any catalog that references it is torn down. |

## Secret round-trip

`ucm generate` cannot recover the `client_secret` of an
`azure_service_principal` — UC does not echo it back on read. The generated
YAML leaves a placeholder with a warning; fill the secret in before the next
deploy.

## See also

- [external_locations](external-locations.md)
- [catalogs](catalogs.md) — uses `storage_root: ${resources.storage_credentials.*.name}`.
