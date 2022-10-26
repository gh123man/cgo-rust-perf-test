#!/bin/bash 

cargo build --release
RUST_BACKTRACE=1 go run . -vrl