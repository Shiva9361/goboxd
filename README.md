# goboxd

goboxd is a secure code execution service that runs user-provided source code in isolated sandboxes using Linux namespaces and nsjail.

Using gin as it does not require too much overhead, widely used and is pretty performant. Morover it is easy to add middleware for basically anything in the future.

Nsjail's r_limit_as enforces limit on virtual memory. This leads to java and javascript crashing with oom even before executing the program. What 
we need is physical memory limit, this is provided by cgroup. The issue is that cgroup along with docker can be a security risk. Given docker is 
just for dev and testing it should be fine. But if deployed, a ephemeral vm or a complete system instead of docker should be used. 

The above is from my understanding, feel free to correct me if I am wrong.

## Features

- **Sandboxed Execution**: Uses nsjail to isolate processes, providing filesystem, network, and resource isolation.
- **Resource Management**: Enforces limits on wall-time, memory usage, and maximum process count.
- **Multi-language Support**: Configurable support for C, C++, Python 3, Java, Javascript, bash and verilog currently.
- **API-Driven**: Provides HTTP endpoints for health monitoring, readiness checks, and code execution.
- **Automated Testing**: Supports executing multiple test cases per request with stdin and expected stdout validation.

## Installation

1. Clone the repository.
2. Ensure docker is setup and running
3. Build the application:
   ```bash
   make build
   ```

## Usage

Start the server:
```bash
make run
```

The server listens on the default Gin port.

### Unit Tests
- test file corresponding to each file is present alongside it.
```bash
make test
```

### Integration Tests
- 5 test files for each langauage that is supported
```bash
make integration
```

### Benchmarks
- uses hey
```bash
make benchmark
```

### API Endpoints

- `GET /healthz`: Basic liveness check.
- `GET /readyz`: Deep readiness check verifying nsjail and configured language runtimes.
- `POST /run`: Execute code. Requires a JSON payload containing language, source code, and test cases.

## Configuration

Language configurations and sandbox flags are managed in `config/settings.yaml`. Each language entry defines build commands, execution commands, and resource limits.

## Documentation

Detailed information is available in the `docs/` directory:
- [Architecture](docs/architecture.md)
- [API Reference](docs/api.md)
- [Benchmarks](docs/benchmarks.md)
