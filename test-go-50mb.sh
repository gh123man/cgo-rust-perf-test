#!/bin/bash 
cargo build --release
./flog -l -b 1024 -r 50000 | go run . | pv > /dev/null