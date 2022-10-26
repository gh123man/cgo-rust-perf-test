package main

//#cgo CFLAGS: -I./target/release/
//#cgo LDFLAGS: -L./target/release -lhelloRust
//
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <helloRust.h>
import "C"
import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unsafe"
)

func main() {

	rust := flag.Bool("rust", false, "use rust")
	vrl := flag.Bool("vrl", false, "use vrl")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	for {
		text, _ := reader.ReadString('\n')
		if *rust {
			fmt.Println(processStringRs(text))
		} else if *vrl {
			text = strings.TrimSpace(text)
			text = fmt.Sprintf("{\"message\":\"%s\"}", text)
			fmt.Println(processStringVrl(text))
		} else {
			fmt.Println(processStringGo(text))
		}
	}
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

func simpleStringRs(str string) string {
	cs := C.CString(str)
	b := C.passthrough(cs)
	s := C.GoString(b)
	defer C.free(unsafe.Pointer(cs))
	defer C.free(unsafe.Pointer(b))
	return s
}

// Copy a string
func simpleStringGo(str string) string {
	return strings.Clone(str)
}
