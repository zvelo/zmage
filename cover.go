package zmage

import (
	"bufio"
	"go/build"
	"io"
	"os"

	"github.com/magefile/mage/sh"
)

func appendCoverage(w io.Writer, fileName string) error {
	c, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = c.Close() }()

	r := bufio.NewReader(c)

	// skip the first "mode:" line in the file
	if _, err = r.ReadString('\n'); err != nil {
		return err
	}

	_, err = io.Copy(w, r)
	return err
}

func CoverOnly(flags ...string) error {
	flags = append(flags, "-race")

	if err := installTestDeps(flags...); err != nil {
		return err
	}

	pkgs, err := List(build.Default)
	if err != nil {
		return err
	}

	coverAll, err := os.OpenFile(".coverage-all.out", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = coverAll.Close() }()

	if _, err = coverAll.WriteString("mode: atomic\n"); err != nil {
		return err
	}

	for _, pkg := range pkgs {
		args := []string{"test", "-coverprofile=.coverage.out", "-covermode=atomic"}
		args = append(args, flags...)
		args = append(args, pkg.ImportPath)
		if err = sh.Run(goexe, args...); err != nil {
			return err
		}

		if err = appendCoverage(coverAll, ".coverage.out"); err != nil {
			return err
		}

		if err = os.Remove(".coverage.out"); err != nil {
			return err
		}
	}

	return nil
}

func Cover(flags ...string) error {
	if err := CoverOnly(flags...); err != nil {
		return err
	}

	return sh.Run(goexe, "tool", "cover", "-html=.coverage-all.out")
}
