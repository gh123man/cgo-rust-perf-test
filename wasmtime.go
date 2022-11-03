package main

import (
	"fmt"
	"log"

	"github.com/bytecodealliance/wasmtime-go"
)

var logImportExports = false

func unpackInt64(val int64) (int32, int32) {
	return int32(val >> 32), int32(val)
}

func printExternType(ty *wasmtime.ExternType) {
	if ft := ty.FuncType(); ft != nil {
		log.Print("\tFunction:", ft)
		for i, param := range ft.Params() {
			log.Printf("\t\tParam %d: type: %s - %s", i, param.Kind().String(), param.String())
		}
		for i, result := range ft.Results() {
			log.Printf("\t\tResult %d: type: %s - %s", i, result.Kind().String(), result.String())
		}
	}
	if gt := ty.GlobalType(); gt != nil {
		log.Print("Global:", gt)
	}
	if mt := ty.MemoryType(); mt != nil {
		log.Print("Memory:", mt)
	}
	if tt := ty.TableType(); tt != nil {
		log.Print("Table:", tt)
	}
}

type WasmtimeRunner struct {
	instance *wasmtime.Instance
	store    *wasmtime.Store
	bufPtr   int32
}

func NewWasmtimeRunner(wasmBytes []byte) *WasmtimeRunner {
	engine := wasmtime.NewEngine()
	module, err := wasmtime.NewModule(engine, compiledWasmBytes)
	if err != nil {
		log.Panicln(err)
	}

	if logImportExports {
		log.Print("Listing imports requested by module")
		for i, imp := range module.Imports() {
			log.Printf("Import #%d - %q", i, *imp.Name())
			log.Print("\tModule:", imp.Module(), "  Type:", *imp.Type())
		}
		log.Print("Listing exports from module")
		for i, exp := range module.Exports() {
			log.Printf("Export #%d - %q", i, exp.Name())
			printExternType(exp.Type())
		}
	}

	// Create a linker with WASI functions defined within it
	linker := wasmtime.NewLinker(engine)
	err = linker.DefineWasi()
	if err != nil {
		log.Panicln(err)
	}

	// Configure WASI imports to write stdout into a file, and then create
	// a `Store` using this wasi configuration.
	wasiConfig := wasmtime.NewWasiConfig()
	store := wasmtime.NewStore(engine)
	store.SetWasi(wasiConfig)
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		log.Panicln(err)
	}

	// Pre-allocate buffer to use
	allocate := instance.GetExport(store, "allocate").Func()

	result, err := allocate.Call(store, bufSize)
	if err != nil {
		log.Panicln(err)
	}

	bufPtr := result.(int32)

	return &WasmtimeRunner{instance, store, bufPtr}
}

func (wr *WasmtimeRunner) runStringInStringOut(input string, funcy *wasmtime.Func) string {
	if len(input) > bufSize {
		log.Panicf("Input string length %d is bigger than the buffer %d.", len(input), bufSize)
	}
	memory := wr.instance.GetExport(wr.store, "memory").Memory()
	memoryBuf := memory.UnsafeData(wr.store)

	inputSize := int32(len(input))

	copy(memoryBuf[wr.bufPtr:], input)

	result, err := funcy.Call(wr.store, wr.bufPtr, inputSize)
	if err != nil {
		log.Panicln(err)
	}

	resultSize := result.(int32)
	// Refresh memoryBuf, after a `.Call` it is invalid
	memoryBuf = memory.UnsafeData(wr.store)

	start := wr.bufPtr
	end := int64(wr.bufPtr + resultSize)

	return string(memoryBuf[start:end])
}

func (wr *WasmtimeRunner) runVrl(input string) string {
	vrl := wr.instance.GetExport(wr.store, "vrl_wasm").Func()

	return wr.runStringInStringOut(input, vrl)
}

func (wr *WasmtimeRunner) runRegex(input string) string {
	vrl := wr.instance.GetExport(wr.store, "regex_wasm").Func()

	return wr.runStringInStringOut(input, vrl)
}

func (wr *WasmtimeRunner) runNoop(input string) string {
	vrl := wr.instance.GetExport(wr.store, "noop_wasm").Func()

	return wr.runStringInStringOut(input, vrl)
}

func (wr *WasmtimeRunner) runNoopDynamicAllocation(input string) string {
	// Load up our exports from the wr.instance
	memory := wr.instance.GetExport(wr.store, "memory").Memory()
	memoryBuf := memory.UnsafeData(wr.store)

	noop := wr.instance.GetExport(wr.store, "noop_wasm_dynamic_allocation").Func()
	allocate := wr.instance.GetExport(wr.store, "allocate").Func()
	deallocate := wr.instance.GetExport(wr.store, "deallocate").Func()

	inputSize := int32(len(input))
	result, err := allocate.Call(wr.store, inputSize)
	if err != nil {
		log.Panicln(err)
	}

	inputPtr := result.(int32)
	defer deallocate.Call(wr.store, inputPtr, inputSize)

	copy(memoryBuf[inputPtr:], input)

	packedPtrSize, err := noop.Call(wr.store, inputPtr, inputSize)
	if err != nil {
		log.Panicln(err)
	}
	noopResultPtr, noopResultSize := unpackInt64(packedPtrSize.(int64))
	defer deallocate.Call(wr.store, int64(noopResultPtr), int64(noopResultSize))

	// Refresh memoryBuf, after a `.Call` it is invalid
	memoryBuf = memory.UnsafeData(wr.store)

	return string(memoryBuf[noopResultPtr : noopResultPtr+noopResultSize])
}

func runWasmtime() {
	runner := NewWasmtimeRunner(compiledWasmBytes)
	res := runner.runNoop("hello wasmtime")
	fmt.Println(res)
}
