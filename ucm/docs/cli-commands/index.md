---
title: "CLI commands"
description: "Reference for every databricks ucm verb: validate, plan, deploy, destroy, summary, init, generate, import, deployment bind/unbind, debug, diff, drift, schema, policy-check."
---

# CLI commands

Reference for every `databricks ucm` verb.

## Core workflow

| Verb | Purpose |
|---|---|
| [validate](validate.md) | Lint `ucm.yml`; run the full mutator chain and tag rules. |
| [plan](plan.md) | Preview the changes `deploy` would make. |
| [deploy](deploy.md) | Apply the configuration to the target workspace. |
| [destroy](destroy.md) | Tear down everything managed by the target. |
| [summary](summary.md) | Summarize resources currently tracked by state. |

## Authoring / onboarding

| Verb | Purpose |
|---|---|
| [init](init.md) | Scaffold a new `ucm.yml` project from a starter template. |
| [generate](generate.md) | Scan an existing account + metastore + workspace and emit `ucm.yml` + seed state. |
| [import](import.md) | Bind state for a single existing UC resource to a `ucm.yml` declaration. |
| [deployment bind](deployment-bind.md) | Attach an existing Databricks resource to a ucm key. |
| [deployment unbind](deployment-unbind.md) | Drop the recorded binding for a ucm key. |

## Governance / troubleshooting

| Verb | Purpose |
|---|---|
| [drift](drift.md) | Compare live UC state against persisted state; alert on out-of-band changes. |
| [diff](diff.md) | Detect which ucm stacks changed since a base git ref. |
| [policy-check](policy-check.md) | Run only the validation mutators (cheap pre-commit hook). |
| [debug](debug.md) | Dump internal ucm state (terraform version, states) for troubleshooting. |
| [schema](schema.md) | Print the JSON schema for `ucm.yml`. |

## Global flags

Inherited by every verb:

| Flag | Description |
|---|---|
| `--target` / `-t` | Target name from `targets:` to merge. Default: the `default: true` target, or an empty target named `default`. |
| `--var="name=value"` | Set the value of a declared variable. Repeatable. See [variables](../variables.md). |

## Environment

| Variable | Purpose |
|---|---|
| `DATABRICKS_UCM_ENGINE` | `terraform` (default) or `direct`. Overrides `ucm.engine`. |
| `DATABRICKS_UCM_VAR_<NAME>` | Set a variable value from the environment. |
| `DATABRICKS_HOST` / `DATABRICKS_CLIENT_ID` / `DATABRICKS_CLIENT_SECRET` | Standard Databricks OAuth M2M auth. |

## See also

- [ucm.yml settings](../settings.md)
- [engines](../engines.md)
