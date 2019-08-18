package parser

import (
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/tokens"
)

type TemplateParser func(string) (*nodes.Template, error)

// Doc = { ( Filter | Tag | HTML ) }
func (p *Parser) parseDocElement() (nodes.Node, error) {
	t := p.Current()

	switch t.Type {
	case tokens.Data:
		n := &nodes.Data{Data: t}
		p.Consume() // consume HTML element
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
