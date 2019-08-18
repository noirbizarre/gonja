// +build bench

package gonja_test

import (
	"io/ioutil"

	"testing"

	"github.com/noirbizarre/gonja"
)

func BenchmarkCache(b *testing.B) {
	cacheSet := gonja.NewSet("cache set", gonja.MustNewLocalFileSystemLoader(""))
	for i := 0; i < b.N; i++ {
		tpl, err := cacheSet.FromCache("testData/complex.tpl")
		if err != nil {
			b.Fatal(err)
		}
		err = tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheDebugOn(b *testing.B) {
	cacheDebugSet := gonja.NewSet("cache set", gonja.MustNewLocalFileSystemLoader(""))
	cacheDebugSet.Debug = true
	for i := 0; i < b.N; i++ {
		tpl, err := cacheDebugSet.FromFile("testData/complex.tpl")
		if err != nil {
			b.Fatal(err)
		}
		err = tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecuteComplexWithSandboxActive(b *testing.B) {
	tpl, err := gonja.FromFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileAndExecuteComplexWithSandboxActive(b *testing.B) {
	buf, err := ioutil.ReadFile("testData/complex.tpl")
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

		err = tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelExecuteComplexWithSandboxActive(b *testing.B) {
	tpl, err := gonja.FromFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkExecuteComplexWithoutSandbox(b *testing.B) {
	s := gonja.NewSet("set without sandbox", gonja.MustNewLocalFileSystemLoader(""))
	tpl, err := s.FromFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileAndExecuteComplexWithoutSandbox(b *testing.B) {
	buf, err := ioutil.ReadFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	preloadedTpl := string(buf)

	s := gonja.NewSet("set without sandbox", gonja.MustNewLocalFileSystemLoader(""))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tpl, err := s.FromString(preloadedTpl)
		if err != nil {
			b.Fatal(err)
		}

		err = tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelExecuteComplexWithoutSandbox(b *testing.B) {
	s := gonja.NewSet("set without sandbox", gonja.MustNewLocalFileSystemLoader(""))
	tpl, err := s.FromFile("testData/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := tpl.ExecuteWriterUnbuffered(tplContext, ioutil.Discard)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
