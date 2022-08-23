package gonja_test

import (
	"fmt"
	"github.com/paradime-io/gonja"
	"github.com/paradime-io/gonja/config"
	"os"
	"path/filepath"
	"testing"
)

func FuzzGonja(f *testing.F) {
	for _, path := range []string{
		"./testData/",
		"./testData/expressions",
		"./testData/filters",
		"./testData/functions",
		"./testData/fuzz",
		"./testData/inheritance",
		"./testData/statements",
		"./testData/tests",
		"./testData/whitespaces",
	} {
		matches, err := filepath.Glob(filepath.Join(path, "*.tpl"))
		fmt.Println(matches)
		if err != nil {
			f.Fatal(err)
		}

		for _, match := range matches {
			fmt.Println(match)
			tpl, err := os.ReadFile(match)
			if err != nil {
				f.Fatalf("Error on FromFile('%s'):\n%s", match, err.Error())
			}
			f.Add(string(tpl))
		}
	}

	f.Fuzz(func(t *testing.T, tpl string) {
		env := gonja.NewEnvironment(config.DefaultConfig, gonja.DefaultLoader)
		_, _ = env.FromString(tpl)
	})
}
