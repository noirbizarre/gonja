package parser

import (
	"fmt"

	"github.com/noirbizarre/gonja/log"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/tokens"
)

// type unary struct {
// 	negative bool
// 	term     nodes.Expression
// 	operator *Token
// }

// func (expr *unary) String() string {
// 	t := expr.GetPositionToken()

// 	return fmt.Sprintf("<Unary sign=%s term=%s Line=%d Col=%d>",
// 		expr.operator.Val, expr.term, t.Line, t.Col)
// }

// func (expr *unary) FilterApplied(name string) bool {
// 	return expr.term.FilterApplied(name)
// }

// func (expr *unary) GetPositionToken() *Token {
// 	return expr.operator
// }

// func (expr *unary) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
// 	value, err := expr.Evaluate(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	writer.WriteString(value.String())
// 	return nil
// }

// func (expr *unary) Evaluate(ctx *ExecutionContext) (*Value, error) {
// 	result, err := expr.term.Evaluate(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if expr.negative {
// 		if result.IsNumber() {
// 			switch {
// 			case result.IsFloat():
// 				result = AsValue(-1 * result.Float())
// 			case result.IsInteger():
// 				result = AsValue(-1 * result.Integer())
// 			default:
// 				return nil, ctx.Error("Operation between a number and a non-(float/integer) is not possible", nil)
// 			}
// 		} else {
// 			return nil, ctx.Error("Negative sign on a non-number expression", expr.GetPositionToken())
// 		}
// 	}

// 	return result, nil
// }

func (p *Parser) ParseMath() (nodes.Expression, error) {
	log.Trace("ParseMath", "current", p.Current())

	expr, err := p.parseConcat()
	if err != nil {
		return nil, err
	}

	for p.Peek(tokens.Add, tokens.Sub) != nil {
		op := BinOp(p.Pop())
		right, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		expr = &nodes.BinaryExpression{
			Left:     expr,
			Right:    right,
			Operator: op,
		}
	}

	log.Trace("ParseMath return", "expr", expr)
	return expr, nil
}

func (p *Parser) parseConcat() (nodes.Expression, error) {
	log.Trace("parseConcat", "current", p.Current())

	expr, err := p.ParseMathPrioritary()
	if err != nil {
		return nil, err
	}

	for p.Peek(tokens.Tilde) != nil {
		op := BinOp(p.Pop())
		right, err := p.ParseMathPrioritary()
		if err != nil {
			return nil, err
		}
		expr = &nodes.BinaryExpression{
			Left:     expr,
			Right:    right,
			Operator: op,
		}
	}

	log.Trace("parseConcat return", "expr", expr)
	return expr, nil
}

func (p *Parser) ParseMathPrioritary() (nodes.Expression, error) {
	log.Trace("ParseMathPrioritary", "current", p.Current())

	expr, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.Peek(tokens.Mul, tokens.Div, tokens.Floordiv, tokens.Mod) != nil {
		op := BinOp(p.Pop())
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		expr = &nodes.BinaryExpression{
			Left:     expr,
			Right:    right,
			Operator: op,
		}
	}

	log.Trace("ParseMathPrioritary return", "expr", expr)
	return expr, nil
}

func (p *Parser) parseUnary() (nodes.Expression, error) {
	log.Trace("parseUnary", "current", p.Current())

	sign := p.Match(tokens.Add, tokens.Sub)

	expr, err := p.ParsePower()
	if err != nil {
		return nil, err
	}

	if sign != nil {
		expr = &nodes.UnaryExpression{
			Operator: sign,
			Negative: sign.Val == "-",
			Term:     expr,
		}
	}

	log.Trace("parseUnary return", "expr", expr)
	return expr, nil
}

func (p *Parser) ParsePower() (nodes.Expression, error) {
	log.Trace("ParsePower", "current", p.Current())

	expr, err := p.ParseVariableOrLiteral()
	if err != nil {
		return nil, err
	}

	for p.Peek(tokens.Pow) != nil {
		op := BinOp(p.Pop())
		right, err := p.ParseVariableOrLiteral()
		if err != nil {
			return nil, err
		}
		expr = &nodes.BinaryExpression{
			Left:     expr,
			Right:    right,
			Operator: op,
		}
	}

	log.Trace("ParsePower return", "type", fmt.Sprintf("%T", expr), "expr", expr)
	return expr, nil
}
