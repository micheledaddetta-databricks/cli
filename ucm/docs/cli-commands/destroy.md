---
title: "ucm destroy"
description: "Tear down everything managed by the current target."
---

# ucm destroy

Runs the initialize → terraform init → terraform destroy → state push
sequence against the selected target. Operates on the already-rendered
terraform config cached from the last apply.

## Syntax

```bash
databricks ucm destroy [flags]
```

## Flags

| Flag | Description |
|---|---|
| `--auto-approve` | Skip interactive approvals for deleting resources. Required on terminals that do not support prompting. |
| `--target`, `-t` | Target from `targets:` to destroy. |
| `--var="name=value"` | Set a declared variable. Repeatable. |

## Examples

```bash
databricks ucm destroy --auto-approve                # Destroy default target
databricks ucm destroy --target dev --auto-approve   # Destroy a specific target
```

## Warnings

`destroy` is a one-way operation. It does *not* check for consumers — if
other principals hold references to a catalog or schema you destroy,
their queries will break immediately on the next UC metadata refresh.

For reversible teardowns, prefer using separate targets (e.g. a short-lived
`sandbox` target) rather than destroying production.

## See also

- [deploy](deploy.md)
- [targets](../targets.md)
