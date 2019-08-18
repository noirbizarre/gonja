package parser

import (
	"fmt"
	"strings"

	"github.com/noirbizarre/gonja/config"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/tokens"
)

// The parser provides you a comprehensive and easy tool to
// work with the template document and arguments provided by
// the user for your custom tag.
//
// The parser works on a token list which will be provided by gonja.
// A token is a unit you can work with. Tokens are either of type identifier,
// string, number, keyword, HTML or symbol.
//
// (See Token's documentation for more about tokens)
type Parser struct {
	Name   string
	Stream *tokens.Stream
	Config *config.Config

	Template       *nodes.Template
	Statements     map[string]StatementParser
	Level          int8
	TemplateParser TemplateParser
}

// Creates a new parser to parse tokens.
// Used inside gonja to parse documents and to provide an easy-to-use
// parser for tag authors
// func NewParser(name string, tokens []*tokens.Token) *Parser {
// 	p := &Parser{
// 		name:          name,
// 		tokens:        tokens,
// 		template:      template,
// 		// bannedStmts:   make(map[string]bool),
// 		// bannedFilters: make(map[string]bool),
// 	}
// 	if len(tokens) > 0 {
// 		p.lastToken = tokens[len(tokens)-1]
// 	}
// 	return p
// }

func NewParser(name string, cfg *config.Config, stream *tokens.Stream) *Parser {
	return &Parser{
		Name:   name,
		Stream: stream,
		Config: cfg,
	}
}

func Parse(input string) (*nodes.Template, error) {
	stream := tokens.Lex(input)
	p := NewParser("parser", config.DefaultConfig, stream)
	return p.Parse()
}

func (p *Parser) Parse() (*nodes.Template, error) {
	// for p.state = parseProg; p.state != nil; {
	// 	p.state = p.state(p)
	// }

	// lex everything
	// t := p.Lexer.NextItem()
	// for ; t.Typ != lex.EOF; t = p.Lexer.NextItem() {
	// 	p.Items = append(p.Items, t)
	// }
	// p.Items = append(p.Items, t)

	// tokens := []*l.Token{}
	// for token := range p.Tokens {
	// 	p.tokens = append(p.tokens, token)
	// }

	return p.ParseTemplate()
}

// Consume one token. It will be gone forever.
func (p *Parser) Consume() {
	p.Stream.Next()
}

// // Consume N tokens. They will be gone forever.
// func (p *Parser) ConsumeN(count int) {
// 	p.idx += count
// }

// Current returns the current token.
func (p *Parser) Current() *tokens.Token {
	return p.Stream.Current()
}

// Next returns and consume the current token
func (p *Parser) Next() *tokens.Token {
	// t := p.Stream.Next()
	// p.Consume()
	// return t
	return p.Stream.Next()
}

func (p *Parser) End() bool {
	return p.Stream.End()
}

// Match returns the CURRENT token if the given type matches.
// Consumes this token on success.
func (p *Parser) Match(types ...tokens.Type) *tokens.Token {
	tok := p.Stream.Current()
	for _, t := range types {
		if tok.Type == t {
			p.Stream.Next()
			return tok
		}
	}
	return nil
}

func (p *Parser) MatchName(names ...string) *tokens.Token {
	t := p.Peek(tokens.Name)
	if t != nil {
		for _, name := range names {
			if t.Val == name {
				return p.Pop()
			}
		}
	}
	// if t != nil && t.Val == name { return p.Pop() }
	return nil
}

// Pop returns the current token and advance to the next
func (p *Parser) Pop() *tokens.Token {
	t := p.Stream.Current()
	p.Stream.Next()
	return t
}

// Peek returns the next token without consuming the current
// if it matches one of the given types
func (p *Parser) Peek(types ...tokens.Type) *tokens.Token {
	tok := p.Stream.Current()
	for _, t := range types {
		if tok.Type == t {
			return tok
		}
	}
	return nil
}

func (p *Parser) PeekName(names ...string) *tokens.Token {
	t := p.Peek(tokens.Name)
	if t != nil {
		for _, name := range names {
			if t.Val == name {
				return t
			}
		}
	}
	// if t != nil && t.Val == name { return t }
	return nil
}

// WrapUntil wraps all nodes between starting tag and "{% endtag %}" and provides
// one simple interface to execute the wrapped nodes.
// It returns a parser to process provided arguments to the tag.
func (p *Parser) WrapUntil(names ...string) (*nodes.Wrapper, *Parser, error) {
	wrapper := &nodes.Wrapper{
		Location: p.Current(),
		Trim:     &nodes.Trim{},
	}

	var args []*tokens.Token

	for !p.Stream.End() {
		// New tag, check whether we have to stop wrapping here
		if begin := p.Match(tokens.BlockBegin); begin != nil {
			ident := p.Peek(tokens.Name)

			if ident != nil {
				// We've found a (!) end-tag

				found := false
				for _, n := range names {
					if ident.Val == n {
						found = true
						break
					}
				}

				// We only process the tag if we've found an end tag
				if found {
					// Okay, endtag found.
					p.Consume() // '{%' tagname
					wrapper.Trim.Left = begin.Val[len(begin.Val)-1] == '-'
					wrapper.LStrip = begin.Val[len(begin.Val)-1] == '+'

					for {
						if end := p.Match(tokens.BlockEnd); end != nil {
							// Okay, end the wrapping here
							wrapper.EndTag = ident.Val
							wrapper.Trim.Right = end.Val[0] == '-'
							stream := tokens.NewStream(args)
							return wrapper, NewParser(p.Name, p.Config, stream), nil
						}
						t := p.Next()
						// p.Consume()
						if t == nil {
							return nil, nil, p.Error("Unexpected EOF.", p.Current())
						}
						args = append(args, t)
					}
				}
			}
			p.Stream.Backup()
		}

		// Otherwise process next element to be wrapped
		node, err := p.parseDocElement()
		if err != nil {
			return nil, nil, err
		}
		wrapper.Nodes = append(wrapper.Nodes, node)
	}

	return nil, nil, p.Error(fmt.Sprintf("Unexpected EOF, expected tag %s.", strings.Join(names, " or ")),
		p.Current())
}

// Skips all nodes between starting tag and "{% endtag %}"
func (p *Parser) SkipUntil(names ...string) error {
	for !p.End() {
		// New tag, check whether we have to stop wrapping here
		if p.Match(tokens.BlockBegin) != nil {
			ident := p.Peek(tokens.Name)

			if ident != nil {
				// We've found a (!) end-tag

				found := false
				for _, n := range names {
					if ident.Val == n {
						found = true
						break
					}
				}

				// We only process the tag if we've found an end tag
				if found {
					// Okay, endtag found.
					p.Consume() // '{%' tagname

					for {
						if p.Match(tokens.BlockEnd) != nil {
							// Done skipping, exit.
							return nil
						}
					}
				}
			} else {
				p.Stream.Backup()
			}
		}
		t := p.Next()
		if t == nil {
			return p.Error("Unexpected EOF.", p.Current())
		}
	}

	return p.Error(fmt.Sprintf("Unexpected EOF, expected tag %s.", strings.Join(names, " or ")), p.Current())
}
