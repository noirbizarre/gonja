package statements

import (
	"fmt"

	"github.com/nikolalohinski/gonja/exec"
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
	"github.com/pkg/errors"
)

type MacroStmt struct {
	*nodes.Macro
}

// func (stmt *MacroStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *MacroStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("MacroStmt(Macro=%s Line=%d Col=%d)", stmt.Macro, t.Line, t.Col)
}

func (stmt *MacroStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	macro, err := exec.MacroNodeToFunc(stmt.Macro, r)
	if err != nil {
		return errors.Wrapf(err, `Unable to parse marco '%s'`, stmt.Name)
	}
	r.Ctx.Set(stmt.Name, macro)
	return nil
}

func (node *MacroStmt) call(ctx *exec.Context, args ...*exec.Value) *exec.Value {
	// argsCtx := make(exec.Context)

	// for k, v := range node.args {
	// 	if v == nil {
	// 		// User did not provided a default value
	// 		argsCtx[k] = nil
	// 	} else {
	// 		// Evaluate the default value
	// 		valueExpr, err := v.Evaluate(ctx)
	// 		if err != nil {
	// 			ctx.Logf(err.Error())
	// 			return AsSafeValue(err.Error())
	// 		}

	// 		argsCtx[k] = valueExpr
	// 	}
	// }

	// if len(args) > len(node.argsOrder) {
	// 	// Too many arguments, we're ignoring them and just logging into debug mode.
	// 	err := ctx.Error(fmt.Sprintf("Macro '%s' called with too many arguments (%d instead of %d).",
	// 		node.name, len(args), len(node.argsOrder)), nil).updateFromTokenIfNeeded(ctx.template, node.position)

	// 	ctx.Logf(err.Error()) // TODO: This is a workaround, because the error is not returned yet to the Execution()-methods
	// 	return AsSafeValue(err.Error())
	// }

	// // Make a context for the macro execution
	// macroCtx := NewChildExecutionContext(ctx)

	// // Register all arguments in the private context
	// macroCtx.Private.Update(argsCtx)

	// for idx, argValue := range args {
	// 	macroCtx.Private[node.argsOrder[idx]] = argValue.Interface()
	// }

	// var b bytes.Buffer
	// err := node.wrapper.Execute(macroCtx, &b)
	// if err != nil {
	// 	return AsSafeValue(err.updateFromTokenIfNeeded(ctx.template, node.position).Error())
	// }

	// return AsSafeValue(b.String())
	return nil
}

func macroParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &nodes.Macro{
		Location: p.Current(),
		Args:     []string{},
		Kwargs:   []*nodes.Pair{},
	}

	name := args.Match(tokens.Name)
	if name == nil {
		return nil, args.Error("Macro-tag needs at least an identifier as name.", nil)
	}
	stmt.Name = name.Val

	if args.Match(tokens.Lparen) == nil {
		return nil, args.Error("Expected '('.", nil)
	}

	for args.Match(tokens.Rparen) == nil {
		argName := args.Match(tokens.Name)
		if argName == nil {
			return nil, args.Error("Expected argument name as identifier.", nil)
		}

		if args.Match(tokens.Assign) != nil {
			// Default expression follows
			expr, err := args.ParseExpression()
			if err != nil {
				return nil, err
			}
			stmt.Kwargs = append(stmt.Kwargs, &nodes.Pair{
				Key:   &nodes.String{argName, argName.Val},
				Value: expr,
			})
			// stmt.Kwargs[argName.Val] = expr
		} else {
			stmt.Args = append(stmt.Args, argName.Val)
		}

		if args.Match(tokens.Rparen) != nil {
			break
		}
		if args.Match(tokens.Comma) == nil {
			return nil, args.Error("Expected ',' or ')'.", nil)
		}
	}

	// if args.MatchName("export") != nil {
	// 	stmt.exported = true
	// }

	if !args.End() {
		return nil, args.Error("Malformed macro-tag.", nil)
	}

	// Body wrapping
	wrapper, endargs, err := p.WrapUntil("endmacro")
	if err != nil {
		return nil, err
	}
	stmt.Wrapper = wrapper

	if !endargs.End() {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	p.Template.Macros[stmt.Name] = stmt

	// if stmt.exported {
	// 	// Now register the macro if it wants to be exported
	// 	_, has := p.template.exportedMacros[stmt.name]
	// 	if has {
	// 		return nil, p.Error(fmt.Sprintf("another macro with name '%s' already exported", stmt.name), start)
	// 	}
	// 	p.template.exportedMacros[stmt.name] = stmt
	// }

	return &MacroStmt{stmt}, nil
}

func init() {
	All.Register("macro", macroParser)
}
