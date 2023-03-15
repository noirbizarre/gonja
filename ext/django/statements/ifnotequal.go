package statements

import (
	"fmt"

	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
)

type IfNotEqualStmt struct {
	Location    *tokens.Token
	var1, var2  nodes.Expression
	thenWrapper *nodes.Wrapper
	elseWrapper *nodes.Wrapper
}

func (stmt *IfNotEqualStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *IfNotEqualStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfNotEqualStmt(Line=%d Col=%d)", t.Line, t.Col)
}

// func (node *IfNotEqualStmt) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
// 	r1, err := node.var1.Evaluate(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	r2, err := node.var2.Evaluate(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	result := !r1.EqualValueTo(r2)

// 	if result {
// 		return node.thenWrapper.Execute(ctx, writer)
// 	}
// 	if node.elseWrapper != nil {
// 		return node.elseWrapper.Execute(ctx, writer)
// 	}
// 	return nil
// }

func ifNotEqualParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	ifnotequalNode := &IfNotEqualStmt{}

	// Parse two expressions
	var1, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	var2, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	ifnotequalNode.var1 = var1
	ifnotequalNode.var2 = var2

	if !args.End() {
		return nil, args.Error("ifequal only takes 2 args.", nil)
	}

	// Wrap then/else-blocks
	wrapper, endargs, err := p.WrapUntil("else", "endifnotequal")
	if err != nil {
		return nil, err
	}
	ifnotequalNode.thenWrapper = wrapper

	if !endargs.End() {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	if wrapper.EndTag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endargs, err = p.WrapUntil("endifnotequal")
		if err != nil {
			return nil, err
		}
		ifnotequalNode.elseWrapper = wrapper

		if !endargs.End() {
			return nil, endargs.Error("Arguments not allowed here.", nil)
		}
	}

	return ifnotequalNode, nil
}

func init() {
	All.Register("ifnotequal", ifNotEqualParser)
}
