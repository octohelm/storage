package main

import (
	"context"
	"os"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	"github.com/octohelm/gengo/pkg/gengo"

	_ "github.com/octohelm/enumeration/devpkg/enumgen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
	_ "github.com/octohelm/storage/deprecated/devpkg/filtergen"
	_ "github.com/octohelm/storage/devpkg/filteropgen"
	_ "github.com/octohelm/storage/devpkg/tablegen"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint: []string{
			cwd,
		},
		OutputFileBaseName: "zz_generated",
		Globals: map[string][]string{
			"gengo:runtimedoc": {},
		},
	})
	if err != nil {
		panic(err)
	}

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
