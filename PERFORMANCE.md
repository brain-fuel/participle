# Performance

Apple M5 Max, Darwin arm64, Go 1.26, upstream Participle v2.1.4. The benchmark
parses 256 assignments into semantic AST values; parser construction is outside
the timed loop. Five runs were recorded with `-benchmem`.

| implementation | ns/op range | B/op | allocs/op |
|---|---:|---:|---:|
| GoForge | 6,226–6,346 | 30,104 | 263 |
| upstream | 242,483–248,945 | 683,760–683,761 | 8,237 |

Using the least favorable endpoints, GoForge is 38.2× faster and uses 95.6%
fewer bytes plus 96.8% fewer allocations. This exceeds the release gates of 2×
throughput and 50% fewer allocations with substantial margin.

```sh
go test -run '^$' -bench BenchmarkAssignments -benchmem -count=5
```
