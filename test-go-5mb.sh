#!/bin/bash 
cargo build
./flog -l -b 1024 -r 5000 | go run . | pv > /dev/null