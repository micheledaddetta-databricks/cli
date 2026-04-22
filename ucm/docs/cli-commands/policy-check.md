---
title: "ucm policy-check"
description: "Run only the ucm validation mutators (tags, naming, required fields). No network I/O."
---

# ucm policy-check

Runs the subset of ucm validation mutators that are cheap enough for a
pre-commit hook. Unlike [`validate`](validate.md), which runs the full
mutator chain, `policy-check` only runs the validation rules (tag
enforcement, naming, required fields). No network I/O.

## Syntax

```bash
databricks ucm policy-check [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--target`, `-t` | Target to check. |
| `--var="name=value"` | Set a declared variable. Repeatable. |

## Example

```bash
databricks ucm policy-check
```

## Pre-commit hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: ucm-policy-check
        name: ucm policy-check
        entry: databricks ucm policy-check
        language: system
        pass_filenames: false
        files: ^ucm\.yml$
```

## Notes

- Only runs validation mutators. Does not flatten nested resources, select
  a target, resolve refs, or talk to the workspace.
- `validate` remains the authoritative gate before `deploy`.

## See also

- [validate](validate.md)
- [load-time mutators](../load-time-mutators.md)
- [tag_validation_rules](../resources/tag-validation-rules.md)
