---
title: "tag_validation_rules"
description: "Declarative ucm-native tag policy. Enforces required tag keys and allowed values on catalogs and schemas."
---

# tag_validation_rules

Declarative tag policy. Independent of any server-side UC tag feature —
this is ucm's own gate, enforced by the `ValidateTags` mutator on
`validate`, `plan`, and `policy-check`.

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `securable_types` | list[string] | yes | Which resource kinds this rule applies to. Currently supported: `catalog`, `schema`. |
| `required` | list[string] | no | Tag keys that must be present on every matching securable. |
| `allowed_values` | map[string]list[string] | no | Restricts values for named keys. Keys not listed accept any value. |

## Example

```yaml
resources:
  tag_validation_rules:
    enforce_ownership:
      securable_types: [catalog, schema]
      required:
        - cost_center
        - data_owner
        - classification
      allowed_values:
        classification: [public, internal, confidential, restricted]
```

## When it runs

The `ValidateTags` mutator runs on `validate`, `plan`, and `policy-check`.
Violations produce error-level diagnostics pointing at the offending
securable's YAML span. `deploy` inherits the same check because it runs
`validate` upstream of apply.

## Tag inheritance

Catalog tags are merged into child schemas before validation runs. See
[schemas](schemas.md) and [load-time mutators](../load-time-mutators.md)
for the exact order.

## See also

- [catalogs](catalogs.md)
- [schemas](schemas.md)
- [cli-commands/validate](../cli-commands/validate.md)
- [cli-commands/policy-check](../cli-commands/policy-check.md)
