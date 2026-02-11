# bark

**bark** is a blazingly fast, zero-allocation structured logging library for Go. Designed for high-throughput systems and low-latency environments, it provides both a traditional JSON output and a specialized binary format for maximum efficiency.

## Features

-   **Zero Allocations**: Leverages `sync.Pool` and pre-allocated buffers to ensure no heap allocations occur during the logging hot path.
    
-   **Dual Format Support**:
    
    -   **JSON**: Human-readable and industry-standard structured logs.
        
    -   **Binary**: A compact, tagged binary protocol for extreme performance and reduced I/O bandwidth.
        
-   **Architectural Optimizations**:
    
    -   **Custom Time Formatting**: Bypasses `time.Format` to avoid layout string parsing overhead.
        
    -   **Optimized Escaping**: Custom JSON string escaping implementation.
        
    -   **Minimal Dependencies**: Only relies on the Go standard library.
        
-   **Rich Type Support**: Chainable API supporting `Int`, `Uint`, `Float`, `Complex`, `Bool`, `Bytes`, `Error`, and `Str`.
    

## Benchmarks

Results obtained on an **Apple M1 (arm64)**. Both loggers achieve **0 B/op** by reusing memory via internal pools.

| Benchmark | Iterations | Time | Memory | Allocs |
| :------ | :--: | :-----------: | :---: | :---------------: |
| JSON | 5,316,435 | 198.1 ns/op | 0 B/op | 0 allocs/op |
| Binary | 12,058,200 | 98.94 ns/op | 0 B/op | 0 allocs/op |
_To run benchmarks yourself:_ `go test -bench=. -benchmem`

## Installation

```
go get github.com/banditmoscow1337/bark

```

## Usage

### JSON Logging

Ideal for cloud environments (ELK, Datadog, etc.) where human readability or standard ingestion is required.

```
package main

import (
	"os"
	"github.com/banditmoscow1337/bark"
)

func main() {
	logger := bark.NewLogger(os.Stdout)
	
	logger.Info().
		Str("user_id", "u123").
		Int("attempt", 3).
		Bool("success", true).
		Msg("user login attempt")
}

```

### Binary Logging

Ideal for internal microservices, high-frequency telemetry, or edge computing where performance and disk/network I/O are the primary constraints.

```
package main

import (
	"os"
	"github.com/banditmoscow1337/bark"
)

func main() {
	// The binary format uses a tagged-length-value approach
	logger := bark.NewBinaryLogger(os.Stdout)
	
	logger.Info().
		Float64("temp", 36.6).
		Uint64("id", 882211).
		Msg("sensor_read")
}

```

## Binary Protocol Specification

The binary format follows a strict structure for fast parsing:

1.  **Header (6 bytes)**: 2 bytes for Type, 4 bytes for Payload Length.
    
2.  **Timestamp (8 bytes)**: Nanoseconds since epoch (Little Endian).
    
3.  **Fields**: `[Key Length (1b)][Key][Tag (1b)][Value]`
    
    -   Strings/Bytes use a 2-byte length prefix.
        
    -   Numbers use standard fixed-width Little Endian encoding.
        

## License

[MIT](LICENSE)
