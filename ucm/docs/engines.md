---
title: "Engines"
description: "Choose between the terraform engine (DAB-style plan/apply) and the direct engine (SDK calls, no terraform)."
---

# Engines

The same `ucm.yml` ships through two interchangeable engines. Pick either
per project; switch by setting `ucm.engine` in config, by exporting
`DATABRICKS_UCM_ENGINE`, or by leaving the default in place.

## Comparison

| | **terraform** (default) | **direct** |
|---|---|---|
| How | Renders a `main.tf.json` and drives `terraform init` + `terraform plan`/`apply`. | Issues SDK calls directly (`w.Catalogs.Create(...)`, etc.). |
| State | `terraform.tfstate` in the configured backend. | `resources.json` per target. |
| Plan diff | Full terraform plan. | DAB-style action list from a local diff. |
| External deps | Terraform binary (resolved automatically) + databricks provider. | None. |
| When to pick | Matches DAB workflows, richer plan diff, familiar tooling. | No terraform binary, fast cold start, fewer moving parts. |

## Selecting an engine

In priority order (highest wins):

1. `DATABRICKS_UCM_ENGINE=direct` environment variable.
2. `ucm.engine: direct` in `ucm.yml`.
3. Default: `terraform`.

```yaml
ucm:
  name: my-deployment
  engine: direct
```

## Engine-specific behavior

A handful of features differ between engines. Each is called out on the
relevant reference page:

- [Cross-resource references](cross-resource-references.md) — resolution
  timing differs.
- [`deployment bind`/`unbind`](cli-commands/deployment-bind.md) — currently
  supported only under the direct engine.
- [`drift`](cli-commands/drift.md) — operates on direct-engine state
  today; terraform support is a follow-up.

## See also

- [cross-resource references](cross-resource-references.md)
- [load-time mutators](load-time-mutators.md)
