package main

//#cgo CFLAGS: -I${SRCDIR}/target/release/
//#cgo LDFLAGS: -L${SRCDIR}/target/aarch64-unknown-linux-gnu/release -L${SRCDIR}/target/aarch64-apple-darwin/release -Wl,-rpath,${SRCDIR}/target/aarch64-unknown-linux-gnu/release -Wl,-rpath,${SRCDIR}/target/aarch64-apple-darwin/release -lhelloRust
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
	"github.com/dustin/go-humanize"
	"go.uber.org/atomic"
)

const (
	sockAddr     = "/tmp/cgo.sock"
	wasmPageSize = 65536
	bufSize      = 2048
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

func main() {
	useRust := flag.Bool("rust", false, "use rust")
	useVrl := flag.Bool("vrl", false, "use vrl")
	useRustNoop := flag.Bool("nooprust", false, "use no-op rust")
	useGoNoop := flag.Bool("noopgo", false, "use no-op go")
	useWazeroNoop := flag.Bool("noopwazero", false, "use no-op wasm via wazero runtime")
	useWazero := flag.Bool("wazero", false, "use vrl running insido wazero")
	useWazeroRegex := flag.Bool("regexwazero", false, "use raw regex running insido wazero")
	useWasmtimeNoop := flag.Bool("noopwasmtime", false, "use no-op wasm via wasmtime runtime")
	useWasmtime := flag.Bool("wasmtime", false, "use vrl running inside wasmtime")
	useWasmtimeRegex := flag.Bool("regexwasmtime", false, "use raw regex running inside wasmtime")
	useBloblang := flag.Bool("bloblang", false, "use bloblang")

	// misc
	stdout := flag.Bool("stdout", false, "Output to stdout")
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

	var wazeroRunner *WazeroRunner
	if *useWazeroNoop || *useWazero || *useWazeroRegex {
		// Choose the context to use for function calls.
		ctx := context.Background()

		wazeroRunner = NewWazeroRunner(ctx, compiledWasmBytes)
		defer wazeroRunner.Close() // This closes everything this Runtime created.
	}

	var wasmtimeRunner *WasmtimeRunner
	if *useWasmtimeNoop || *useWasmtime || *useWasmtimeRegex {
		wasmtimeRunner = NewWasmtimeRunner(compiledWasmBytes)
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
		if *useRust {
			output(processStringRs(text))
		} else if *useVrl {
			text = strings.TrimSpace(text)
			text = fmt.Sprintf("{\"message\":\"%s\"}", text)
			output(processStringVrl(text))
		} else if *useBloblang {
			output(processStringBloblang(exe, text))
		} else if *useRustNoop {
			output(noopStringRs(text))
		} else if *useGoNoop {
			output(simpleStringGo(text))
		} else if *useWazeroNoop {
			output(wazeroRunner.runNoop(text))
		} else if *useWazeroRegex {
			output(wazeroRunner.runRegex(text))
		} else if *useWazero {
			output(wazeroRunner.runVrl(text))
		} else if *useWasmtimeNoop {
			output(wasmtimeRunner.runNoop(text))
		} else if *useWasmtime {
			output(wasmtimeRunner.runVrl(text))
		} else if *useWasmtimeRegex {
			output(wasmtimeRunner.runRegex(text))
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
