package zmage

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

func installFile(ctx build.Context, dir string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	pkg, err := ctx.Import(dir, pwd, 0)
	if err != nil {
		return "", err
	}

	return filepath.Join(pkg.BinDir, filepath.Base(dir)), nil
}

func ctxToEnv(ctx build.Context) map[string]string {
	env := map[string]string{
		"GOARCH": ctx.GOARCH,
		"GOOS":   ctx.GOOS,
		"GOROOT": ctx.GOROOT,
		"GOPATH": ctx.GOPATH,
	}

	if ctx.CgoEnabled {
		env["CGO_ENABLED"] = "1"
	}

	return env
}

func goBuild(ctx build.Context, args ...string) error {
	ld, err := ldFlags()
	if err != nil {
		return err
	}

	args = append([]string{"build", "-i", "-ldflags", ld}, args...)
	return sh.RunWith(ctxToEnv(ctx), goexe, args...)
}

func buildSources(ctx build.Context, dir string) ([]string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	pkg, err := ctx.Import(dir, pwd, 0)
	if err != nil {
		return nil, err
	}

	var sources []string
	cache := map[string]*build.Package{
		pkg.ImportPath: pkg,
	}
	todo := []*build.Package{pkg}

	for len(todo) > 0 {
		p := todo[0]
		todo = todo[1:]

		for _, files := range [][]string{
			p.GoFiles,
			p.CgoFiles,
		} {
			for _, f := range files {
				sources = append(sources, filepath.Join(p.Dir, f))
			}
		}

		for _, i := range p.Imports {
			if !strings.Contains(i, ".") {
				continue
			}

			if _, ok := cache[i]; ok {
				continue
			}

			t, err := ctx.Import(i, pwd, 0)
			if err != nil {
				return nil, err
			}

			cache[i] = t
			todo = append(todo, t)
		}
	}

	return sources, nil
}

func shouldBuild(ctx build.Context, dir, file string) (bool, error) {
	files, err := buildSources(ctx, dir)
	if err != nil {
		return false, err
	}

	if !filepath.IsAbs(file) {
		pwd, err := os.Getwd()
		if err != nil {
			return false, err
		}

		file = filepath.Join(pwd, file)
	}

	if _, err = os.Stat(file); os.IsNotExist(err) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	modified, err := Modified(file, files...)
	if err != nil {
		return false, err
	}

	return modified, nil
}

func Build(ctx build.Context, pkg, target string, args ...string) error {
	ok, err := shouldBuild(ctx, pkg, target)
	if !ok || err != nil {
		return err
	}

	args = append(args, "-o", target, pkg)
	return goBuild(ctx, args...)
}

func Install(ctx build.Context, pkg string, args ...string) error {
	file, err := installFile(ctx, pkg)
	if err != nil {
		return err
	}

	ok, err := shouldBuild(ctx, file, pkg)
	if !ok || err != nil {
		return err
	}

	ld, err := ldFlags()
	if err != nil {
		return err
	}

	args = append([]string{"install", "-v", "-ldflags", ld}, args...)
	args = append(args, pkg)

	return sh.RunWith(ctxToEnv(ctx), goexe, args...)
}
