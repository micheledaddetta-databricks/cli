---
title: "Includes"
description: "Split a ucm project across multiple files using glob patterns under the root include: key."
---

# Includes

For larger projects you can split `ucm.yml` across multiple files and glue
them together with the root `include:` key.

## Syntax

```yaml
# ucm.yml
ucm:
  name: my-deployment

include:
  - resources/catalogs/*.yml
  - resources/schemas/*.yml
  - shared/*.yml
```

Each glob is resolved relative to the root `ucm.yml`. Matched files are
loaded and deep-merged into the root tree in the order they are discovered.

## Rules

- Only the root `ucm.yml` can declare `include:`. Included files cannot
  themselves include further files.
- Glob metacharacters (`*`, `?`, `[`, `]`, `^`) are valid in the pattern but
  not in the path to the root `ucm.yml` itself.
- A literal (non-glob) entry that matches no file is an error. A glob with no
  match only logs a warning — mirroring DAB behavior.
- `include:` entries must be relative paths. Absolute paths are rejected.

## Example: resource-per-file

```
my-deployment/
├── ucm.yml
├── resources/
│   ├── catalogs/
│   │   ├── sales.yml
│   │   └── marketing.yml
│   └── schemas/
│       ├── sales.yml
│       └── marketing.yml
└── shared/
    └── grants.yml
```

```yaml
# ucm.yml
ucm: { name: my-deployment }

workspace:
  host: https://workspace.cloud.databricks.com

include:
  - resources/catalogs/*.yml
  - resources/schemas/*.yml
  - shared/grants.yml
```

```yaml
# resources/catalogs/sales.yml
resources:
  catalogs:
    sales:
      name: sales_prod
      comment: sales domain catalog
```

## See also

- [ucm.yml settings](settings.md) — the root schema that `include:` fans out.
- [targets](targets.md) — targets are declared alongside resources and can
  also be split via `include:`.
