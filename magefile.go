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

var Default = All

// Helpers
//====

// split a single string command into a list of string
func split(cmd string) (string, []string) {
	var args = strings.Fields(cmd)
	return args[0], args[1:]
}

// execute a command with output on stdout
func run(cmd string) error {
	var args = strings.Fields(cmd)
	return sh.RunV(args[0], args[1:]...)
}

// execute a command with an OK/KO message
func runOK(cmd string, ok string, ko string) error {
	err := run(cmd)
	if err != nil {
		return exit(ko)
	}
	success(ok)
	return nil
}

// silently get the output of a command
func output(cmd string) (string, error) {
	var exe, args = split(cmd)
	return sh.Output(exe, args...)
}

// display a success message
func success(str string) {
	fmt.Printf("%s %s \n", Green("✔"), str)
}

// display a failure message
func failure(str string) {
	fmt.Printf("%s %s \n", Red("✖"), str)
}

// exit with an error message and exit code -1
func exit(str string) error {
	return mg.Fatalf(-1, "%s %s", Red("✖"), str)
}

// filter a list of strings using a test function
func grep(list []string, test func(string) bool) []string {
	out := make([]string, 0)
	for _, line := range list {
		if test(line) {
			out = append(out, line)
		}
	}
	return out
}

// Tasks
//====

// Clean the workdir
func Clean() error {
	return run("go clean")
}

// Run the test suite
func Test() error {
	return runOK("gotest -race ./... -tags integration", "Tests succeed", "Tests failed")
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

// Run tests with coverage
func Cover() error {
	cmd := `gotest -race ./... -tags integration -cover ` +
		`-coverpkg=%s -covermode=atomic -coverprofile=coverage.out`
	return runOK(
		fmt.Sprintf(cmd, _coverPkg()),
		"Tests (with coverage) succeed",
		"Tests (with coverage) failed",
	)
}

// Run tests with coverage and generate an HTML report
func CoverHtml() error {
	mg.Deps(Cover)
	return runOK(
		`go tool cover -html=coverage.out -o coverage.html`,
		`Coverage report generated in coverage.html`,
		`Coverage report generation failed`,
	)
}

// Execute static analysis
func Lint() error {
	return runOK("golangci-lint run", "Code is fine", "There is some lints to fix 👆")
}

// Execute the benchmark suite
func Bench() error {
	return runOK(
		`gotest -v -bench . -cpu 1,2,4 -tags bench -run Benchmark`,
		`Benchmark done`,
		`Benchmark failed`,
	)
}

// Compile code
func Build() error {
	return runOK(`go build -v`, `Build success`, `Build failed`)
}

// Lint, Build, Test
func All() error {
	mg.Deps(Lint)
	mg.Deps(Build)
	mg.Deps(Test)
	return nil
}
