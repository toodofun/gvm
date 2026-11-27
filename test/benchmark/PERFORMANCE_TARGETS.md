# Performance Targets and Guidelines

## Overview

This document defines performance targets for gvm's security-critical components and provides guidelines for continuous performance monitoring.

## Benchmark Suite

Location: `test/benchmark/`

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./test/benchmark/...

# Run specific benchmark
go test -bench=BenchmarkSafeExecutor_Execute -benchmem ./test/benchmark/

# Run with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./test/benchmark/

# Run with memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./test/benchmark/
```

## Performance Targets

### SafeExecutor Performance

**Benchmark:** `BenchmarkSafeExecutor_Execute`

**Target:**
- **Throughput:** > 10,000 executions/sec
- **Latency:** < 100μs per execution
- **Memory Allocations:** < 100 bytes per execution

**Rationale:** Command execution is a hot path in version management operations. The executor must be fast to minimize overhead during install/switch operations.

### Command Parsing Performance

**Benchmark:** `BenchmarkParseCommand`

**Target:**
- **Throughput:** > 100,000 parses/sec
- **Latency:** < 10μs per parse
- **Memory Allocations:** < 200 bytes per parse

**Rationale:** Command parsing happens for every shell command execution. Must be extremely fast to avoid bottlenecking user workflows.

### Version Validation Performance

**Benchmark:** `BenchmarkValidateVersion`

**Target:**
- **Throughput:** > 1,000,000 validations/sec
- **Latency:** < 1μs per validation
- **Memory Allocations:** 0 (no heap allocations)

**Rationale:** Version validation is called frequently during version listing, comparison, and resolution. Should be essentially free performance-wise.

### Path Validation Performance

**Benchmark:** `BenchmarkValidatePath`

**Target:**
- **Throughput:** > 500,000 validations/sec
- **Latency:** < 2μs per validation
- **Memory Allocations:** < 50 bytes per validation

**Rationale:** Path validation is security-critical and called on every file system operation. Must be fast enough not to discourage use.

### Path Safety Check Performance

**Benchmark:** `BenchmarkIsPathSafe`

**Target:**
- **Throughput:** > 100,000 checks/sec
- **Latency:** < 10μs per check
- **Memory Allocations:** < 500 bytes per check

**Rationale:** Path traversal prevention requires syscalls and absolute path resolution. More expensive but still must be fast.

## Continuous Monitoring

### Baseline Establishment

1. **Initial Baseline:** Run benchmarks on clean state
   ```bash
   go test -bench=. -benchmem ./test/benchmark/... > baseline.txt
   ```

2. **Document Results:** Record baseline results in `baseline.txt`

3. **Version Control:** Commit baseline to track performance over time

### Regression Detection

**Automated Checks:**
- Run benchmarks in CI/CD pipeline
- Compare against baseline using `benchstat`
- Fail build if performance regresses > 20%

**Manual Reviews:**
- Review benchmark results weekly
- Investigate any > 10% performance changes
- Document reasons for expected performance changes

### Performance Regression Testing

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Compare current vs baseline
benchstat baseline.txt current.txt
```

**Interpreting Results:**
- **~ (tilde):** No statistically significant change
- **+ (plus):** Performance improved (faster)
- **- (minus):** Performance regressed (slower)

## Performance Optimization Guidelines

### When to Optimize

1. **Before:** Always write benchmarks first (TDD)
2. **Measure:** Profile before optimizing
3. **Target:** Focus on hot paths identified by profiling
4. **Verify:** Re-run benchmarks after optimization

### Optimization Priorities

1. **Critical Path:** Version installation and switching
2. **Security Validation:** Must not discourage secure practices
3. **User Experience:** CLI responsiveness

### Common Anti-Patterns

❌ **Don't:**
- Remove security checks for performance
- Optimize without benchmarking
- Micro-optimize non-hot paths
- Cache security validation results

✅ **Do:**
- Profile first, optimize second
- Focus on algorithmic improvements
- Reduce allocations in hot paths
- Maintain security guarantees

## Benchmark Maintenance

### Adding New Benchmarks

1. Create benchmark function: `func BenchmarkXxx(b *testing.B)`
2. Follow naming convention: `Benchmark<FunctionBeingTested>`
3. Use `b.ResetTimer()` before timed section
4. Document performance targets in this file

### Updating Benchmarks

When code changes affect performance:
1. Update benchmark if interface changed
2. Re-run all benchmarks
3. Update baseline.txt if change is intentional
4. Document reason for performance change

### Benchmark Best Practices

```go
func BenchmarkExample(b *testing.B) {
    // Setup (not timed)
    setup()

    b.ResetTimer()  // Reset timer before loop
    for i := 0; i < b.N; i++ {
        // Code to benchmark
        functionUnderTest()
    }
}
```

**Tips:**
- Keep benchmarks simple and focused
- Avoid I/O in benchmarks (or mock it)
- Use `b.ReportAllocs()` to track allocations
- Run multiple times for stable results

## References

- [Go Benchmark Documentation](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [benchstat Tool](https://golang.org/x/perf/cmd/benchstat)
- [pprof Profiling](https://golang.org/pkg/net/http/pprof/)

## Changelog

### 2026-03-29
- Initial performance targets established
- Benchmark suite created for security modules
- Baseline documentation started (pending Go installation fix)
