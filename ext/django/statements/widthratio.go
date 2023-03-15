package statements

import (
	"fmt"
	"math"

	"github.com/nikolalohinski/gonja/exec"
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
)

type WidthRatioStmt struct {
	Location     *tokens.Token
	current, max nodes.Expression
	width        nodes.Expression
	ctxName      string
}

func (stmt *WidthRatioStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *WidthRatioStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("WidthRatioStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *WidthRatioStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	current := r.Eval(stmt.current)
	if current.IsError() {
		return current
	}

	max := r.Eval(stmt.max)
	if max.IsError() {
		return max
	}

	width := r.Eval(stmt.width)
	if width.IsError() {
		return width
	}

	value := int(math.Ceil(current.Float()/max.Float()*width.Float() + 0.5))

	if stmt.ctxName == "" {
		r.WriteString(fmt.Sprintf("%d", value))
	} else {
		r.Ctx.Set(stmt.ctxName, value)
	}

	return nil
}

func widthratioParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &WidthRatioStmt{
		Location: p.Current(),
	}

	current, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	stmt.current = current

	max, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	stmt.max = max

	width, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	stmt.width = width

	if args.MatchName("as") != nil {
		// Name follows
		nameToken := args.Match(tokens.Name)
		if nameToken == nil {
			return nil, args.Error("Expected name (identifier).", nil)
		}
		stmt.ctxName = nameToken.Val
	}

	if !args.End() {
		return nil, args.Error("Malformed widthratio-tag args.", nil)
	}

	return stmt, nil
}

func init() {
	All.Register("widthratio", widthratioParser)
}
