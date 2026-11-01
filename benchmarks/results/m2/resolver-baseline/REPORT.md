# Eloque M2 resolver baseline

Go benchmark: `go test ./internal/semantic/resolver -run ^$ -bench ^BenchmarkResolve -benchmem -benchtime=1s -count=10`

| Benchmark | Samples | Median ns/op | Median B/op | Median allocs/op |
|---|---:|---:|---:|---:|
| `BenchmarkResolve/declarations/10` | 10 | 662.95 | 968.00 | 8.00 |
| `BenchmarkResolve/declarations/100` | 10 | 5,002.50 | 7,960.00 | 8.00 |
| `BenchmarkResolve/declarations/1000` | 10 | 55,170.50 | 95,688.50 | 10.00 |
| `BenchmarkResolve/declarations/3000` | 10 | 159,275.50 | 232,224.00 | 14.00 |
| `BenchmarkResolve/mixed/10` | 10 | 1,482.50 | 1,336.00 | 8.00 |
| `BenchmarkResolve/mixed/100` | 10 | 12,798.00 | 11,160.00 | 8.00 |
| `BenchmarkResolve/mixed/1000` | 10 | 122,666.50 | 128,458.00 | 10.00 |
| `BenchmarkResolve/mixed/3000` | 10 | 386,735.00 | 330,535.00 | 14.00 |
| `BenchmarkResolve/references/10` | 10 | 408.85 | 448.00 | 6.00 |
| `BenchmarkResolve/references/100` | 10 | 1,912.50 | 1,248.00 | 6.00 |
| `BenchmarkResolve/references/1000` | 10 | 16,893.00 | 8,544.00 | 6.00 |
| `BenchmarkResolve/references/3000` | 10 | 48,433.50 | 24,929.00 | 6.00 |
| `BenchmarkResolve/sparse-undefined/10` | 10 | 653.45 | 1,504.00 | 8.00 |
| `BenchmarkResolve/sparse-undefined/100` | 10 | 2,211.50 | 2,304.00 | 8.00 |
| `BenchmarkResolve/sparse-undefined/1000` | 10 | 18,026.50 | 11,936.00 | 18.00 |
| `BenchmarkResolve/sparse-undefined/3000` | 10 | 52,423.50 | 33,058.00 | 39.00 |
| `BenchmarkResolve/undefined/10` | 10 | 1,275.50 | 3,536.00 | 15.00 |
| `BenchmarkResolve/undefined/100` | 10 | 11,883.00 | 35,907.00 | 108.00 |
| `BenchmarkResolve/undefined/1000` | 10 | 98,668.50 | 276,818.00 | 1,011.00 |
| `BenchmarkResolve/undefined/3000` | 10 | 442,793.50 | 1,045,337.00 | 3,013.00 |

These measurements isolate the M2 resolver; they do not include parsing or full semantic analysis.

Raw results: `samples.csv` and `go_benchmark.log`.
