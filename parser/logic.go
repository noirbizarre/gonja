package parser

import (
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/tokens"
	log "github.com/sirupsen/logrus"
)

var compareOps = []tokens.Type{
	tokens.Eq, tokens.Ne,
	tokens.Gt, tokens.Gteq,
	tokens.Lt, tokens.Lteq,
	// "in", "not in") != nil || p.PeekOne(TokenSymbol, "==", "<=", ">=", "!=", "<>", ">", "<"
}

func BinOp(token *tokens.Token) *nodes.BinOperator {
	return &nodes.BinOperator{token}
}

// type negation struct {
// 	term     Expression
// 	operator *Token
// }

// func (expr *negation) String() string {
// 	t := expr.GetPositionToken()

// 	return fmt.Sprintf("<Negation term=%s Line=%d Col=%d>", expr.term, t.Line, t.Col)
// }

// func (expr *negation) FilterApplied(name string) bool {
// 	return expr.term.FilterApplied(name)
// }

// func (expr *negation) GetPositionToken() *Token {
// 	return expr.operator
// }

// func (expr *negation) Evaluate(ctx *ExecutionContext) (*Value, error) {
// 	result, err := expr.term.Evaluate(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result.Negate(), nil
// }

func (p *Parser) ParseLogicalExpression() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseLogicalExpression")
	return p.parseOr()
}

func (p *Parser) parseOr() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseOr")

	var expr nodes.Expression

	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.CurrentName("or") != nil {
		op := BinOp(p.Pop())
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		expr = &nodes.BinaryExpression{
			Left:     expr,
			Right:    right,
			Operator: op,
		}
	}

	log.WithFields(log.Fields{
		"expr": expr,
	}).Trace("parseOr return")
	return expr, nil
}

func (p *Parser) parseAnd() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseAnd")

	var expr nodes.Expression

	expr, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.CurrentName("and") != nil {
		op := BinOp(p.Pop())
		// binExpr :=

		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		// binExpr.right = right
		expr = &nodes.BinaryExpression{
			Left:     expr,
			Right:    right,
			Operator: op,
		}
	}

	log.WithFields(log.Fields{
		"expr": expr,
	}).Trace("parseAnd return")
	return expr, nil
}

func (p *Parser) parseNot() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseNot")

	op := p.MatchName("not")
	expr, err := p.parseCompare()
	if err != nil {
		return nil, err
	}

	if op != nil {
		expr = &nodes.Negation{
			Operator: op,
			Term:     expr,
		}
	}

	log.WithFields(log.Fields{
		"expr": expr,
	}).Trace("parseNot return")
	return expr, nil
}

func (p *Parser) parseCompare() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseCompare")

	var expr nodes.Expression

	expr, err := p.ParseMath()
	if err != nil {
		return nil, err
	}

	// for p.PeekOne(TokenKeyword, "in", "not in") != nil || p.PeekOne(TokenSymbol, "==", "<=", ">=", "!=", "<>", ">", "<") != nil {
	for p.Current(compareOps...) != nil || p.CurrentName("in", "not") != nil {

		op := p.Pop()
		// if op = p.MatchOne(TokenKeyword, "in", "not in"); op == nil {
		// 	if op = p.MatchOne(TokenSymbol, "==", "<=", ">=", "!=", "<>", ">", "<"); op == nil {
		// 		return nil, p.Error("Unexpected operator %s", p.Current())
		// 	}
		// }

		right, err := p.ParseMath()
		if err != nil {
			return nil, err
		}

		if right != nil {
			expr = &nodes.BinaryExpression{
				Left:     expr,
				Operator: BinOp(op),
				Right:    right,
			}
		}
	}

	expr, err = p.ParseTest(expr)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"expr": expr,
	}).Trace("parseCompare return")
	return expr, nil
}
