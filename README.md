# grpcannon

A lightweight gRPC load testing CLI with configurable concurrency and latency histograms.

---

## Installation

```bash
go install github.com/yourname/grpcannon@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/grpcannon.git && cd grpcannon && go build -o grpcannon .
```

---

## Usage

```bash
grpcannon [options] <target>
```

### Example

```bash
grpcannon \
  --proto ./api/service.proto \
  --call helloworld.Greeter/SayHello \
  --data '{"name": "world"}' \
  --concurrency 50 \
  --requests 1000 \
  localhost:50051
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--proto` | | Path to the `.proto` file |
| `--call` | | Fully qualified method name |
| `--data` | | JSON request payload |
| `--concurrency` | `10` | Number of concurrent workers |
| `--requests` | `200` | Total number of requests |
| `--timeout` | `5s` | Per-request timeout |

### Sample Output

```
Summary:
  Total requests : 1000
  Concurrency    : 50
  Duration       : 3.24s
  Throughput     : 308.6 req/s

Latency histogram:
  p50  : 12.4ms
  p90  : 28.1ms
  p95  : 41.7ms
  p99  : 89.3ms
  max  : 134.2ms
```

---

## License

MIT © 2024 yourname