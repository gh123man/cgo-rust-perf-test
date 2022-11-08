#!/usr/bin/env bash

targets=("--target wasm32-wasi")
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    targets+=("--target aarch64-unknown-linux-gnu")
elif [[ "$OSTYPE" == "darwin"* ]]; then
    targets+=("--target aarch64-apple-darwin")
fi

target_param=$(IFS=" "; echo "${targets[*]}")

cargo build --release ${target_param} \
    && go build
