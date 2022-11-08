#!/bin/bash
./build.sh
./flog -l -b 1024 -r 5000 | go run .
