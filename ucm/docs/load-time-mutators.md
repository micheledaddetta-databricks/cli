---
title: "Load-time mutators"
description: "The automatic transformations applied to ucm.yml before any engine runs: flatten, inherit, resolve, validate."
---

# Load-time mutators

ucm runs a fixed pipeline of mutators against the loaded config before any
engine sees the tree. You do not invoke these directly, but knowing the
order and what each does explains several downstream rules.

## Pipeline order

1. **ProcessRootIncludes** — expands each glob in `include:` and merges
   matched files into the root tree. See [includes](includes.md).
2. **InitializeVariables** — normalizes `variables:` declarations and
   applies `--var`, env, and per-target overrides.
3. **FlattenNestedResources** — lifts nested schemas and grants out of
   catalogs (and grants out of schemas), injecting parent references.
   After this step, every resource lives in a top-level flat map.
4. **InheritCatalogTags** — merges a catalog's `tags` into every child
   schema unless the schema sets `tag_inherit: false`. Schema keys win on
   collisions.
5. **DefineDefaultTarget** / **SelectDefaultTarget** / **SelectTarget** —
   picks the active target (from `--target` or `default: true`) and folds
   its overrides into the top-level tree.
6. **ResolveVariableReferences** — substitutes `${var.NAME}` and
   `${resources.X.Y.Z}` references. See [variables](variables.md) and
   [cross-resource references](cross-resource-references.md).
7. **ValidateTags** — runs on `validate`/`plan`/`policy-check` only.
   Enforces the `tag_validation_rules:*` declarations against every
   matching securable.
8. **ResolveResourceReferences** (direct engine only) — rewrites
   `${resources.*}` refs to literal strings for SDK calls. The terraform
   engine preserves the refs and runs its own `Interpolate` pass later,
   rewriting to `${databricks_*}` form.

## Observability

Run `ucm debug` to dump the post-mutator tree for troubleshooting. See
[debug](cli-commands/debug.md).

## See also

- [resources/tag-validation-rules](resources/tag-validation-rules.md)
- [cli-commands/validate](cli-commands/validate.md)
- [cli-commands/policy-check](cli-commands/policy-check.md)
