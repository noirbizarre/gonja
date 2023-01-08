package gonja_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/paradime-io/gonja"
	"github.com/paradime-io/gonja/config"
	"github.com/pmezard/go-difflib/difflib"

	tu "github.com/paradime-io/gonja/testutils"
)

var testCases = []struct {
	name                  string
	trim_blocks           bool
	lstrip_blocks         bool
	keep_trailing_newline bool
}{
	{"default", false, false, false},
	{"trim_blocks", true, false, false},
	{"lstrip_blocks", false, true, false},
	{"keep_trailing_newline", false, false, true},
	{"all", true, true, true},
}

const source = "testData/whitespaces/source.tpl"
const result = "testData/whitespaces/%s.out"

func TestWhiteSpace(t *testing.T) {
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			cfg := config.NewConfig()
			env := gonja.NewEnvironment(cfg, gonja.DefaultLoader)
			env.TrimBlocks = test.trim_blocks
			env.LstripBlocks = test.lstrip_blocks
			env.KeepTrailingNewline = test.keep_trailing_newline

			tpl, err := env.FromFile(source)
			if err != nil {
				t.Fatalf("Error on FromFile('%s'): %s", source, err.Error())
			}
			output := fmt.Sprintf(result, test.name)
			expected, rerr := os.ReadFile(output)
			if rerr != nil {
				t.Fatalf("Error on ReadFile('%s'): %s", output, rerr.Error())
			}
			rendered, err := tpl.ExecuteBytes(tu.Fixtures)
			if err != nil {
				t.Fatalf("Error on Execute('%s'): %s", source, err.Error())
			}
			// rendered = testTemplateFixes.fixIfNeeded(match, rendered)
			if bytes.Compare(expected, rendered) != 0 {
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(expected)),
					B:        difflib.SplitLines(string(rendered)),
					FromFile: "Expected",
					ToFile:   "Rendered",
					Context:  2,
					Eol:      "\n",
				}
				result, _ := difflib.GetUnifiedDiffString(diff)
				t.Errorf("%s rendered with diff:\n%v", source, result)
			}
		})
	}
}
