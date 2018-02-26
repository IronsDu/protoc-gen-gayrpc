package main

import (
	"io/ioutil"
	"os"
	"github.com/golang/protobuf/proto"
	"github.com/IronsDu/protoc-gen-gayrpc/generator"
)

func main() {
	g := generator.NewGenerator("protoc-gen-gayrpc")

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	g.WrapTypes()
	g.CommandLineParameters(g.Request.GetParameter())

	g.GenerateAllFiles()

	data, err = proto.Marshal(g.Response)
	if err != nil {
		g.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		g.Error(err, "failed to write output proto")
	}
}
