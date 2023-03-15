package statements

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/nikolalohinski/gonja/exec"
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
)

type IncludeStmt struct {
	// tpl               *Template
	Location      *tokens.Token
	Filename      string
	FilenameExpr  nodes.Expression
	Template      *nodes.Template
	IgnoreMissing bool
	WithContext   bool
	IsEmpty       bool
}

func (stmt *IncludeStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *IncludeStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IncludeStmt(Filename=%s Line=%d Col=%d)", stmt.Filename, t.Line, t.Col)
}

func (stmt *IncludeStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	if stmt.IsEmpty {
		return nil
	}
	sub := r.Inherit()

	if stmt.FilenameExpr != nil {
		filenameValue := r.Eval(stmt.FilenameExpr)
		if filenameValue.IsError() {
			return errors.Wrap(filenameValue, `Unable to evaluate filename`)
		}

		filename := filenameValue.String()
		included, err := r.Loader.GetTemplate(filename)
		if err != nil {
			if stmt.IgnoreMissing {
				return nil
			} else {
				return errors.Wrapf(err, `Unable to load template '%s'`, filename)
			}
		}
		sub.Template = included
		sub.Root = included.Root

	} else {
		sub.Root = stmt.Template
	}

	return sub.Execute()
}

type IncludeEmptyStmt struct{}

// func (node *IncludeEmptyStmt) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
// 	return nil
// }

func includeParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &IncludeStmt{
		Location: p.Current(),
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

	if args.MatchName("ignore") != nil {
		if args.MatchName("missing") != nil {
			stmt.IgnoreMissing = true
		} else {
			args.Stream.Backup()
		}
	}

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
			if stmt.IgnoreMissing {
				stmt.IsEmpty = true
			} else {
				return nil, errors.Wrapf(err, `Unable to parse included template '%s'`, stmt.Filename)
			}
		} else {
			stmt.Template = tpl
		}
	}

	if !args.End() {
		return nil, args.Error("Malformed 'include'-tag args.", nil)
	}

	return stmt, nil
}

func init() {
	All.Register("include", includeParser)
}
