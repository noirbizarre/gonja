package gonja

import (
	"github.com/noirbizarre/gonja/config"
	// "github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/loaders"
)


var (
	// debug  bool // internal debugging
	// logger = log.New(os.Stdout, "[gonja] ", log.LstdFlags|log.Lshortfile)

	// DefaultLoader allows the default un-sandboxed access to the local file
	// system and is being used by the DefaultSet.
	DefaultLoader = loaders.MustNewFileSystemLoader("")

	// DefaultEnv is an environment created for quick/standalone template rendering.
	DefaultEnv = NewEnvironment(config.DefaultConfig, DefaultLoader)

	// Methods on the default set
	FromString = DefaultEnv.FromString
	FromBytes  = DefaultEnv.FromBytes
	FromFile   = DefaultEnv.FromFile
	// FromCache            = DefaultEnv.FromCache
	// RenderTemplateString = DefaultEnv.RenderTemplateString
	// RenderTemplateFile   = DefaultEnv.RenderTemplateFile

	// Globals for the default set
	Globals = DefaultEnv.Globals
	// Context = exec.Context
)

// // Must panics, if a Template couldn't successfully parsed. This is how you
// // would use it:
// //     var baseTemplate = gonja.Must(gonja.FromFile("templates/base.html"))
// func Must(tpl *Template, err error) *Template {
// 	if err != nil {
// 		panic(err)
// 	}
// 	return tpl
// }

