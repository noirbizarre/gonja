package parser

import (
	// "fmt"
	"strconv"
	"strings"

	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/tokens"
	log "github.com/sirupsen/logrus"
)

func (p *Parser) parseNumber() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseNumber")
	t := p.Match(tokens.Integer, tokens.Float)
	if t == nil {
		return nil, p.Error("Expected a number", t)
	}

	if t.Type == tokens.Integer {
		i, err := strconv.Atoi(t.Val)
		if err != nil {
			return nil, p.Error(err.Error(), t)
		}
		nr := &nodes.Integer{
			Location: t,
			Val:      i,
		}
		return nr, nil
	} else {
		f, err := strconv.ParseFloat(t.Val, 64)
		if err != nil {
			return nil, p.Error(err.Error(), t)
		}
		fr := &nodes.Float{
			Location: t,
			Val:      f,
		}
		return fr, nil
	}
}

func (p *Parser) parseString() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseString")
	t := p.Match(tokens.String)
	if t == nil {
		return nil, p.Error("Expected a string", t)
	}
	str := strconv.Quote(t.Val)
	replaced := strings.Replace(str, `\\`, "\\", -1)
	newstr, err := strconv.Unquote(replaced)
	if err != nil {
		return nil, p.Error(err.Error(), t)
	}
	sr := &nodes.String{
		Location: t,
		Val:      newstr,
	}
	return sr, nil
}

func (p *Parser) parseCollection() (nodes.Expression, error) {
	switch p.Current().Type {
	case tokens.Lbracket:
		return p.parseList()
	case tokens.Lparen:
		return p.parseTuple()
	case tokens.Lbrace:
		return p.parseDict()
	default:
		return nil, nil
	}
}

func (p *Parser) parseList() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseList")
	t := p.Match(tokens.Lbracket)
	if t == nil {
		return nil, p.Error("Expected [", t)
	}

	if p.Match(tokens.Rbracket) != nil {
		// Empty list
		return &nodes.List{Location: t, Val: []nodes.Expression{}}, nil
	}

	expr, err := p.ParseExpressionWithInlineIfs()
	if err != nil {
		return nil, err
	}
	list := []nodes.Expression{expr}

	for p.Match(tokens.Comma) != nil {
		if p.Peek(tokens.Rbracket) != nil {
			// Trailing coma
			break
		}
		expr, err := p.ParseExpressionWithInlineIfs()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			return nil, p.Error("Expected a value", p.Current())
		}
		list = append(list, expr)
	}

	if p.Match(tokens.Rbracket) == nil {
		return nil, p.Error("Expected ]", p.Current())
	}

	return &nodes.List{Location: t, Val: list}, nil
}

func (p *Parser) parseTuple() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseTuple")
	t := p.Match(tokens.Lparen)
	if t == nil {
		return nil, p.Error("Expected (", t)
	}
	expr, err := p.ParseExpressionWithInlineIfs()
	if err != nil {
		return nil, err
	}
	list := []nodes.Expression{expr}

	trailingComa := false

	for p.Match(tokens.Comma) != nil {
		if p.Peek(tokens.Rparen) != nil {
			// Trailing coma
			trailingComa = true
			break
		}
		expr, err := p.ParseExpressionWithInlineIfs()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			return nil, p.Error("Expected a value", p.Current())
		}
		list = append(list, expr)
	}

	if p.Match(tokens.Rparen) == nil {
		return nil, p.Error("Unbalanced parenthesis", t)
		// return nil, p.Error("Expected )", p.Current())
	}

	if len(list) > 1 || trailingComa {
		return &nodes.Tuple{Location: t, Val: list}, nil
	} else {
		return expr, nil
	}
}

func (p *Parser) parsePair() (*nodes.Pair, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parsePair")
	key, err := p.ParseExpressionWithInlineIfs()
	if err != nil {
		return nil, err
	}

	if p.Match(tokens.Colon) == nil {
		return nil, p.Error("Expected \":\"", p.Current())
	}
	value, err := p.ParseExpressionWithInlineIfs()
	if err != nil {
		return nil, err
	}
	return &nodes.Pair{
		Key:   key,
		Value: value,
	}, nil
}

func (p *Parser) parseDict() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("parseDict")
	t := p.Match(tokens.Lbrace)
	if t == nil {
		return nil, p.Error("Expected {", t)
	}

	dict := &nodes.Dict{
		Token: t,
		Pairs: []*nodes.Pair{},
	}

	if p.Peek(tokens.Rbrace) == nil {
		pair, err := p.parsePair()
		if err != nil {
			return nil, err
		}
		dict.Pairs = append(dict.Pairs, pair)
	}

	for p.Match(tokens.Comma) != nil {
		if p.Peek(tokens.Rbrace) == nil {
			pair, err := p.parsePair()
			if err != nil {
				return nil, err
			}
			dict.Pairs = append(dict.Pairs, pair)
		}
	}

	if p.Match(tokens.Rbrace) == nil {
		return nil, p.Error("Expected }", p.Current())
	}

	return dict, nil
}

func (p *Parser) ParseVariable() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseVariable")

	t := p.Match(tokens.Name)
	if t == nil {
		return nil, p.Error("Expected an identifier.", t)
	}

	switch t.Val {
	case "true", "True":
		br := &nodes.Bool{
			Location: t,
			Val:      true,
		}
		return br, nil
	case "false", "False":
		br := &nodes.Bool{
			Location: t,
			Val:      false,
		}
		return br, nil
	}

	var variable nodes.Node = &nodes.Name{Name: t}

	for !p.Stream.EOF() {
		if dot := p.Match(tokens.Dot); dot != nil {
			getAttr := &nodes.GetAttr{
				Location: dot,
				Node:     variable,
			}
			tok := p.Match(tokens.Name, tokens.Integer)
			tokenType := tokens.Error
			if tok != nil {
				tokenType = tok.Type
			}
			switch tokenType {
			case tokens.Name:
				getAttr.Attr = tok.Val
			case tokens.Integer:
				i, err := strconv.Atoi(tok.Val)
				if err != nil {
					return nil, p.Error(err.Error(), tok)
				}
				getAttr.Index = i
			default:
				return nil, p.Error("This token is not allowed within a variable name.", p.Current())
			}
			variable = getAttr
			continue
		} else if bracket := p.Match(tokens.Lbracket); bracket != nil {
			arg, argErr := p.ParseExpressionWithInlineIfs()
			if argErr != nil {
				return nil, argErr
			}
			getItem := &nodes.GetItem{
				Location: bracket,
				Node:     variable,
				Arg:      &arg,
				Index:    0,
			}

			variable = getItem
			if p.Match(tokens.Rbracket) == nil {
				return nil, p.Error("Unbalanced bracket", bracket)
			}
			continue

		} else if lparen := p.Match(tokens.Lparen); lparen != nil {
			call := &nodes.Call{
				Location: lparen,
				Func:     variable,
				Args:     []nodes.Expression{},
				Kwargs:   map[string]nodes.Expression{},
			}
			// if p.Peek(tokens.VariableEnd) != nil {
			// 	return nil, p.Error("Filter parameter required after '('.", nil)
			// }

			for p.Match(tokens.Comma) != nil || p.Match(tokens.Rparen) == nil {
				// Python allows trailing commas like: f(foo='bar',) - just ignore it
				if p.Match(tokens.Rparen) != nil {
					break
				}

				// TODO: Handle multiple args and kwargs
				v, err := p.ParseExpressionWithInlineIfs()
				if err != nil {
					return nil, err
				}

				if p.Match(tokens.Assign) != nil {
					key := v.Position().Val
					value, errValue := p.ParseExpressionWithInlineIfs()
					if errValue != nil {
						return nil, errValue
					}
					call.Kwargs[key] = value
				} else {
					call.Args = append(call.Args, v)
				}
			}
			variable = call
			// We're done parsing the function call, next variable part
			continue
		}

		// No dot or function call? Then we're done with the variable parsing
		break
	}

	return variable, nil
}

// IDENT | IDENT.(IDENT|NUMBER)...
func (p *Parser) ParseVariableOrLiteral() (nodes.Expression, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseVariableOrLiteral")
	t := p.Current()

	if t == nil {
		return nil, p.Error("Unexpected EOF, expected a number, string, keyword or identifier.", p.Current())
	}

	// Is first part a number or a string, there's nothing to resolve (because there's only to return the value then)
	switch t.Type {
	case tokens.Integer, tokens.Float:
		return p.parseNumber()

	case tokens.String:
		return p.parseString()

	case tokens.Lparen, tokens.Lbrace, tokens.Lbracket:
		return p.parseCollection()

	case tokens.Name:
		return p.ParseVariable()

	default:
		return nil, p.Error("Expected either a number, string, keyword or identifier.", t)
	}
}
