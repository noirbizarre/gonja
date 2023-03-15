package statements

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/nikolalohinski/gonja/exec"
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
)

type ImportStmt struct {
	Location     *tokens.Token
	Filename     string
	FilenameExpr nodes.Expression
	As           string
	WithContext  bool
	Template     *nodes.Template
}

func (stmt *ImportStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *ImportStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("ImportStmt(Line=%d Col=%d)", t.Line, t.Col)
}
func (stmt *ImportStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	var imported map[string]*nodes.Macro
	macros := map[string]exec.Macro{}

	if stmt.FilenameExpr != nil {
		filenameValue := r.Eval(stmt.FilenameExpr)
		if filenameValue.IsError() {
			return errors.Wrap(filenameValue, `Unable to evaluate filename`)
		}

		filename := filenameValue.String()
		tpl, err := r.Loader.GetTemplate(filename)
		if err != nil {
			return errors.Wrapf(err, `Unable to load template '%s'`, filename)
		}
		imported = tpl.Root.Macros

	} else {
		imported = stmt.Template.Macros
	}

	for name, macro := range imported {
		fn, err := exec.MacroNodeToFunc(macro, r)
		if err != nil {
			return errors.Wrapf(err, `Unable to import macro '%s'`, name)
		}
		macros[name] = fn
	}

	r.Ctx.Set(stmt.As, macros)
	return nil
}

type FromImportStmt struct {
	Location     *tokens.Token
	Filename     string
	FilenameExpr nodes.Expression
	WithContext  bool
	Template     *nodes.Template
	As           map[string]string
	Macros       map[string]*nodes.Macro // alias/name -> macro instance
}

func (stmt *FromImportStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *FromImportStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("FromImportStmt(Line=%d Col=%d)", t.Line, t.Col)
}
func (stmt *FromImportStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	var imported map[string]*nodes.Macro

	if stmt.FilenameExpr != nil {
		filenameValue := r.Eval(stmt.FilenameExpr)
		if filenameValue.IsError() {
			return errors.Wrap(filenameValue, `Unable to evaluate filename`)
		}

		filename := filenameValue.String()
		tpl, err := r.Loader.GetTemplate(filename)
		if err != nil {
			return errors.Wrapf(err, `Unable to load template '%s'`, filename)
		}
		imported = tpl.Root.Macros

	} else {
		imported = stmt.Template.Macros
	}

	for alias, name := range stmt.As {
		node := imported[name]
		fn, err := exec.MacroNodeToFunc(node, r)
		if err != nil {
			return errors.Wrapf(err, `Unable to import macro '%s'`, name)
		}
		r.Ctx.Set(alias, fn)
	}
	return nil
}

func importParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &ImportStmt{
		Location: p.Current(),
		// Macros:   map[string]*nodes.Macro{},
	}

	if args.End() {
		return nil, args.Error("You must at least specify one macro to import.", nil)
	}

	if tok := args.Match(tokens.String); tok != nil {
		stmt.Filename = tok.Val
	} else {
		expr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		stmt.FilenameExpr = expr
	}
	if args.MatchName("as") == nil {
		return nil, args.Error(`Expected "as" keyword`, args.Current())
	}

	alias := args.Match(tokens.Name)
	if alias == nil {
		return nil, args.Error("Expected macro alias name (identifier)", args.Current())
	}
	stmt.As = alias.Val

	if tok := args.MatchName("with", "without"); tok != nil {
		if args.MatchName("context") != nil {
			stmt.WithContext = tok.Val == "with"
		} else {
			args.Stream.Backup()
		}
	}

	// Preload static template
	if stmt.Filename != "" {
		tpl, err := p.TemplateParser(stmt.Filename)
		if err != nil {
			return nil, errors.Wrapf(err, `Unable to parse imported template '%s'`, stmt.Filename)
		} else {
			stmt.Template = tpl
		}
	}

	return stmt, nil
}

func fromParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &FromImportStmt{
		Location: p.Current(),
		As:       map[string]string{},
		// Macros:   map[string]*nodes.Macro{},
	}

	if args.End() {
		return nil, args.Error("You must at least specify one macro to import.", nil)
	}

	if tok := args.Match(tokens.String); tok != nil {
		stmt.Filename = tok.Val
	} else {
		filename, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		stmt.FilenameExpr = filename
	}

	if args.MatchName("import") == nil {
		return nil, args.Error("Expected import keyword", args.Current())
	}

	for !args.End() {
		name := args.Match(tokens.Name)
		if name == nil {
			return nil, args.Error("Expected macro name (identifier).", args.Current())
		}

		// asName := macroNameToken.Val
		if args.MatchName("as") != nil {
			alias := args.Match(tokens.Name)
			if alias == nil {
				return nil, args.Error("Expected macro alias name (identifier).", nil)
			}
			// asName = aliasToken.Val
			stmt.As[alias.Val] = name.Val
		} else {
			stmt.As[name.Val] = name.Val
		}

		// macroInstance, has := tpl.exportedMacros[macroNameToken.Val]
		// if !has {
		// 	return nil, args.Error(fmt.Sprintf("Macro '%s' not found (or not exported) in '%s'.", macroNameToken.Val,
		// 		stmt.filename), macroNameToken)
		// }

		// stmt.macros[asName] = macroInstance
		if tok := args.MatchName("with", "without"); tok != nil {
			if args.MatchName("context") != nil {
				stmt.WithContext = tok.Val == "with"
				break
			} else {
				args.Stream.Backup()
			}
		}

		if args.End() {
			break
		}

		if args.Match(tokens.Comma) == nil {
			return nil, args.Error("Expected ','.", nil)
		}
	}

	// Preload static template
	if stmt.Filename != "" {
		tpl, err := p.TemplateParser(stmt.Filename)
		if err != nil {
			return nil, errors.Wrapf(err, `Unable to parse imported template '%s'`, stmt.Filename)
		} else {
			stmt.Template = tpl
		}
	}

	return stmt, nil
}

func init() {
	All.Register("import", importParser)
	All.Register("from", fromParser)
}
