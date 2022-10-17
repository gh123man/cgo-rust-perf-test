#!/bin/bash 
cargo build
./flog -l -b 1024 -r 50000 | go run . -rust | pv > /dev/null