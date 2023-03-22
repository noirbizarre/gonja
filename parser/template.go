package parser

import (
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/tokens"
)

type TemplateParser func(string) (*nodes.Template, error)

func (p *Parser) parseDocElement() (nodes.Node, error) {
	t := p.Current()
	switch t.Type {
	case tokens.Data:
		n := &nodes.Data{
			Data: t,
			Trim: nodes.Trim{
				Left: t.Trim,
			},
		}
		if next := p.Peek(tokens.VariableBegin, tokens.CommentBegin, tokens.BlockBegin); next != nil {
			if len(next.Val) > 0 && next.Val[len(next.Val)-1] == '-' {
				n.Trim.Right = true
			}
		}
		p.Consume()
		return n, nil
	case tokens.EOF:
		p.Consume()
		return nil, nil
	case tokens.CommentBegin:
		return p.ParseComment()
	case tokens.VariableBegin:
		return p.ParseExpressionNode()
	case tokens.BlockBegin:
		return p.ParseStatementBlock()
	}
	return nil, p.Error("Unexpected token (only HTML/tags/filters in templates allowed)", t)
}

func (p *Parser) ParseTemplate() (*nodes.Template, error) {
	tpl := &nodes.Template{
		Name:   p.Name,
		Blocks: nodes.BlockSet{},
		Macros: map[string]*nodes.Macro{},
	}
	p.Template = tpl

	for !p.Stream.End() {
		node, err := p.parseDocElement()
		if err != nil {
			return nil, err
		}
		if node != nil {
			tpl.Nodes = append(tpl.Nodes, node)
		}
	}
	return tpl, nil
}
