#!/bin/bash
./build.sh
./flog -l -b 1024 -r 50000 | go run . -rust
