package statements

import (
	// "github.com/noirbizarre/gonja/exec"
	"fmt"

	"github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/parser"
	"github.com/noirbizarre/gonja/tokens"
	"github.com/pkg/errors"
)

type FirstofStmt struct {
	Location *tokens.Token
	Args     []nodes.Expression
}

func (stmt *FirstofStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *FirstofStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("FirstofStmt(Args=%s, Line=%d Col=%d)", stmt.Args, t.Line, t.Col)
}

func (stmt *FirstofStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	for _, arg := range stmt.Args {
		val := r.Eval(arg)
		if val.IsError() {
			return val
		}

		if val.IsTrue() {
			if err := r.RenderValue(val); err != nil {
				return errors.Wrap(err, `Unable to execute 'firstof' statement`)
			}
			return nil
		}
	}

	return nil
}

func firstofParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &FirstofStmt{
		Location: p.Current(),
	}

	for !args.End() {
		node, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Args = append(stmt.Args, node)
	}

	return stmt, nil
}

func init() {
	All.MustRegister("firstof", firstofParser)
}
