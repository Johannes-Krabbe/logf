# logf

An interactive JSON log viewer for the terminal. Pipe in structured logs and filter them in real time.

```
your-app | logf
```

## Features

- **Live filtering** — type filter expressions to narrow down logs instantly
- **Pretty-print transform** — render raw JSON lines as human-readable, colored output
- **Nested field access** — use dot notation for nested JSON fields (e.g. `request.method="GET"`)

## Filter syntax

Filters are typed directly into the prompt. Multiple filters are combined with `&` or spaces (AND logic).

| Operator | Example | Description |
|----------|---------|-------------|
| `=` | `level=info` | Exact match |
| `!=` | `level!=debug` | Not equal |
| `>` `>=` `<` `<=` | `status>400` | Numeric comparison |
| `~` | `msg~timeout` | Contains substring |

Quoted values are supported: `msg="hello world"`.

## Configuration

Create a `logf.json` file in your working directory:

```json
{
  "transformLogs": true,
  "transform": {
    "timestamp": "timestamp",
    "level": "level",
    "message": "msg"
  }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `transformLogs` | `false` | Enable pretty-print transform |
| `transform.timestamp` | `"timestamp"` | Dot-path to the timestamp field |
| `transform.level` | `"level"` | Dot-path to the level field |
| `transform.message` | auto (`"message"` then `"msg"`) | Dot-path to the message field |

### Pretty-print output

When `transformLogs` is enabled, JSON lines are rendered as:

```
[28.02.2026 20:12:29] INFO: request completed
    reqId: "Xkfqp7kcSs"
    request: {
      "method": "GET"
    }
    durationMs: 3
```

Levels are colored: red for error, yellow for warn, green for info, gray for debug.

Timestamps are parsed from RFC3339, unix seconds, or unix milliseconds and displayed in local time.

Non-JSON lines pass through unchanged. Filtering always operates on the raw JSON, regardless of transform.
