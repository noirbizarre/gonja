package statements

import (
	"fmt"

	"github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/parser"
	"github.com/noirbizarre/gonja/tokens"
	"github.com/pkg/errors"
)

type RawStmt struct {
	Data *nodes.Data
	// Content string
}

func (stmt *RawStmt) Position() *tokens.Token { return stmt.Data.Position() }
func (stmt *RawStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("RawStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *RawStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	if _, err := r.WriteString(stmt.Data.Data.Val); err != nil {
		return errors.Wrap(err, `Unable to execute raw statement`)
	}
	return nil
}

func rawParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &RawStmt{}

	wrapper, _, err := p.WrapUntil("endraw")
	if err != nil {
		return nil, err
	}
	node := wrapper.Nodes[0]
	data, ok := node.(*nodes.Data)
	if ok {
		stmt.Data = data
	} else {
		return nil, p.Error("raw statement can only contains a single data node", node.Position())
	}

	if !args.End() {
		return nil, args.Error("raw statement doesn't accept parameters.", args.Current())
	}

	return stmt, nil
}

func init() {
	All.MustRegister("raw", rawParser)
}
