#!/usr/bin/env bash

arch=$(uname -m)
if [[ "$arch" == "arm64" ]]; then
    arch="aarch64"
fi
targets=("--target wasm32-wasi")
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    targets+=("--target $arch-unknown-linux-gnu")
elif [[ "$OSTYPE" == "darwin"* ]]; then
    targets+=("--target $arch-apple-darwin")
fi

target_param=$(IFS=" "; echo "${targets[*]}")

cargo build --release ${target_param} \
    && go build
