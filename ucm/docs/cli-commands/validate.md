---
title: "ucm validate"
description: "Validate ucm.yml for errors, warnings, and policy violations."
---

# ucm validate

Runs the full ucm mutator chain (includes, variables, flatten, inherit,
select target, resolve refs, tag validation) against the selected target.
Useful as a CI gate before `ucm deploy`.

## Syntax

```bash
databricks ucm validate [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--strict` | Treat warnings as errors. Exits non-zero if any warning is emitted. |
| `--target`, `-t` | Target from `targets:` to validate. |
| `--var="name=value"` | Set a declared variable. Repeatable. |

## Examples

```bash
databricks ucm validate                   # Validate default target
databricks ucm validate --target prod     # Validate a specific target
databricks ucm validate --strict          # Fail on warnings too
```

## Exit status

| Exit | Meaning |
|---|---|
| 0 | No errors. (With `--strict`, also no warnings.) |
| 1 | At least one error. Or, with `--strict`, at least one warning. |

## See also

- [policy-check](policy-check.md) — runs only the validation mutators (cheaper).
- [load-time mutators](../load-time-mutators.md) — what validate executes.
- [tag_validation_rules](../resources/tag-validation-rules.md) — ucm's tag policy gate.
