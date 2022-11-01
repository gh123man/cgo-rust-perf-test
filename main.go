package main

//#cgo CFLAGS: -I${SRCDIR}/target/release/
//#cgo LDFLAGS: -L${SRCDIR}/target/release -Wl,-rpath,${SRCDIR}/target/release -lhelloRust
//
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include "helloRust.h"
import "C"
import (
	"bufio"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/dustin/go-humanize"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"go.uber.org/atomic"
)

const (
	sockAddr     = "/tmp/cgo.sock"
	wasmPageSize = 65536
)

func getUdsReader() *bufio.Reader {
	if _, err := os.Stat(sockAddr); err == nil {
		if err := os.RemoveAll(sockAddr); err != nil {
			log.Fatal(err)
		}
	}

	listener, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Waiting for connection...")
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Accepted connection from ", conn.RemoteAddr().Network())

	return bufio.NewReader(conn)
}

type throughputRecorder struct {
	start      time.Time
	totalBytes atomic.Float64
}

func (tr *throughputRecorder) Record(nBytes int) {
	if time.Time.IsZero(tr.start) {
		tr.start = time.Now()
	}
	tr.totalBytes.Add(float64(nBytes))
}

func (tr *throughputRecorder) AvgThroughput() string {
	now := time.Now()

	elapsed := now.Sub(tr.start)
	avgBytes := uint64(tr.totalBytes.Load() / elapsed.Seconds())
	return fmt.Sprintf("%s / second", humanize.Bytes(avgBytes))
}

func getBlackholeWriter(tr *throughputRecorder) func(a ...any) (int, error) {

	blackhole := func(a ...any) (int, error) {
		s := a[0].(string)
		l := len(s)
		tr.Record(l)
		return l, nil
	}

	return blackhole
}

type OutFunc func(a ...any) (int, error)

// rustWasm was compiled using `cargo build --release --target wasm32-wasi`
// VRL currently cannot build on wasm32-unknown-unknown, so we target wasm32-wasi
//
//go:embed target/wasm32-wasi/release/helloRust.wasm
var compiledWasmBytes []byte

func instantiateWazero(ctx context.Context) (wazero.Runtime, api.Module) {
	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	// Instantiate a WebAssembly module that exports
	// "allocate", "deallocate" and "noop_wasm"
	mod, err := r.InstantiateModuleFromBinary(ctx, compiledWasmBytes)
	if err != nil {
		log.Panicln(err)
	}

	return r, mod
}

func unpackUInt64(val uint64) (uint32, uint32) {
	return uint32(val >> 32), uint32(val)
}

func unpackInt64(val int64) (int32, int32) {
	return int32(val >> 32), int32(val)
}

func invokeNoopViaWazero(ctx context.Context, mod api.Module, input string) string {
	// Get references to WebAssembly functions
	noop := mod.ExportedFunction("noop_wasm")
	allocate := mod.ExportedFunction("allocate")
	deallocate := mod.ExportedFunction("deallocate")

	inputSize := uint64(len(input))

	// Instead of an arbitrary memory offset, use Rust's allocator. Notice
	// there is nothing string-specific in this allocation function. The same
	// function could be used to pass binary serialized data to Wasm.
	results, err := allocate.Call(ctx, inputSize)
	if err != nil {
		log.Panicln(err)
	}

	inputPtr := results[0]
	// This pointer was allocated by Rust, but owned by Go, So, we have to
	// deallocate it when finished
	defer deallocate.Call(ctx, inputPtr, inputSize)

	// The pointer is a linear memory offset, which is where we write the input string.
	if !mod.Memory().Write(ctx, uint32(inputPtr), []byte(input)) {
		log.Panicf("Memory.Write(%d, %d) out of range of memory size %d",
			inputPtr, inputSize, mod.Memory().Size(ctx))
	}

	// Invoke 'noop' passing in the pointer+size of the input string
	// Result is a packed ptr+size of a rust-allocated string
	packedPtrSize, err := noop.Call(ctx, inputPtr, inputSize)
	if err != nil {
		log.Panicln(err)
	}
	noopResultPtr, noopResultSize := unpackUInt64(packedPtrSize[0])
	// This pointer was allocated by Rust, but owned by Go, So, we have to
	// deallocate it when finished
	defer deallocate.Call(ctx, uint64(noopResultPtr), uint64(noopResultSize))

	// The pointer is a linear memory offset, which is where we write the input string.
	resultStringBytes, ok := mod.Memory().Read(ctx, noopResultPtr, noopResultSize)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range of memory size %d",
			noopResultPtr, noopResultSize, mod.Memory().Size(ctx))
	}
	res := string(resultStringBytes)
	return res
}

func runWazero() {
	// Choose the context to use for function calls.
	ctx := context.Background()
	r, wazeroMod := instantiateWazero(ctx)
	defer r.Close(ctx) // This closes everything this Runtime created.

	fmt.Println("Result from wazero wasm call:", invokeNoopViaWazero(ctx, wazeroMod, "Hello wasm world"))
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

func runWasmWithWasmtime(input string) string {
	engine := wasmtime.NewEngine()
	module, err := wasmtime.NewModule(engine, compiledWasmBytes)
	if err != nil {
		log.Panicln(err)
	}

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

	// Load up our exports from the instance
	memory := instance.GetExport(store, "memory").Memory()
	memoryBuf := memory.UnsafeData(store)

	noop := instance.GetExport(store, "noop_wasm").Func()
	allocate := instance.GetExport(store, "allocate").Func()
	deallocate := instance.GetExport(store, "deallocate").Func()

	inputSize := int32(len(input))
	result, err := allocate.Call(store, inputSize)
	if err != nil {
		log.Panicln(err)
	}

	inputPtr := result.(int32)
	defer deallocate.Call(store, inputPtr, inputSize)

	copy(memoryBuf[inputPtr:], input)

	packedPtrSize, err := noop.Call(store, inputPtr, inputSize)
	if err != nil {
		log.Panicln(err)
	}
	noopResultPtr, noopResultSize := unpackInt64(packedPtrSize.(int64))
	defer deallocate.Call(store, int64(noopResultPtr), int64(noopResultSize))
	// Refresh memoryBuf, after a `.Call` it is invalid
	memoryBuf = memory.UnsafeData(store)

	return string(memoryBuf[noopResultPtr : noopResultPtr+noopResultSize])
}

func main() {
	//runWazero()
	fmt.Println(runWasmWithWasmtime("hello rusty wasmy world"))
	return
	rust := flag.Bool("rust", false, "use rust")
	noopRust := flag.Bool("nooprust", false, "use no-op rust")
	noopGo := flag.Bool("noopgo", false, "use no-op go")
	vrl := flag.Bool("vrl", false, "use vrl")
	stdout := flag.Bool("stdout", false, "Output to stdout")
	useBloblang := flag.Bool("bloblang", false, "use bloblang")
	useUds := flag.Bool("uds", false, "accept data from UDS")
	flag.Parse()

	var reader *bufio.Reader
	if *useUds {
		reader = getUdsReader()
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	var exe *bloblang.Executor
	if *useBloblang {
		exe = setupBloblang()
	}

	var output OutFunc
	if *stdout {
		output = fmt.Println
	} else {
		throughputRecorder := throughputRecorder{}
		output = getBlackholeWriter(&throughputRecorder)
		go func() {
			oneSecond, err := time.ParseDuration("1s")
			if err != nil {
				panic(err)
			}

			for {
				time.Sleep(oneSecond)
				fmt.Println(throughputRecorder.AvgThroughput())
			}
		}()
	}

	for {
		text, _ := reader.ReadString('\n')
		if *rust {
			output(processStringRs(text))
		} else if *vrl {
			text = strings.TrimSpace(text)
			text = fmt.Sprintf("{\"message\":\"%s\"}", text)
			output(processStringVrl(text))
		} else if *useBloblang {
			output(processStringBloblang(exe, text))
		} else if *noopRust {
			output(noopStringRs(text))
		} else if *noopGo {
			output(simpleStringGo(text))
		} else {
			output(processStringGo(text))
		}

		runtime.Gosched()
	}
}

func noopStringRs(str string) string {
	cs := C.CString(str)
	b := C.noop(cs)
	s := C.GoString(b)
	defer C.free(unsafe.Pointer(cs))
	defer C.free(unsafe.Pointer(b))
	return s
}

func processStringRs(str string) string {
	cs := C.CString(str)
	b := C.transform(cs)
	s := C.GoString(b)
	defer C.free(unsafe.Pointer(cs))
	defer C.free(unsafe.Pointer(b))
	return s
}

func processStringVrl(str string) string {
	cs := C.CString(str)
	b := C.transform_vrl(cs)
	s := C.GoString(b)
	defer C.free(unsafe.Pointer(cs))
	defer C.free(unsafe.Pointer(b))
	return s
}

var r = regexp.MustCompile(`\b\w{4}\b`)

func processStringGo(str string) string {
	return r.ReplaceAllString(str, "gogo")
}

// Copy a string
func simpleStringGo(str string) string {
	return strings.Clone(str)
}
