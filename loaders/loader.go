package loaders

import (
	"io"
)

// TemplateLoader allows to implement a virtual file system.
type Loader interface {
	// Abs calculates the path to a given template. Whenever a path must be resolved
	// due to an import from another template, the base equals the parent template's path.
	// Abs(base, name string) string

	// Get returns an io.Reader where the template's content can be read from.
	Get(path string) (io.Reader, error)

	// Resolve the given path in the current context
	Path(path string) (string, error)
}
