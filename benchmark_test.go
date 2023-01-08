//go:build bench
// +build bench

package gonja_test

import (
	"io/ioutil"
	"testing"

	"github.com/paradime-io/gonja"

	tu "github.com/paradime-io/gonja/testutils"
)

func BenchmarkFromCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tpl, err := gonja.FromCache("testData/complex.tpl")
		if err != nil {
			b.Fatal(err)
		}
		_, err = tpl.Execute(tu.Fixtures)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFromFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tpl, err := gonja.FromFile("testData/complex.tpl")
		if err != nil {
			b.Fatal(err)
		}
		_, err = tpl.Execute(tu.Fixtures)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecute(b *testing.B) {
	tpl, err := gonja.FromFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = tpl.Execute(tu.Fixtures)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileAndExecute(b *testing.B) {
	buf, err := os.ReadFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	preloadedTpl := string(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tpl, err := gonja.FromString(preloadedTpl)
		if err != nil {
			b.Fatal(err)
		}

		_, err = tpl.Execute(tu.Fixtures)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelExecute(b *testing.B) {
	tpl, err := gonja.FromFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tpl.Execute(tu.Fixtures)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
