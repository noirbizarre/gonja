package parser

import (
	"github.com/noirbizarre/gonja/log"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/tokens"
)

// ParseFilterExpression parses an optionnal filter chain for a node
func (p *Parser) ParseFilterExpression(expr nodes.Expression) (nodes.Expression, error) {
	log.Trace("ParseFilterExpression", "current", p.Current())

	if p.Peek(tokens.Pipe) != nil {

		filtered := &nodes.FilteredExpression{
			Expression: expr,
		}
		for p.Match(tokens.Pipe) != nil {
			// Parse one single filter
			filter, err := p.ParseFilter()
			if err != nil {
				return nil, err
			}

			// Check sandbox filter restriction
			// if _, isBanned := p.template.set.bannedFilters[filter.name]; isBanned {
			// 	return nil, p.Error(fmt.Sprintf("Usage of filter '%s' is not allowed (sandbox restriction active).", filter.name), nil)
			// }

			filtered.Filters = append(filtered.Filters, filter)
		}
		expr = filtered
	}

	log.Trace("ParseFilterExpression return", "expr", expr)
	return expr, nil
}

// ParseExpression parses an expression with optionnal filters
// Nested expression shoulds call this method
func (p *Parser) ParseExpression() (nodes.Expression, error) {
	log.Trace("ParseExpression", "current", p.Current())
	var expr nodes.Expression

	expr, err := p.ParseLogicalExpression()
	if err != nil {
		return nil, err
	}

	expr, err = p.ParseFilterExpression(expr)
	if err != nil {
		return nil, err
	}

	log.Trace("ParseExpression return", "expr", expr)
	return expr, nil
}

func (p *Parser) ParseExpressionNode() (nodes.Node, error) {
	log.Trace("ParseExpressionNode", "current", p.Current())

	tok := p.Match(tokens.VariableBegin)
	if tok == nil {
		return nil, p.Error("'{{' expected here", p.Current())
	}

	node := &nodes.Output{
		Start: tok,
		Trim: &nodes.Trim{
			Left: tok.Val[len(tok.Val)-1] == '-',
		},
	}

	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	if expr == nil {
		return nil, p.Error("Expected an expression.", p.Current())
	}
	node.Expression = expr

	tok = p.Match(tokens.VariableEnd)
	if tok == nil {
		return nil, p.Error("'}}' expected here", p.Current())
	}
	node.End = tok
	node.Trim.Right = tok.Val[0] == '-'

	log.Trace("parseExpressionNode return", "node", node)
	return node, nil
}
