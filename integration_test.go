// +build integration

package gonja_test

import (
	"testing"

	tu "github.com/nikolalohinski/gonja/testutils"
)

func TestTemplates(t *testing.T) {
	// Add a global to the default set
	root := "./testData"
	env := tu.TestEnv(root)
	env.Globals.Set("this_is_a_global_variable", "this is a global text")
	tu.GlobTemplateTests(t, root, env)
}

func TestExpressions(t *testing.T) {
	root := "./testData/expressions"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestFilters(t *testing.T) {
	root := "./testData/filters"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestFunctions(t *testing.T) {
	root := "./testData/functions"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestTests(t *testing.T) {
	root := "./testData/tests"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestStatements(t *testing.T) {
	root := "./testData/statements"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

// func TestCompilationErrors(t *testing.T) {
// 	tu.GlobErrorTests(t, "./testData/errors/compilation")
// }

// func TestExecutionErrors(t *testing.T) {
// 	tu.GlobErrorTests(t, "./testData/errors/execution")
// }
