---
title: "Unity Catalog Management (ucm)"
description: "Declarative management of Unity Catalog resources (catalogs, schemas, grants, credentials, locations, volumes, connections) at enterprise scale."
---

# Unity Catalog Management (ucm)

`databricks ucm` brings DAB-style declarative configuration to Unity Catalog.
You declare catalogs, schemas, grants, storage credentials, external
locations, volumes, and connections in a single `ucm.yml` file and deploy
them through either a terraform or direct engine.

## When to use ucm

- You manage UC metastores at enterprise scale and want a single source of
  truth checked into git.
- You want bundle-style targets, variables, and includes so the same config
  can drive dev, staging, and prod.
- You need tag policy enforcement, drift detection, and plan/deploy parity
  with [Declarative Automation Bundles (DABs)](https://docs.databricks.com/en/dev-tools/bundles/index.html).

## Quickstart

```bash
databricks ucm init default             # Scaffold a ucm.yml project
databricks ucm validate                 # Lint config + run tag rules
databricks ucm plan --target dev        # Preview changes
databricks ucm deploy --target prod     # Apply changes
```

## Documentation

- [ucm.yml settings](settings.md) — root schema: `ucm`, `workspace`, `account`, `resources`, `targets`, `variables`, `include`.
- [Engines](engines.md) — terraform vs direct deployment engines.
- [Cross-resource references](cross-resource-references.md) — literal vs `${resources.X.Y.Z}` refs.
- [Variables](variables.md) — `variables:` block, `--var`, `${var.x}`.
- [Includes](includes.md) — split a project across files with globs.
- [Targets](targets.md) — per-environment overrides.
- [Load-time mutators](load-time-mutators.md) — flatten, inherit, resolve, validate pipeline.
- [Resources reference](resources/index.md) — one page per resource kind.
- [CLI commands](cli-commands/index.md) — one page per verb.

## See also

- [Declarative Automation Bundles](https://docs.databricks.com/en/dev-tools/bundles/index.html) — the sibling DAB tool that ucm mirrors.
- [Unity Catalog overview](https://docs.databricks.com/en/data-governance/unity-catalog/index.html).
