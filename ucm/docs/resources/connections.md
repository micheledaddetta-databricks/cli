---
title: "connections"
description: "Declare Unity Catalog foreign-catalog connections (MySQL, Snowflake, PostgreSQL, ...)."
---

# connections

A UC foreign-catalog connection — the federation link that lets a foreign
catalog reference tables in MySQL, PostgreSQL, Snowflake, Redshift, BigQuery,
and other supported systems.

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Connection name in UC. |
| `connection_type` | string | yes | `MYSQL`, `POSTGRESQL`, `SNOWFLAKE`, `REDSHIFT`, `BIGQUERY`, or any other value accepted by the SDK. |
| `options` | map[string]string | yes | Connection-specific configuration (host, port, user, password, ...). Must contain enough keys for UC to authenticate. |
| `comment` | string | no | Human-readable description. |
| `properties` | map[string]string | no | Additional connection properties. |
| `read_only` | bool | no | Whether the connection is read-only. Default `false`. |

## Example (PostgreSQL)

```yaml
resources:
  connections:
    warehouse_pg:
      name: warehouse_pg
      connection_type: POSTGRESQL
      comment: read-only replica of the ops warehouse
      options:
        host: ops-pg.acme.internal
        port: "5432"
        user: uc_reader
        password: ${var.warehouse_pg_password}
      read_only: true
```

## Example (Snowflake)

```yaml
resources:
  connections:
    snowflake_fin:
      name: snowflake_fin
      connection_type: SNOWFLAKE
      options:
        host: acme-fin.snowflakecomputing.com
        user: UC_SVC
        password: ${var.sf_fin_password}
        db: FINANCE
        warehouse: FIN_WH
```

## Engines

| Engine | Behavior |
|---|---|
| terraform | `databricks_connection.<key>`. |
| direct | `w.Connections.Create` / `.Update` / `.Delete`. |

## Secret handling

Use [`variables`](../variables.md) + the `DATABRICKS_UCM_VAR_*` environment
variables (or `--var`) to inject passwords and tokens at deploy time — do
not commit secret material to `ucm.yml`.

## See also

- [variables](../variables.md)
- [grants](grants.md) — connections can be granted with `securable.type: connection`.
