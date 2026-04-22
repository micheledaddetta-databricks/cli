---
title: "ucm schema"
description: "Print the JSON schema for ucm.yml."
---

# ucm schema

Emits the JSON schema for `ucm.yml`. Pipe into a file and point your editor
at it for autocomplete and validation.

## Syntax

```bash
databricks ucm schema
```

## Example

```bash
databricks ucm schema > ucm.schema.json
```

### VS Code

Add to `.vscode/settings.json`:

```json
{
  "yaml.schemas": {
    "./ucm.schema.json": "ucm.yml"
  }
}
```

### JetBrains IDEs

Settings → Languages & Frameworks → Schemas and DTDs → JSON Schema
Mappings → add a mapping from the generated file to `ucm.yml`.

## See also

- [ucm.yml settings](../settings.md) — the hand-written equivalent.
- [validate](validate.md) — schema conformance + mutator chain.
