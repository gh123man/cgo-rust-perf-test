#!/bin/bash 
cargo build --release
./flog -l -b 1024 -r 5000 | go run . -rust | pv > /dev/null