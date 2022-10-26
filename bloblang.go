package main

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
)

func setupBloblang() *bloblang.Executor {
	env := bloblang.NewEnvironment().WithoutFunctions("env", "file")

	mapping := `
root = this.re_replace_all("\\b\\w{4}\\b", "gogo")
`

	exe, err := env.Parse(mapping)
	if err != nil {
		panic(err)
	}

	return exe
}

func processStringBloblang(exe *bloblang.Executor, text string) string {
	res, err := exe.Query(text)
	if err != nil {
		panic(err)
	}

	return res.(string)
}
