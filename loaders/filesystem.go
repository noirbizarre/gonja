package loaders

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// FilesystemLoader represents a local filesystem loader with basic
// BaseDirectory capabilities. The access to the local filesystem is unrestricted.
type FilesystemLoader struct {
	root string
}

// MustNewFileSystemLoader creates a new FilesystemLoader instance
// and panics if there's any error during instantiation. The parameters
// are the same like NewFileSystemLoader.
func MustNewFileSystemLoader(root string) *FilesystemLoader {
	fs, err := NewFileSystemLoader(root)
	if err != nil {
		log.Panic(err)
	}
	return fs
}

// NewFileSystemLoader creates a new FilesystemLoader and allows
// templatesto be loaded from disk (unrestricted). If any base directory
// is given (or being set using SetBaseDir), this base directory is being used
// for path calculation in template inclusions/imports. Otherwise the path
// is calculated based relatively to the including template's path.
func NewFileSystemLoader(root string) (*FilesystemLoader, error) {
	fs := &FilesystemLoader{}
	if root != "" {
		if err := fs.SetBaseDir(root); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

// SetBaseDir sets the template's base directory. This directory will
// be used for any relative path in filters, tags and From*-functions to determine
// your template. See the comment for NewFileSystemLoader as well.
func (fs *FilesystemLoader) SetBaseDir(path string) error {
	// Make the path absolute
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		path = abs
	}

	// Check for existence
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.Errorf("The given path '%s' is not a directory.", path)
	}

	fs.root = path
	return nil
}

// Get reads the path's content from your local filesystem.
func (fs *FilesystemLoader) Get(path string) (io.Reader, error) {
	realPath, err := fs.Path(path)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadFile(realPath)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

// Path resolves a filename relative to the base directory. Absolute paths are allowed.
// When there's no base dir set, the absolute path to the filename
// will be calculated based on either the provided base directory (which
// might be a path of a template which includes another template) or
// the current working directory.
func (fs *FilesystemLoader) Path(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}

	// root := fs.root
	if fs.root == "" {
		root, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(root, name), nil
	} else {
		return filepath.Join(fs.root, name), nil
	}

	// // Our own base dir has always priority; if there's none
	// // we use the path provided in base.
	// var err error
	// if fs.root == "" {
	// 	if base == "" {
	// 		base, err = os.Getwd()
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		return filepath.Join(base, name)
	// 	}

	// 	return filepath.Join(filepath.Dir(base), name)
	// }

	// return filepath.Join(fs.root, name)
}

// SandboxedFilesystemLoader is still WIP.
type SandboxedFilesystemLoader struct {
	*FilesystemLoader
}

// NewSandboxedFilesystemLoader creates a new sandboxed local file system instance.
func NewSandboxedFilesystemLoader(root string) (*SandboxedFilesystemLoader, error) {
	fs, err := NewFileSystemLoader(root)
	if err != nil {
		return nil, err
	}
	return &SandboxedFilesystemLoader{
		FilesystemLoader: fs,
	}, nil
}

// Move sandbox to a virtual fs

/*
if len(set.SandboxDirectories) > 0 {
    defer func() {
        // Remove any ".." or other crap
        resolvedPath = filepath.Clean(resolvedPath)

        // Make the path absolute
        absPath, err := filepath.Abs(resolvedPath)
        if err != nil {
            panic(err)
        }
        resolvedPath = absPath

        // Check against the sandbox directories (once one pattern matches, we're done and can allow it)
        for _, pattern := range set.SandboxDirectories {
            matched, err := filepath.Match(pattern, resolvedPath)
            if err != nil {
                panic("Wrong sandbox directory match pattern (see http://golang.org/pkg/path/filepath/#Match).")
            }
            if matched {
                // OK!
                return
            }
        }

        // No pattern matched, we have to log+deny the request
        set.logf("Access attempt outside of the sandbox directories (blocked): '%s'", resolvedPath)
        resolvedPath = ""
    }()
}
*/
