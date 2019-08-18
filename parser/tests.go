package parser

import (
	log "github.com/sirupsen/logrus"

	"github.com/noirbizarre/gonja/nodes"
)

func (p *Parser) ParseTest(expr nodes.Expression) (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseTest")

	expr, err := p.ParseFilterExpression(expr)
	if err != nil {
		return nil, err
	}

	if p.MatchName("is") != nil {
		not := p.MatchName("not")
		ident := p.Next()

		test := &nodes.TestCall{
			Token:  ident,
			Name:   ident.Val,
			Args:   []nodes.Expression{},
			Kwargs: map[string]nodes.Expression{},
		}

		arg, err := p.ParseExpression()
		if err == nil && arg != nil {
			test.Args = append(test.Args, arg)
		}

		

		// // Check for test-argument (2 tokens needed: ':' ARG)
		// if p.Match(tokens.Lparen) != nil {
		// 	if p.Peek(tokens.VariableEnd) != nil {
		// 		return nil, p.Error("Filter parameter required after '('.", nil)
		// 	}

		// 	for p.Match(tokens.Comma) != nil || p.Match(tokens.Rparen) == nil {
		// 		// TODO: Handle multiple args and kwargs
		// 		v, err := p.ParseExpression()
		// 		if err != nil {
		// 			return nil, err
		// 		}

		// 		if p.Match(tokens.Assign) != nil {
		// 			key := v.Position().Val
		// 			value, errValue := p.ParseExpression()
		// 			if errValue != nil {
		// 				return nil, errValue
		// 			}
		// 			test.Kwargs[key] = value
		// 		} else {
		// 			test.Args = append(test.Args, v)
		// 		}
		// 	}
		// } else {
		// 	arg, err := p.ParseExpression()
		// 	if err == nil && arg != nil {
		// 		test.Args = append(test.Args, arg)
		// 	}
		// }

		expr = &nodes.TestExpression{
			Expression: expr,
			Test:       test,
		}

		if not != nil {
			expr = &nodes.Negation{expr, not}
		}
	}

	log.WithFields(log.Fields{
		"expr": expr,
	}).Trace("parseTest return")
	return expr, nil
}
