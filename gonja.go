package gonja

import (
	"github.com/noirbizarre/gonja/config"
	"github.com/noirbizarre/gonja/loaders"
)

var (
	// DefaultLoader is being used by the DefaultSet.
	DefaultLoader = loaders.MustNewFileSystemLoader("")

	// DefaultEnv is an environment created for quick/standalone template rendering.
	DefaultEnv = NewEnvironment(config.DefaultConfig, DefaultLoader)

	// Methods on the default set
	FromString = DefaultEnv.FromString
	FromBytes  = DefaultEnv.FromBytes
	FromFile   = DefaultEnv.FromFile
	FromCache  = DefaultEnv.FromCache

	// Globals for the default set
	Globals = DefaultEnv.Globals
)
