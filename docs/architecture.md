# Architecture

## Overview
goboxd is a secure code execution service that runs user-submitted code in a sandboxed environment using `nsjail`. It supports multiple programming languages via a configurable backend.

## Components

### 1. API Server (`internal/server.go`)
- Built with the Gin web framework.
- Handles incoming HTTP requests for code execution and health/readiness checks.
- Manages unique UIDs for process isolation to prevent cross-process interference in the sandbox.

### 2. Code Runner (`internal/coderunner.go`)
- Orchestrates the compilation and execution lifecycle.
- Loads language configurations from `config/settings.yaml`.
- Uses `nsjail` to enforce resource limits (CPU time, memory, process count) and filesystem isolation.

### 3. Sandbox Environment
- **nsjail**: A light-weight process isolation tool for Linux.
- **Temporary Workspaces**: Each execution gets a unique temporary directory mounted into the jail, which is cleaned up immediately after execution.
- **Resource Constraints**: Wall time and memory limits are enforced at the sandbox level. Memory limits are managed via cgroups to avoid OOM issues with languages like Java and JavaScript.

## Execution Flow
1. **Request Validation**: Validates the language and payload structure.
2. **Workspace Setup**: Creates a temporary directory and writes the source code.
3. **Compilation (Optional)**: Executes the configured build command within `nsjail`.
4. **Execution**: Runs the resulting artifact/script for each provided test case inside `nsjail`.
5. **Result Aggregation**: Collects stdout, stderr, and execution metrics to return to the client.
6. **Cleanup**: Deletes the temporary workspace and releases the unique UID.
