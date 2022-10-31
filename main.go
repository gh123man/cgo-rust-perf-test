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
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"go.uber.org/atomic"
)

const (
	sockAddr = "/tmp/cgo.sock"
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
var greetWasm []byte

func runWasmWithWazero(input string) string {
	// Choose the context to use for function calls.
	ctx := context.Background()

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx) // This closes everything this Runtime created.
	fmt.Println("Wazero 1")

	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	fmt.Println("Wazero 1.5")
	// Instantiate a WebAssembly module that exports
	// "allocate", "deallocate" and "noop_wasm"
	mod, err := r.InstantiateModuleFromBinary(ctx, greetWasm)
	if err != nil {
		log.Panicln(err)
	}

	fmt.Println("Wazero 2")
	// Get references to WebAssembly functions
	noop := mod.ExportedFunction("noop_wasm")
	allocate := mod.ExportedFunction("allocate")
	deallocate := mod.ExportedFunction("deallocate")
	fmt.Println("Wazero 3")

	inputSize := uint64(len(input))

	// Instead of an arbitrary memory offset, use Rust's allocator. Notice
	// there is nothing string-specific in this allocation function. The same
	// function could be used to pass binary serialized data to Wasm.
	results, err := allocate.Call(ctx, inputSize)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println("Wazero 4")

	inputPtr := results[0]
	// This pointer was allocated by Rust, but owned by Go, So, we have to
	// deallocate it when finished
	defer deallocate.Call(ctx, inputPtr, inputSize)

	fmt.Println("Wazero 5")
	// The pointer is a linear memory offset, which is where we write the input string.
	if !mod.Memory().Write(ctx, uint32(inputPtr), []byte(input)) {
		log.Panicf("Memory.Write(%d, %d) out of range of memory size %d",
			inputPtr, inputSize, mod.Memory().Size(ctx))
	}

	fmt.Println("Wazero 6")
	// Invoke 'noop' passing in the pointer+size of the input string
	// Result is a packed ptr+size of a rust-allocated string
	packedPtrSize, err := noop.Call(ctx, inputPtr, inputSize)
	if err != nil {
		log.Panicln(err)
	}
	noopResultPtr := uint32(packedPtrSize[0] >> 32)
	noopResultSize := uint32(packedPtrSize[0])
	// This pointer was allocated by Rust, but owned by Go, So, we have to
	// deallocate it when finished
	defer deallocate.Call(ctx, uint64(noopResultPtr), uint64(noopResultSize))
	fmt.Println("Wazero 7")

	// The pointer is a linear memory offset, which is where we write the input string.
	resultStringBytes, ok := mod.Memory().Read(ctx, noopResultPtr, noopResultSize)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range of memory size %d",
			noopResultPtr, noopResultSize, mod.Memory().Size(ctx))
	}
	fmt.Println("Wazero 8")
	res := string(resultStringBytes)
	return res
}

/*
func runWasmWithWasmer() {
	// Wasmer-go uses cgo bindings to talk to wasmer which is the core runtime
	// wasmer-go currently has disabled aarch64,linux support :(
	// Giving up on wasmer as this is a supported environment
	wasmBytes, _ := ioutil.ReadFile("./target/wasm32-wasi/debug/helloRust.wasm")

	store := wasmer.NewStore(wasmer.NewEngine())
	module, _ := wasmer.NewModule(store, wasmBytes)

	importObject := wasmer.NewImportObject()

	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		panic(err)
	}

	start, err := instance.Exports.GetWasiStartFunction()
	if err != nil {
		panic(err)
	}
	start()

	HelloWorld, err := instance.Exports.GetFunction("noop")
	if err != nil {
		panic(err)
	}
	result, _ := HelloWorld()
	fmt.Println(result)
}
*/

func runWasmWithWasmtime() {
	// Wasmtime has stripped support for 'interface types' which is a proposal
	// for WASM that defines how non-primitive data-types can be passed/returned
	// from wasm.
	// As a result, the below example will not work.
	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)
	module, err := wasmtime.NewModuleFromFile(engine, "./target/wasm32-wasi/debug/helloRust.wasm")
	if err != nil {
		panic(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		panic(err)
	}

	noop := instance.GetExport(store, "noop").Func()
	val, err := noop.Call(store, "hello world")
	if err != nil {
		panic(err)
	}
	fmt.Printf("noop(\"hello world\") = %v\n", val.(string))
}

func main() {
	fmt.Println("Result from wazero wasm call:", runWasmWithWazero("Hello wasm world"))
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
