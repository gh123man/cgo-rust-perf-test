package main

//#cgo CFLAGS: -I./target/debug/
//#cgo LDFLAGS: -L./target/debug -lhelloRust
//
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <helloRust.h>
import "C"
import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"unsafe"
)

func main() {

	reader := bufio.NewReader(os.Stdin)

	for {
		text, _ := reader.ReadString('\n')
		// fmt.Print(processString(text))
		fmt.Print(processStringGo(text))
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

var r = regexp.MustCompile(`\b\w{4}\b`)

func processStringGo(str string) string {
	return r.ReplaceAllLiteralString(str, "gogo")
}
