//+build mage

package main

import (
	"fmt"
	"strings"

	. "github.com/logrusorgru/aurora"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	packageName = "github.com/noirbizarre/gonja"
)

////
// Helpers
////

func split(cmd string) (string, []string) {
	var args = strings.Fields(cmd)
	return args[0], args[1:]
}

func run(cmd string) error {
	var args = strings.Fields(cmd)
	return sh.RunV(args[0], args[1:]...)
}

func output(cmd string) (string, error) {
	var exe, args = split(cmd)
	return sh.Output(exe, args...)
}

func coverPkgs() string {
	return ""
}

func success(str string) {
	fmt.Printf("%s %s \n", Green("✔"), str)
}

func failure(str string) {
	fmt.Printf("%s %s \n", Red("✖"), str)
}

func exit(str string) error {
	return mg.Fatalf(-1, "%s %s", Red("✖"), str)
}

func grep(list []string, test func(string) bool) []string {
	out := make([]string, 0)
	for _, line := range list {
		if test(line) {
			out = append(out, line)
		}
	}
	return out
}

////
// Tasks
////

func Test() error {
	return run("gotest -race -v ./... -tags integration")
}

func _coverPkg() string {
	out, err := output("go list ./...")
	if err != nil {
		exit(err.Error())
	}
	pkgs := strings.Split(out, "\n")
	pkgs = grep(pkgs, func(line string) bool {
		return !strings.HasSuffix(line, "testutils")
	})
	return strings.Join(pkgs, ",")
}

func Cover() error {
	cmd := `gotest -race -v ./... -tags integration -cover -coverpkg=%s -covermode=atomic`
	return run(fmt.Sprintf(cmd, _coverPkg()))
}

func CoverHtml() error {
	cmd := `gotest -race -v ./... -tags integration -cover ` +
		`-coverpkg=%s -covermode=atomic -coverprofile=coverage.out`

	cmd = fmt.Sprintf(cmd, _coverPkg())
	if err := run(cmd); err != nil {
		return err
	}
	return run(`go tool cover -html=coverage.out -o coverage.html`)
}

func Lint() error {
	return run("golangci-lint run")
}

func Bench() error {
	return run("gotest -v -bench . -cpu 1,2,4 -tags bench -run Benchmark")
}

func Build() error {
	err := run("go build -v")
	if err != nil {
		return exit("Build failed")
	}
	success("Build success")
	return nil
}
