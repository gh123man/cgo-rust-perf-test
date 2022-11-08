#!/bin/bash

./build.sh
RUST_BACKTRACE=1 go run . -vrl -stdout
