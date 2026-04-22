---
title: "Cross-resource references"
description: "How ucm resolves literal string fields vs ucm-managed ${resources.X.Y.Z} references, and what differs per engine."
---

# Cross-resource references

Every string field in `ucm.yml` accepts two forms: a **literal** value or a
**ucm-managed reference** of the form `${resources.<kind>.<key>.<field>}`.

## Literal (bring-your-own)

The named object is expected to exist already. ucm references it read-only
and does not attempt to create or modify it.

```yaml
resources:
  catalogs:
    partner:
      name: partner_prod
      storage_root: preexisting_partner_cred   # literal credential name
```

## ucm-managed reference

The referenced object is declared in the same file. Resolution is
engine-specific but transparent to you.

```yaml
resources:
  storage_credentials:
    sales_cred:
      name: sales_cred
      aws_iam_role:
        role_arn: arn:aws:iam::111122223333:role/uc-sales
  catalogs:
    sales:
      name: sales_prod
      storage_root: ${resources.storage_credentials.sales_cred.name}
```

## Per-engine resolution

| Engine | How the ref resolves |
|---|---|
| terraform | Rewritten to `${databricks_storage_credential.sales_cred.name}` in the rendered `main.tf.json`. Terraform's own graph orders the dependency. |
| direct | Resolved at load time to the literal value (`sales_cred`). Apply order is declared explicitly: storage_credentials → catalogs → schemas → grants (reverse on delete). |

Unknown refs (typo'd kind or missing key) fail with a clear error at
`validate`/`plan` time.

## See also

- [engines](engines.md)
- [load-time mutators](load-time-mutators.md)
- [variables](variables.md) for the parallel `${var.NAME}` substitution.
