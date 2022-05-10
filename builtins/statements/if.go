package statements

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/paradime-io/gonja/exec"
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/parser"
	"github.com/paradime-io/gonja/tokens"
)

type IfStmt struct {
	Location   *tokens.Token
	Conditions []nodes.Expression
	Wrappers   []*nodes.Wrapper
}

func (stmt *IfStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *IfStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (node *IfStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	for i, condition := range node.Conditions {
		result := r.Eval(condition)
		if result.IsError() {
			return result
		}

		if result.IsTrue() {
			return r.ExecuteWrapper(node.Wrappers[i])
		}
		// Last condition?
		if len(node.Conditions) == i+1 && len(node.Wrappers) > i+1 {
			return r.ExecuteWrapper(node.Wrappers[i+1])
		}
	}
	return nil
}

func ifParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	log.WithFields(log.Fields{
		"arg":     args.Current(),
		"current": p.Current(),
	}).Trace("ParseIf")
	ifNode := &IfStmt{
		Location: args.Current(),
	}

	// Parse first and main IF condition
	condition, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	ifNode.Conditions = append(ifNode.Conditions, condition)

	if !args.End() {
		return nil, args.Error("If-condition is malformed.", nil)
	}

	// Check the rest
	for {
		wrapper, tagArgs, err := p.WrapUntil("elif", "else", "endif")
		if err != nil {
			return nil, err
		}
		ifNode.Wrappers = append(ifNode.Wrappers, wrapper)

		if wrapper.EndTag == "elif" {
			// elif can take a condition
			condition, err = tagArgs.ParseExpression()
			if err != nil {
				return nil, err
			}
			ifNode.Conditions = append(ifNode.Conditions, condition)

			if !tagArgs.End() {
				return nil, tagArgs.Error("Elif-condition is malformed.", nil)
			}
		} else {
			if !tagArgs.End() {
				// else/endif can't take any conditions
				return nil, tagArgs.Error("Arguments not allowed here.", nil)
			}
		}

		if wrapper.EndTag == "endif" {
			break
		}
	}

	return ifNode, nil
}

func init() {
	All.Register("if", ifParser)
}
