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

////
// Tasks
////

func Test() error {
	return run("gotest -race -v ./... -tags integration")
}

func Cover() error {
	// COVER_PKGS=$(go list ./... | grep -v testutils | tr '\n' ',')
	// gotest -race -v ./... -tags integration -cover -coverpkg=$COVER_PKGS -covermode=atomic
	return run("gotest -race -v ./... -tags integration")
}

func CoverHtml() error {
	// COVER_PKGS=$(go list ./... | grep -v testutils | tr '\n' ',')
	// gotest -race -v ./... -tags integration -coverpkg=$COVER_PKGS -covermode=atomic -coverprofile=coverage.out
	// go tool cover -html=coverage.out -o coverage.html
	return run("gotest -race -v ./... -tags integration")
}

func Bench() error {
	return run("gotest -v -bench . -cpu 1,2,4 -tags bench -run Benchmark")
}

func Build() error {
	err := run("go build")
	if err != nil {
		return exit("Build failed")
	}
	success("Build success")
	return nil
}
