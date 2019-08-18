package gonja

import (
	"io/ioutil"

	"github.com/goph/emperror"

	"github.com/noirbizarre/gonja/builtins"
	"github.com/noirbizarre/gonja/config"
	"github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/loaders"
)

type Environment struct {
	*exec.EvalConfig
	Loader loaders.Loader
}

func NewEnvironment(cfg *config.Config, loader loaders.Loader) *Environment {
	env := &Environment{
		EvalConfig: exec.NewEvalConfig(cfg),
		Loader:     loader,
	}
	env.EvalConfig.Loader = env
	env.Filters.Update(builtins.Filters)
	env.Statements.Update(builtins.Statements)
	env.Tests.Update(builtins.Tests)
	env.Globals.Merge(builtins.Globals)
	env.Globals.Set("gonja", map[string]interface{}{
		"version": VERSION,
	})
	return env
}

// // FromCache is a convenient method to cache templates. It is thread-safe
// // and will only compile the template associated with a filename once.
// // If Environment.Debug is true (for example during development phase),
// // FromCache() will not cache the template and instead recompile it on any
// // call (to make changes to a template live instantaneously).
// func (env *Environment) FromCache(filename string) (*Template, error) {
// 	if env.Config.Debug {
// 		// Recompile on any request
// 		return env.FromFile(filename)
// 	}
// 	// Cache the template
// 	cleanedFilename := env.resolveFilename(nil, filename)

// 	env.templateCacheMutex.Lock()
// 	defer env.templateCacheMutex.Unlock()

// 	tpl, has := env.templateCache[cleanedFilename]

// 	// Cache miss
// 	if !has {
// 		tpl, err := env.FromFile(cleanedFilename)
// 		if err != nil {
// 			return nil, err
// 		}
// 		env.templateCache[cleanedFilename] = tpl
// 		return tpl, nil
// 	}

// 	// Cache hit
// 	return tpl, nil
// }

// FromString loads a template from string and returns a Template instance.
func (env *Environment) FromString(tpl string) (*exec.Template, error) {
	return exec.NewTemplate("string", tpl, env.EvalConfig)
	// return newTemplateString(env, []byte(tpl))
}

// FromBytes loads a template from bytes and returns a Template instance.
func (env *Environment) FromBytes(tpl []byte) (*exec.Template, error) {
	// return newTemplateString(env, tpl)
	return exec.NewTemplate("bytes", string(tpl), env.EvalConfig)
}

// FromFile loads a template from a filename and returns a Template instance.
func (env *Environment) FromFile(filename string) (*exec.Template, error) {
	// fd, err := env.Loader.Get(env.resolveFilename(nil, filename))
	fd, err := env.Loader.Get(filename)
	if err != nil {
		return nil, emperror.With(err, "filename", filename)
	}
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, emperror.With(err, "filename", filename)
	}

	// return newTemplate(env, filename, false, buf)
	return exec.NewTemplate(filename, string(buf), env.EvalConfig)
}

func (env *Environment) GetTemplate(filename string) (*exec.Template, error) {
	return env.FromFile(filename)
}

// func (env *Environment) ParseTemplate(filename string) (*nodes.Template, error) {
// 	tpl, err := env.FromFile(filename)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, `Unable to parse template "%s"`, filename)
// 	}
// 	return tpl.Root, nil
// }
