package main

import (
	"context"

	"github.com/go-courier/logr/slog"

	"github.com/go-courier/logr"

	"github.com/octohelm/gengo/pkg/gengo"

	_ "github.com/octohelm/enumeration/devpkg/enumgen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
	_ "github.com/octohelm/storage/devpkg/filtergen"
	_ "github.com/octohelm/storage/devpkg/filteropgen"
	_ "github.com/octohelm/storage/devpkg/tablegen"
)

func main() {
	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint: []string{
			"github.com/octohelm/storage/testdata/model",
			"github.com/octohelm/storage/testdata/model/filter",
			"github.com/octohelm/storage/testdata/model/filter/v2",
			"github.com/octohelm/storage/testdata/model/aggregate",
			"github.com/octohelm/storage/internal/testutil",
			"github.com/octohelm/storage/pkg/dal",
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
