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

func main() {

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
