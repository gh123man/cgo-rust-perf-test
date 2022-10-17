
## Prerequisites 

- rust toolchain
- go toolchain 

That's it to run the benchmarks, just run `cargo build` first. 

To run the test scripts

- either be on an M1 mac and use the `flog` binary in the repo, [or build this first](https://github.com/DataDog/flog)
- `brew install pv` (pipe viewer)


## Benchmark results

```
goos: darwin
goarch: arm64
pkg: cgotest
BenchmarkRustRegex-10          	       9	 120307569 ns/op	11840053 B/op	   30000 allocs/op
BenchmarkGoRegex-10            	       5	 202554858 ns/op	43821992 B/op	   70044 allocs/op
BenchmarkRustPassthrough-10    	     164	   7245905 ns/op	11840022 B/op	   30000 allocs/op
BenchmarkGoPassthrough-10      	    1059	   1154868 ns/op	11520027 B/op	   10000 allocs/op
PASS
ok  	cgotest	7.443s
```