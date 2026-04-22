---
title: "ucm deploy"
description: "Apply ucm configuration to the target Databricks account/workspace."
---

# ucm deploy

Runs the full deploy sequence: initialize → build → terraform init →
terraform apply → state push. A failure mid-apply leaves the remote state
on the previous seq; re-running the command will re-attempt from a fresh
pull.

## Syntax

```bash
databricks ucm deploy [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--target`, `-t` | Target from `targets:` to deploy. |
| `--var="name=value"` | Set a declared variable. Repeatable. |

## Examples

```bash
databricks ucm deploy                   # Deploy the default target
databricks ucm deploy --target prod     # Deploy a specific target
databricks ucm deploy --var="region=eu-west-1"
```

## Auth

Requires an authenticated workspace client. Standard Databricks OAuth M2M
variables are honored:

```bash
export DATABRICKS_HOST=https://workspace.cloud.databricks.com
export DATABRICKS_CLIENT_ID=<sp-client-id>
export DATABRICKS_CLIENT_SECRET=<sp-secret>
databricks ucm deploy
```

## Pre-apply validation

`deploy` runs the full mutator chain upstream of apply — there is no need
to run `ucm validate` separately, though doing so in CI is good practice
because it fails faster.

## See also

- [plan](plan.md)
- [destroy](destroy.md)
- [validate](validate.md)
- [engines](../engines.md)
