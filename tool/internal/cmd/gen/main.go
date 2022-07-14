package main

import (
	"context"

	"github.com/go-courier/logr"

	"github.com/octohelm/gengo/pkg/gengo"

	_ "github.com/octohelm/storage/devpkg/enumgen"
	_ "github.com/octohelm/storage/devpkg/tablegen"
)

func main() {
	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint: []string{
			"github.com/octohelm/storage/testdata/model",
		},
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		panic(err)
	}

	ctx := logr.WithLogger(context.Background(), logr.StdLogger())
	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
