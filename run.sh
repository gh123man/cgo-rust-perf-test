#!/bin/bash 

cargo build
RUST_BACKTRACE=1 go run . -rust