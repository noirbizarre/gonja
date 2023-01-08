package statements

import (
	"fmt"
	"strings"

	"github.com/paradime-io/gonja/exec"
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/parser"
	"github.com/paradime-io/gonja/tokens"
	"github.com/pkg/errors"
)

type SetStmt struct {
	Location   *tokens.Token
	Target     nodes.Expression
	Expression *nodes.Expression
	Wrapper    *nodes.Wrapper
}

func (stmt *SetStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *SetStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("SetStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *SetStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	// Evaluate expression
	var value *exec.Value
	if stmt.Expression != nil {
		value = r.Eval(*stmt.Expression)
		if value.IsError() {
			return value
		}
	} else if stmt.Wrapper != nil {
		ri := r.Inherit()
		var buffer strings.Builder
		ri.Out = &buffer
		if execErr := ri.ExecuteWrapper(stmt.Wrapper); execErr != nil {
			return execErr
		}
		value = exec.AsValue(ri.String())
	} else {
		return fmt.Errorf("no value is given in the set block")
	}

	switch n := stmt.Target.(type) {
	case *nodes.Name:
		r.Ctx.Set(n.Name.Val, value.Interface())
	case *nodes.GetAttr:
		target := r.Eval(n.Node)
		if target.IsError() {
			return errors.Wrapf(target, `Unable to evaluate target %s`, n)
		}
		if n.Attr == "" {
			return errors.Errorf(`Not implemented to evaluate GetAttr at %d`, n.Index) // TODO: implement
		} else if err := target.Set(n.Attr, value.Interface()); err != nil {
			return errors.Wrapf(err, `Unable to set value on "%s"`, n.Attr)
		}
	case *nodes.GetItem:
		target := r.Eval(n.Node)
		if target.IsError() {
			return errors.Wrapf(target, `Unable to evaluate target %s`, n)
		}
		if n.Arg == nil {
			return errors.Errorf(`Not implemented to evaluate GetItem at %d`, n.Index) // TODO: implement
		} else if err := target.Set(r.Eval(*n.Arg).String(), value.Interface()); err != nil {
			return errors.Wrapf(err, `Unable to set value on "%s"`, n.Arg)
		}
	default:
		return errors.Errorf(`Illegal set target node %s`, n)
	}

	return nil
}

func setParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &SetStmt{
		Location: p.Current(),
	}

	// Parse variable name
	ident, err := args.ParseVariable()
	if err != nil {
		return nil, errors.Wrap(err, `Unable to parse identifier`)
	}
	switch n := ident.(type) {
	case *nodes.Name, *nodes.Call, *nodes.GetItem, *nodes.GetAttr:
		stmt.Target = n
	default:
		return nil, errors.Errorf(`Unexpected set target %s`, n)
	}

	if args.Match(tokens.Assign) == nil {
		wrapper, endArgs, err := p.WrapUntil("endset")
		if err != nil {
			return nil, err
		}
		stmt.Wrapper = wrapper

		if !endArgs.End() {
			return nil, endArgs.Error("Arguments not allowed here.", nil)
		}

		return stmt, nil
	} else {

		// Variable expression
		expr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Expression = &expr

		// Remaining arguments
		if !args.End() {
			return nil, args.Error("Malformed 'set'-tag args.", args.Current())
		}

		return stmt, nil
	}
}

func init() {
	All.Register("set", setParser)
}
