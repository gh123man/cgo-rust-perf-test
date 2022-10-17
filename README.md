
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
BenchmarkRustRegex1-10               	   83535	     12462 ns/op	    1184 B/op	       3 allocs/op
BenchmarkRustRegex10-10              	    9603	    124965 ns/op	   11840 B/op	      30 allocs/op
BenchmarkRustRegex100-10             	     970	   1236614 ns/op	  118402 B/op	     300 allocs/op
BenchmarkRustRegex1000-10            	      90	  12440297 ns/op	 1184008 B/op	    3000 allocs/op
BenchmarkRustRegex10000-10           	       9	 120249370 ns/op	11840042 B/op	   30000 allocs/op
BenchmarkRustRegex100000-10          	       1	1182302042 ns/op	118400576 B/op	  300006 allocs/op
BenchmarkGoRegex1-10                 	   57272	     20347 ns/op	    4381 B/op	       7 allocs/op
BenchmarkGoRegex10-10                	    5848	    204508 ns/op	   43803 B/op	      70 allocs/op
BenchmarkGoRegex100-10               	     594	   2053824 ns/op	  438393 B/op	     700 allocs/op
BenchmarkGoRegex1000-10              	      58	  20642690 ns/op	 4383709 B/op	    7004 allocs/op
BenchmarkGoRegex10000-10             	       5	 204252575 ns/op	43866638 B/op	   70051 allocs/op
BenchmarkGoRegex100000-10            	       1	2033331541 ns/op	438069768 B/op	  700412 allocs/op
BenchmarkRustPassthrough1-10         	 1638686	       728.7 ns/op	    1184 B/op	       3 allocs/op
BenchmarkRustPassthrough10-10        	  165696	      7460 ns/op	   11840 B/op	      30 allocs/op
BenchmarkRustPassthrough100-10       	   16524	     73184 ns/op	  118400 B/op	     300 allocs/op
BenchmarkRustPassthrough1000-10      	    1580	    729162 ns/op	 1184001 B/op	    3000 allocs/op
BenchmarkRustPassthrough10000-10     	     164	   7358990 ns/op	11840024 B/op	   30000 allocs/op
BenchmarkRustPassthrough100000-10    	      15	  73081711 ns/op	118400243 B/op	  300002 allocs/op
BenchmarkGoPassthrough1-10           	 9461560	       121.0 ns/op	    1152 B/op	       1 allocs/op
BenchmarkGoPassthrough10-10          	 1000000	      1175 ns/op	   11520 B/op	      10 allocs/op
BenchmarkGoPassthrough100-10         	  103849	     12657 ns/op	  115200 B/op	     100 allocs/op
BenchmarkGoPassthrough1000-10        	    9368	    125586 ns/op	 1152003 B/op	    1000 allocs/op
BenchmarkGoPassthrough10000-10       	    1030	   1225275 ns/op	11520027 B/op	   10000 allocs/op
BenchmarkGoPassthrough100000-10      	     100	  12555695 ns/op	115200277 B/op	  100002 allocs/op
PASS
ok  	cgotest	33.895s
```