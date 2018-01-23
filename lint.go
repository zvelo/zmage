package zmage

import (
	"context"
	"go/build"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, Vet)

	pkgs, err := List(build.Default)
	if err != nil {
		return err
	}

	var pkgNames []string
	for _, pkg := range pkgs {
		pkgNames = append(pkgNames, pkg.ImportPath)
	}

	_, err = sh.Exec(nil, os.Stderr, nil, "golint", pkgNames...)
	return err
}
