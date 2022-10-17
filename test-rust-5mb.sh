#!/bin/bash 
cargo build
./flog -l -b 1024 -r 5000 | go run . -rust | pv > /dev/null