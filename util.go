package zmage

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

var (
	goexe  = "go"
	gotest = "go test"
	docker = sh.RunCmd("docker")
	env    = map[string]string{"GODEBUG": "cgocheck=2"}
)

func init() {
	if exe := os.Getenv("GOEXE"); exe != "" {
		goexe = exe
		gotest = goexe + " test"
	}

	if exe := os.Getenv("GOTEST"); exe != "" {
		gotest = exe
	}
}

type onceData struct {
	data string
	err  error
	once sync.Once
}

func ldFlags() (string, error) {
	v, err := version()
	if err != nil {
		return "", err
	}

	hash, err := commitHash()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"-X main.version=%s -X main.gitCommit=%s -X main.buildDate=%s",
		v, hash, buildDate(),
	), nil
}

var versionData onceData

func version() (string, error) {
	versionData.once.Do(func() {
		versionData.data, versionData.err = sh.Output("git", "describe", "--tags", "--always", "--dirty=\"-dev\"")
	})
	return versionData.data, versionData.err
}

var commitHashData onceData

func commitHash() (string, error) {
	commitHashData.once.Do(func() {
		commitHashData.data, commitHashData.err = sh.Output("git", "rev-parse", "--short", "HEAD")
	})
	return commitHashData.data, commitHashData.err
}

var buildDateData onceData

func buildDate() string {
	buildDateData.once.Do(func() {
		buildDateData.data = time.Now().Format("2006-01-02T15:14:05Z")
	})
	return buildDateData.data
}

var branchData onceData

func branch() (string, error) {
	branchData.once.Do(func() {
		branchData.data, branchData.err = sh.Output("git", "symbolic-ref", "--short", "-q", "HEAD")
	})
	return branchData.data, branchData.err
}

func installTestDeps(flags ...string) error {
	args := append([]string{"test", "-i"}, flags...)
	args = append(args, "./...")

	return sh.RunWith(env, goexe, args...)
}

func Vet() error {
	return sh.RunWith(env, goexe, "vet", "./...")
}

func Modified(file string, files ...string) (bool, error) {
	modified, err := target.Path(file, files...)
	if os.IsNotExist(err) {
		return true, nil
	}

	return modified, err
}

func Touch(file string) error {
	return sh.Run("touch", file)
}

func Vendor() error {
	if err := sh.Run("dep", "ensure"); err != nil {
		return err
	}
	return sh.Run("dep", "prune")
}

func Clean(files ...string) error {
	files = append(files,
		"./.image-stamp",
		"./.coverage.out",
		"./.coverage-all.out",
	)

	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}

	return nil
}
