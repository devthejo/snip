package xjsonnet

import (
	"fmt"
	"io/ioutil"
	"os"

	jsonnet "github.com/google/go-jsonnet"
)

func RenderEnv(src string, envMap map[string]string) (string, error) {
	vm := jsonnet.MakeVM()

	cwd, _ := os.Getwd()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{cwd},
	})

	vm.NativeFunction(NativeFunctionParseJson5())
	vm.NativeFunction(NativeFunctionParseBool())
	vm.NativeFunction(NativeFunctionEnv(envMap))
	vm.NativeFunction(NativeFunctionEnvOr(envMap))
	vm.NativeFunction(NativeFunctionEnviron(envMap))

	var bytes []byte
	var err error
	bytes, err = ioutil.ReadFile(src)
	input := string(bytes)

	var output string

	if err != nil {
		var op string
		switch typedErr := err.(type) {
		case *os.PathError:
			op = typedErr.Op
			err = typedErr.Err
		}
		if op == "open" {
			return output, fmt.Errorf("Opening input file: %s: %s\n", src, err.Error())
		} else if op == "read" {
			return output, fmt.Errorf("Reading input file: %s: %s\n", src, err.Error())
		} else {
			return output, err
		}
	}

	output, err = vm.EvaluateSnippet(src, input)

	if err != nil {
		return output, err
	}

	return output, nil

}
