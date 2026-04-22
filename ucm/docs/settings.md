---
title: "ucm.yml settings"
description: "Root schema for a ucm.yml project: ucm, workspace, account, resources, targets, variables, include."
---

# ucm.yml settings

Every ucm project starts with a `ucm.yml` that follows the same top-level
shape. This page documents each root key.

## Skeleton

```yaml
ucm:
  name: my-deployment           # required; uniquely identifies the deployment
  engine: direct                # optional; "terraform" (default) or "direct".

workspace:
  host: https://<workspace>.cloud.databricks.com
  profile: my-profile

account:                        # optional; only for account-scoped verbs
  account_id: <uuid>
  host: https://accounts.cloud.databricks.com

variables:                      # optional; see variables.md
  region:
    default: us-east-1

include:                        # optional; see includes.md
  - resources/*.yml

resources:
  catalogs: { ... }
  schemas: { ... }
  grants: { ... }
  storage_credentials: { ... }
  external_locations: { ... }
  volumes: { ... }
  connections: { ... }
  tag_validation_rules: { ... }

targets:                        # optional; see targets.md
  dev:
    default: true
    workspace: { host: https://dev... }
    resources: { ... }
  prod:
    workspace: { host: https://prod... }
    resources: { ... }
```

## `ucm`

Top-level metadata for the deployment.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Uniquely identifies this deployment. Used to scope state. |
| `engine` | string | no | `terraform` (default) or `direct`. Overridable with `DATABRICKS_UCM_ENGINE`. See [engines](engines.md). |
| `target` | string | no | Read-only. Populated by `SelectTarget`; do not set manually. |

## `workspace`

The workspace this deployment targets. Required for any workspace-scoped
resource (catalogs, schemas, grants, volumes, ...).

| Field | Type | Required | Description |
|---|---|---|---|
| `host` | string | yes for workspace ops | Workspace URL. |
| `profile` | string | no | Profile in `~/.databrickscfg` to resolve auth from. |

## `account`

The Databricks account hosting the metastore. Only needed for account-scoped
verbs — today most ucm flows read it read-only.

| Field | Type | Required | Description |
|---|---|---|---|
| `account_id` | string | no | Account UUID. |
| `host` | string | no | Usually `https://accounts.cloud.databricks.com`. |

## `resources`

Map of every declared resource, keyed by kind. See the [resources
reference](resources/index.md) for the full list and per-kind field tables.

## `targets`

Per-environment overrides. See [targets](targets.md).

## `variables`

Inputs declared at the root and consumed with `${var.NAME}`. See
[variables](variables.md).

## `include`

Glob patterns of additional files to merge into the root. See
[includes](includes.md).

## See also

- [engines](engines.md)
- [cross-resource references](cross-resource-references.md)
- [variables](variables.md)
- [targets](targets.md)
- [resources reference](resources/index.md)
