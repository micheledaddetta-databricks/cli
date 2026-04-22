---
title: "ucm init"
description: "Scaffold a new ucm.yml project from a starter template."
---

# ucm init

Scaffold a new `ucm.yml` project from a starter template. Templates can be
built-in, a local directory, or a remote Git repository.

## Syntax

```bash
databricks ucm init [TEMPLATE_PATH] [flags]
```

`TEMPLATE_PATH` is optional. It can be:

- A built-in template name (e.g. `default`, `brownfield`, `multienv`).
- A local filesystem path to a ucm template directory.
- A Git repository URL (`https://` or `git@`).

When omitted, `init` interactively prompts from the list of built-in
templates.

## Flags

| Flag | Description |
|---|---|
| `--config-file` | JSON file containing key/value pairs of input parameters required for template initialization. |
| `--template-dir` | Subdirectory path within a Git repository containing the template. |
| `--output-dir` | Directory to write the initialized template to. Default: current directory. |
| `--branch` | Git branch to use for template initialization. Mutually exclusive with `--tag`. |
| `--tag` | Git tag to use for template initialization. Mutually exclusive with `--branch`. |

## Examples

```bash
databricks ucm init                       # Interactive built-in picker
databricks ucm init default               # Minimal catalog + schema + grant
databricks ucm init brownfield            # Stub for 'ucm generate' follow-up
databricks ucm init multienv              # dev/staging/prod targets
databricks ucm init ./my-template         # Initialize from a local directory
databricks ucm init --output-dir ./new default
databricks ucm init https://github.com/acme/ucm-templates --template-dir prod
```

## Auth

`init` requires an authenticated workspace client even when the template
does not call workspace helpers — the template renderer exposes helpers
(`workspace_host`, `short_name`, ...) that depend on it.

## See also

- [generate](generate.md) — for existing UC deployments.
- [ucm.yml settings](../settings.md)
