package statements

import (
	// "bytes"

	"fmt"
	"strings"

	"github.com/paradime-io/gonja/exec"
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/parser"
	"github.com/paradime-io/gonja/tokens"
)

type IfChangedStmt struct {
	Location    *tokens.Token
	watchedExpr []nodes.Expression
	lastValues  []*exec.Value
	lastContent string
	thenWrapper *nodes.Wrapper
	elseWrapper *nodes.Wrapper
}

func (stmt *IfChangedStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *IfChangedStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfChangedStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *IfChangedStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	if len(stmt.watchedExpr) == 0 {
		// Check against own rendered body
		var out strings.Builder
		sub := r.Inherit()
		sub.Out = &out
		err := sub.ExecuteWrapper(stmt.thenWrapper)
		if err != nil {
			return err
		}

		str := out.String()
		if stmt.lastContent != str {
			// Rendered content changed, output it
			r.WriteString(str)
			stmt.lastContent = str
		}
	} else {
		nowValues := make([]*exec.Value, 0, len(stmt.watchedExpr))
		for _, expr := range stmt.watchedExpr {
			val := r.Eval(expr)
			if val.IsError() {
				return val
			}
			nowValues = append(nowValues, val)
		}

		// Compare old to new values now
		changed := len(stmt.lastValues) == 0

		for idx, oldVal := range stmt.lastValues {
			if !oldVal.EqualValueTo(nowValues[idx]) {
				changed = true
				break // we can stop here because ONE value changed
			}
		}

		stmt.lastValues = nowValues

		if changed {
			// Render thenWrapper
			err := r.ExecuteWrapper(stmt.thenWrapper)
			if err != nil {
				return err
			}
		} else {
			// Render elseWrapper
			err := r.ExecuteWrapper(stmt.elseWrapper)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ifchangedParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &IfChangedStmt{
		Location: p.Current(),
	}

	for !args.End() {
		// Parse condition
		expr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		stmt.watchedExpr = append(stmt.watchedExpr, expr)
	}

	if !args.End() {
		return nil, args.Error("Ifchanged-arguments are malformed.", nil)
	}

	// Wrap then/else-blocks
	wrapper, endArgs, err := p.WrapUntil("else", "endifchanged")
	if err != nil {
		return nil, err
	}
	stmt.thenWrapper = wrapper

	if !endArgs.End() {
		return nil, endArgs.Error("Arguments not allowed here.", nil)
	}

	if wrapper.EndTag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endArgs, err = p.WrapUntil("endifchanged")
		if err != nil {
			return nil, err
		}
		stmt.elseWrapper = wrapper

		if !endArgs.End() {
			return nil, endArgs.Error("Arguments not allowed here.", nil)
		}
	}

	return stmt, nil
}

func init() {
	All.Register("ifchanged", ifchangedParser)
}
