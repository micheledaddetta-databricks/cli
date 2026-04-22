---
title: "ucm debug"
description: "Dump internal ucm state (terraform version, state paths) for troubleshooting."
---

# ucm debug

Exposes internal ucm state for troubleshooting and for tooling integrations
(e.g. the Databricks VSCode extension). The group is hidden from the top
level `ucm --help` output; it is a surface for tooling, not an end-user
workflow.

## Subcommands

### ucm debug terraform

Prints the terraform binary version and the Databricks provider version ucm
pins, along with URLs for air-gap downloads.

```bash
databricks ucm debug terraform
databricks ucm debug terraform -o json
```

The `-o json` form emits:

```json
{
  "terraform": {
    "version": "1.5.5",
    "providerHost": "registry.terraform.io",
    "providerSource": "databricks/databricks",
    "providerVersion": "1.x.y"
  }
}
```

### ucm debug states

Lists the state files ucm maintains for the selected target so users and
tooling can see which engine's state the target is currently bound to.

```bash
databricks ucm debug states
databricks ucm debug states --target prod
databricks ucm debug states --force-pull       # bypass local cache
```

Checked locations under `.databricks/ucm/<target>/`:

- `ucm.json` (overall state marker)
- `terraform/terraform.tfstate` (terraform engine)
- `resources.json` (direct engine)

## See also

- [summary](summary.md)
- [engines](../engines.md)
