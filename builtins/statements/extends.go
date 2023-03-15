package statements

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/nikolalohinski/gonja/exec"
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
)

type ExtendsStmt struct {
	Location    *tokens.Token
	Filename    string
	WithContext bool
}

func (stmt *ExtendsStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *ExtendsStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("ExtendsStmt(Filename=%s Line=%d Col=%d)", stmt.Filename, t.Line, t.Col)
}

func (node *ExtendsStmt) Execute(r *exec.Renderer) error {
	return nil
}

func extendsParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &ExtendsStmt{
		Location: p.Current(),
	}

	if p.Level > 1 {
		return nil, args.Error(`The 'extends' statement can only be defined at root level`, p.Current())
	}

	if p.Template.Parent != nil {
		return nil, args.Error("This template has already one parent.", args.Current())
	}

	// var filename nodes.Node
	if filename := args.Match(tokens.String); filename != nil {
		stmt.Filename = filename.Val
		tpl, err := p.TemplateParser(stmt.Filename)
		if err != nil {
			return nil, errors.Wrapf(err, `Unable to parse parent template '%s'`, stmt.Filename)
		}
		p.Template.Parent = tpl
	} else {
		return nil, args.Error("Tag 'extends' requires a template filename as string.", args.Current())
	}

	if tok := args.MatchName("with", "without"); tok != nil {
		if args.MatchName("context") != nil {
			stmt.WithContext = tok.Val == "with"
		} else {
			args.Stream.Backup()
		}
	}

	if !args.End() {
		return nil, args.Error("Tag 'extends' does only take 1 argument.", nil)
	}

	return stmt, nil
}

func init() {
	All.Register("extends", extendsParser)
}
