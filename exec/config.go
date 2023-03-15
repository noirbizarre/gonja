package exec

import (
	"github.com/pkg/errors"

	"github.com/nikolalohinski/gonja/config"
	"github.com/nikolalohinski/gonja/nodes"
)

type EvalConfig struct {
	*config.Config
	Filters    *FilterSet
	Globals    *Context
	Statements *StatementSet
	Tests      *TestSet
	Loader     TemplateLoader
}

func NewEvalConfig(cfg *config.Config) *EvalConfig {
	return &EvalConfig{
		Config:     cfg,
		Globals:    EmptyContext(),
		Filters:    &FilterSet{},
		Statements: &StatementSet{},
		Tests:      &TestSet{},
	}
}

func (cfg *EvalConfig) Inherit() *EvalConfig {
	return &EvalConfig{
		Config:     cfg.Config.Inherit(),
		Globals:    cfg.Globals,
		Filters:    cfg.Filters,
		Statements: cfg.Statements,
		Tests:      cfg.Tests,
		Loader:     cfg.Loader,
	}
}

func (cfg *EvalConfig) GetTemplate(filename string) (*nodes.Template, error) {
	tpl, err := cfg.Loader.GetTemplate(filename)
	if err != nil {
		return nil, errors.Wrapf(err, `Unable to parse template "%s"`, filename)
	}
	return tpl.Root, nil
}
