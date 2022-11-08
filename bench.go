package main

import (
	"context"
	"fmt"
	"log"
	"strings"
)

const (
	BenchmarkRuns  = 100_000
	BenchmarkInput = "Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern"
)

type StringInStringOut func(in string) string

type Scenario struct {
	environment string
	description string
	runner      StringInStringOut
	result      string
}

func generateBenchmarkTable() string {
	wazeroCtx := context.Background()
	wazeroRunner := NewWazeroRunner(wazeroCtx, compiledWasmBytes)
	defer wazeroRunner.Close()

	wasmtimeRunner := NewWasmtimeRunner(compiledWasmBytes)

	// Step 1, generate the scenarios that we want to run
	// - processStringRs, processStringGo, useVrl
	scenarios := []*Scenario{
		// String Copy
		{"Go", "String Copy", simpleStringGo, ""},
		{"Rust (FFI)", "String Copy", noopStringRs, ""},
		{"Rust (WASM Wazero)", "String Copy", func(s string) string { return wazeroRunner.runNoop(s) }, ""},
		{"Rust (WASM Wasmtime)", "String Copy", func(s string) string { return wasmtimeRunner.runNoop(s) }, ""},

		// Regex
		{"Go", "Regex Substitution", processStringGo, ""},
		{"Rust (FFI)", "Regex Substitution", processStringRs, ""},
		{"Rust (WASM Wazero)", "Regex Substitution", func(s string) string { return wazeroRunner.runRegex(s) }, ""},
		{"Rust (WASM Wasmtime)", "Regex Substitution", func(s string) string { return wasmtimeRunner.runRegex(s) }, ""},

		// VRL
		{"Rust (FFI)", "VRL Replace", processStringVrl, ""},
		{"Rust (WASM Wazero)", "VRL Replace", func(s string) string { return wazeroRunner.runVrl(s) }, ""},
		{"Rust (WASM Wasmtime)", "VRL Replace", func(s string) string { return wasmtimeRunner.runVrl(s) }, ""},
	}

	// Step 2, run each one for N amount of logs and grab average throughput
	// from throughput recorder

	for _, scenario := range scenarios {
		throughputRecorder := throughputRecorder{}
		outputFn := getBlackholeWriter(&throughputRecorder)

		// TODO switch this to a time-based run maybe?
		for i := 0; i < BenchmarkRuns; i++ {
			outputFn(scenario.runner(BenchmarkInput))
		}

		scenario.result = throughputRecorder.AvgThroughput()
		log.Printf("Scenario %q %q finished with result: %s", scenario.environment, scenario.description, scenario.result)
	}

	// Step 3, construct markdown table with this data
	var b strings.Builder
	fmt.Fprintf(&b, "| Execution Environment | Scenario | Result |\n")
	fmt.Fprintf(&b, "| --------------------- | -------- | ------ |\n")
	for _, scenario := range scenarios {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", scenario.environment, scenario.description, scenario.result)
	}

	return b.String()
}
