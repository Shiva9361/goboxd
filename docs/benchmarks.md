# Performance Benchmarks

**Environment**
* **Host Hardware:** 16GB RAM, Intel Core i7-11800H
* **Target:** `goboxd:dev` runtime image
* **Endpoint:** `POST /run`

**Workload: Python 3 ("Hello World")**
* **Payload:** `print('Hello, World!')`
* **Sandbox Limits:** 2s wall time, 64MB memory

### Results

| Concurrency | Total Requests | Requests / Sec | p50 Latency | p95 Latency | p99 Latency | Success Rate |
|-------------|----------------|----------------|-------------|-------------|-------------|--------------|
| **1** | 100            | 1899.35        | 0.30 ms     | 1.20 ms     | 4.00 ms     | 100% (200 OK)|
| **10** | 500            | 5667.69        | 1.40 ms     | 4.30 ms     | 5.50 ms     | 100% (200 OK)|
| **50** | 1000           | 9091.79        | 2.70 ms     | 20.70 ms    | 31.80 ms    | 100% (200 OK)|
| **100** | 2000           | 11464.76       | 3.20 ms     | 32.30 ms    | 64.00 ms    | 100% (200 OK)|
