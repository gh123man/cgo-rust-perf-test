package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func instantiateWazero(ctx context.Context) (wazero.Runtime, api.Module) {
	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	// Instantiate a WebAssembly module that exports
	// "allocate", "deallocate" and "noop_wasm"
	mod, err := r.Instantiate(ctx, compiledWasmBytes)
	if err != nil {
		log.Panicln(err)
	}

	return r, mod
}

func unpackUInt64(val uint64) (uint32, uint32) {
	return uint32(val >> 32), uint32(val)
}

type WazeroRunner struct {
	ctx     context.Context
	mod     api.Module
	runtime wazero.Runtime
	bufPtr  uint32
}

func NewWazeroRunner(ctx context.Context, wasmBytes []byte) *WazeroRunner {
	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	mod, err := r.Instantiate(ctx, wasmBytes)
	if err != nil {
		log.Panicln(err)
	}

	allocate := mod.ExportedFunction("allocate")

	results, err := allocate.Call(ctx, bufSize)
	if err != nil {
		log.Panicln(err)
	}

	bufPtr := results[0]

	return &WazeroRunner{
		ctx: ctx, mod: mod, runtime: r, bufPtr: uint32(bufPtr),
	}
}

func (wr *WazeroRunner) executeStringInStringOut(input string, funcy api.Function) string {
	if len(input) > bufSize {
		log.Panicf("Input string length %d is bigger than the buffer %d.", len(input), bufSize)
	}

	if !wr.mod.Memory().Write(wr.bufPtr, []byte(input)) {
		log.Panicf("Memory.Write(%d, %d) out of range of memory size %d",
			wr.bufPtr, len(input), wr.mod.Memory().Size())
	}

	results, err := funcy.Call(wr.ctx, uint64(wr.bufPtr), uint64(len(input)))
	if err != nil {
		log.Panicln(err)
	}

	resultSize := results[0]

	resultStringBytes, ok := wr.mod.Memory().Read(wr.bufPtr, uint32(resultSize))
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range of memory size %d",
			wr.bufPtr, resultSize, wr.mod.Memory().Size())
	}
	res := string(resultStringBytes)
	return res
}

func (wr *WazeroRunner) runVrl(input string) string {
	vrl := wr.mod.ExportedFunction("vrl_wasm")

	return wr.executeStringInStringOut(input, vrl)
}

func (wr *WazeroRunner) runRegex(input string) string {
	vrl := wr.mod.ExportedFunction("regex_wasm")

	return wr.executeStringInStringOut(input, vrl)
}

func (wr *WazeroRunner) runNoop(input string) string {
	noop := wr.mod.ExportedFunction("noop_wasm")
	return wr.executeStringInStringOut(input, noop)
}

func (wr *WazeroRunner) runNoopDynamicAllocation(input string) string {
	noop := wr.mod.ExportedFunction("noop_wasm_dynamic_allocation")
	allocate := wr.mod.ExportedFunction("allocate")
	deallocate := wr.mod.ExportedFunction("deallocate")

	inputSize := uint64(len(input))

	// Instead of an arbitrary memory offset, use Rust's allocator. Notice
	// there is nothing string-specific in this allocation function. The same
	// function could be used to pass binary serialized data to Wasm.
	results, err := allocate.Call(wr.ctx, inputSize)
	if err != nil {
		log.Panicln(err)
	}

	inputPtr := results[0]
	// This pointer was allocated by Rust, but owned by Go, So, we have to
	// deallocate it when finished
	defer deallocate.Call(wr.ctx, inputPtr, inputSize)

	// The pointer is a linear memory offset, which is where we write the input string.
	if !wr.mod.Memory().Write(uint32(inputPtr), []byte(input)) {
		log.Panicf("Memory.Write(%d, %d) out of range of memory size %d",
			inputPtr, inputSize, wr.mod.Memory().Size())
	}

	// Invoke 'noop' passing in the pointer+size of the input string
	// Result is a packed ptr+size of a rust-allocated string
	packedPtrSize, err := noop.Call(wr.ctx, inputPtr, inputSize)
	if err != nil {
		log.Panicln(err)
	}
	noopResultPtr, noopResultSize := unpackUInt64(packedPtrSize[0])
	// This pointer was allocated by Rust, but owned by Go, So, we have to
	// deallocate it when finished
	defer deallocate.Call(wr.ctx, uint64(noopResultPtr), uint64(noopResultSize))

	// The pointer is a linear memory offset, which is where we write the input string.
	resultStringBytes, ok := wr.mod.Memory().Read(noopResultPtr, noopResultSize)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range of memory size %d",
			noopResultPtr, noopResultSize, wr.mod.Memory().Size())
	}
	res := string(resultStringBytes)
	return res
}

func (wr *WazeroRunner) Close() {
	wr.runtime.Close(wr.ctx)
}

func runWazero() {
	// Choose the context to use for function calls.
	ctx := context.Background()

	runner := NewWazeroRunner(ctx, compiledWasmBytes)
	defer runner.Close() // This closes everything this Runtime created.

	res := runner.runNoop("hello wazero")
	fmt.Println(res)
}
