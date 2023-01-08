package parser

import (
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/tokens"
	log "github.com/sirupsen/logrus"
)

// ParseFilterExpression parses an optional filter chain for a node
func (p *Parser) ParseFilterExpression(expr nodes.Expression) (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseFilterExpression")

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

	log.WithFields(log.Fields{
		"expr": expr,
	}).Trace("ParseFilterExpression return")
	return expr, nil
}

// ParseExpression parses an expression with optional filters
// Nested expression should call this method

func (p *Parser) ParseExpression() (nodes.Expression, error) {
	return p.parseExpression(false)
}

func (p *Parser) ParseExpressionWithInlineIfs() (nodes.Expression, error) {
	return p.parseExpression(true)
}

func (p *Parser) parseExpression(withInlineIfs bool) (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseExpression")
	var expr nodes.Expression

	expr, err := p.ParseLogicalExpression()
	if err != nil {
		return nil, err
	}

	if withInlineIfs && p.PeekName("if") != nil {
		BinOp(p.Pop())
		condition, conditionErr := p.ParseLogicalExpression()
		if conditionErr != nil {
			return nil, conditionErr
		}

		var falseBranch nodes.Expression = &nodes.String{
			Location: p.Current(),
			Val:      "",
		}
		if p.PeekName("else") != nil {
			BinOp(p.Pop())
			var falseBranchErr error
			falseBranch, falseBranchErr = p.ParseExpression()
			if falseBranchErr != nil {
				return nil, falseBranchErr
			}
		}

		expr = &nodes.InlineIfExpression{
			TrueBranch:  expr,
			FalseBranch: falseBranch,
			Condition:   condition,
		}
	}

	expr, err = p.ParseFilterExpression(expr)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"expr": expr,
	}).Trace("ParseExpression return")
	return expr, nil
}

func (p *Parser) ParseExpressionNode() (nodes.Node, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseExpressionNode")

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

	expr, err := p.ParseExpressionWithInlineIfs()
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

	log.WithFields(log.Fields{
		"node": node,
	}).Trace("parseExpressionNode return")
	return node, nil
}
