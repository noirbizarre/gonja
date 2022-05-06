package statements

import (
	"fmt"

	"github.com/paradime-io/gonja/exec"
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/parser"
	"github.com/paradime-io/gonja/tokens"
)

type cycleValue struct {
	node  *CycleStatement
	value *exec.Value
}

type CycleStatement struct {
	position *tokens.Token
	args     []nodes.Expression
	idx      int
	asName   string
	silent   bool
}

func (stmt *CycleStatement) Position() *tokens.Token { return stmt.position }
func (stmt *CycleStatement) String() string {
	t := stmt.Position()
	return fmt.Sprintf("CycleStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (cv *cycleValue) String() string {
	return cv.value.String()
}

func (stmt *CycleStatement) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	item := stmt.args[stmt.idx%len(stmt.args)]
	stmt.idx++

	val := r.Eval(item)
	if val.IsError() {
		return val
	}

	if t, ok := val.Interface().(*cycleValue); ok {
		// {% cycle "test1" "test2"
		// {% cycle cycleitem %}

		// Update the cycle value with next value
		item := t.node.args[t.node.idx%len(t.node.args)]
		t.node.idx++

		val := r.Eval(item)
		if val.IsError() {
			return val
		}

		t.value = val

		if !t.node.silent {
			r.WriteString(val.String())
		}
	} else {
		// Regular call

		cycleValue := &cycleValue{
			node:  stmt,
			value: val,
		}

		if stmt.asName != "" {
			r.Ctx.Set(stmt.asName, cycleValue)
		}
		if !stmt.silent {
			r.WriteString(val.String())
		}
	}

	return nil
}

// HINT: We're not supporting the old comma-separated list of expressions argument-style
func cycleParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	cycleNode := &CycleStatement{
		position: p.Current(),
	}

	for !args.End() {
		node, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		cycleNode.args = append(cycleNode.args, node)

		if args.MatchName("as") != nil {
			// as

			name := args.Match(tokens.Name)
			if name == nil {
				return nil, args.Error("Name (identifier) expected after 'as'.", nil)
			}
			cycleNode.asName = name.Val

			if args.MatchName("silent") != nil {
				cycleNode.silent = true
			}

			// Now we're finished
			break
		}
	}

	if !args.End() {
		return nil, args.Error("Malformed cycle-tag.", nil)
	}

	return cycleNode, nil
}

func init() {
	All.Register("cycle", cycleParser)
}
