package zmage

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

func Fmt() error {
	pkgs, err := List(build.Default)
	if err != nil {
		return err
	}

	var files []string
	for _, pkg := range pkgs {
		fileLists := [][]string{
			pkg.GoFiles,
			pkg.CgoFiles,
			pkg.TestGoFiles,
			pkg.XTestGoFiles,
		}

		for _, list := range fileLists {
			for _, file := range list {
				files = append(files, filepath.Join(pkg.Dir, file))
			}
		}
	}

	s, err := sh.Output("gofmt", append([]string{"-l", "-s"}, files...)...)

	if err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if s != "" {
		files = strings.Split(s, "\n")
		for i, file := range files {
			if files[i], err = filepath.Rel(pwd, file); err != nil {
				return err
			}
			files[i] = "./" + files[i]
		}

		fmt.Fprintln(os.Stderr, "The following files need `gofmt -s`:")
		fmt.Fprintln(os.Stderr, strings.Join(files, "\n"))
		return fmt.Errorf("improperly formatted go files")
	}

	return nil
}
