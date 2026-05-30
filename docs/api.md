# API Documentation

## Endpoints

### Health Check
`GET /healthz`

Returns the operational status of the service.

**Response:**
- `200 OK`
```json
{
  "status": "ok"
}
```

### Run Code
`POST /run`

Executes source code against a set of test cases.

**Request Body:**
| Field | Type | Description |
| :--- | :--- | :--- |
| `language` | string | Language ID (e.g., `c`, `cpp`, `py3`) |
| `source` | string | Source code content |
| `source_filename` | string | Name of the source file (e.g., `main.c`) |
| `artifact_filename` | string | Name of the output binary (for compiled languages) |
| `build` | object | Optional build configuration (limits, flags) |
| `run` | object | Execution configuration (limits, flags) |
| `tests` | array | List of test cases (stdin and expected stdout) |

**Example Request:**
```json
{
  "language": "c",
  "source": "#include <stdio.h>\nint main() { printf(\"hello world\"); return 0; }",
  "source_filename": "main.c",
  "artifact_filename": "main",
  "run": {
    "limits": { "wall_time_s": 2, "memory_kb": 65536 }
  },
  "tests": [
    { "stdin": "", "expected_stdout": "hello world" }
  ]
}
```

**Response Body:**
| Field | Type | Description |
| :--- | :--- | :--- |
| `status` | string | Overall execution status |
| `build` | object | Details of the compilation phase (stdout, stderr, status) |
| `tests` | array | Results for each test case execution |

**Example Response:**
```json
{
  "status": "ok",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": ""
  },
  "tests": [
    {
      "status": "ok",
      "stdout": "hello world",
      "stderr": "",
      "duration_ms": 10,
      "memory_peak_kb": 1024
    }
  ]
}
```
