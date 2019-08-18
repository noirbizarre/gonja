package parser

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/tokens"
)

type StatementParser func(parser *Parser, args *Parser) (nodes.Statement, error)

// Tag = "{%" IDENT ARGS "%}"
func (p *Parser) ParseStatement() (nodes.Statement, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseStatement")

	if p.Match(tokens.BlockBegin) == nil {
		return nil, p.Error("'{%' expected here", p.Current())
	}

	name := p.Match(tokens.Name)
	if name == nil {
		return nil, p.Error("Expected a statement name here", p.Current())
	}

	// Check for the existing statement
	stmtParser, exists := p.Statements[name.Val]
	if !exists {
		// Does not exists
		return nil, p.Error(fmt.Sprintf("Statement '%s' not found (or beginning not provided)", name.Val), name)
	}

	// Check sandbox tag restriction
	// if _, isBanned := p.bannedStmts[tokenName.Val]; isBanned {
	// 	return nil, p.Error(fmt.Sprintf("Usage of statement '%s' is not allowed (sandbox restriction active).", tokenName.Val), tokenName)
	// }

	var args []*tokens.Token
	for p.Peek(tokens.BlockEnd) == nil && !p.Stream.End() {
		// Add token to args
		args = append(args, p.Next())
		// p.Consume() // next token
	}

	// EOF?
	// if p.Remaining() == 0 {
	// 	return nil, p.Error("Unexpectedly reached EOF, no statement end found.", p.lastToken)
	// }

	if p.Match(tokens.BlockEnd) == nil {
		return nil, p.Error(fmt.Sprintf(`Expected end of block "%s"`, p.Config.BlockEndString), p.Current())
	}

	argParser := NewParser("statement", p.Config, tokens.NewStream(args))
	// argParser := newParser(p.name, argsToken, p.template)
	// if len(argsToken) == 0 {
	// 	// This is done to have nice EOF error messages
	// 	argParser.lastToken = tokenName
	// }

	p.Level++
	defer func() { p.Level-- }()
	return stmtParser(p, argParser)
}

// type StatementParser func(parser *Parser, args *Parser) (nodes.Stmt, error)

func (p *Parser) ParseStatementBlock() (*nodes.StatementBlock, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseStatementBlock")

	begin := p.Match(tokens.BlockBegin)
	if begin == nil {
		return nil, errors.Errorf(`Expected "%s" got "%s"`, p.Config.BlockStartString, p.Current())
	}

	name := p.Match(tokens.Name)
	if name == nil {
		return nil, p.Error("Expected a statement name here", p.Current())
	}

	// Check for the existing statement
	stmtParser, exists := p.Statements[name.Val]
	if !exists {
		// Does not exists
		return nil, p.Error(fmt.Sprintf("Statement '%s' not found (or beginning not provided)", name.Val), name)
	}

	// Check sandbox tag restriction
	// if _, isBanned := p.bannedStmts[tokenName.Val]; isBanned {
	// 	return nil, p.Error(fmt.Sprintf("Usage of statement '%s' is not allowed (sandbox restriction active).", tokenName.Val), tokenName)
	// }

	log.Trace("args")
	var args []*tokens.Token
	for p.Peek(tokens.BlockEnd) == nil && !p.Stream.End() {
		log.Trace("for args")
		// Add token to args
		args = append(args, p.Next())
		// p.Consume() // next token
	}
	log.Trace("loop ended")

	// EOF?
	// if p.Remaining() == 0 {
	// 	return nil, p.Error("Unexpectedly reached EOF, no statement end found.", p.lastToken)
	// }

	end := p.Match(tokens.BlockEnd)
	if end == nil {
		return nil, p.Error(fmt.Sprintf(`Expected end of block "%s"`, p.Config.BlockEndString), p.Current())
	}
	log.WithFields(log.Fields{
		"args": args,
	}).Trace("Matched end block")

	stream := tokens.NewStream(args)
	log.WithFields(log.Fields{
		"stream": stream,
	}).Trace("Got stream")
	argParser := NewParser(fmt.Sprintf("%s:args", name.Val), p.Config, stream)
	log.Trace("argparser")
	// argParser := newParser(p.name, argsToken, p.template)
	// if len(argsToken) == 0 {
	// 	// This is done to have nice EOF error messages
	// 	argParser.lastToken = tokenName
	// }

	// p.template.level++
	// defer func() { p.template.level-- }()
	stmt, err := stmtParser(p, argParser)
	if err != nil {
		return nil, errors.Wrapf(err, `Unable to parse statement "%s"`, name.Val)
	}
	log.Trace("got stmt and return")
	return &nodes.StatementBlock{
		Location: begin,
		Name:     name.Val,
		Stmt:     stmt,
		LStrip:   begin.Val[len(begin.Val)-1] == '+',
		Trim: &nodes.Trim{
			Left:  begin.Val[len(begin.Val)-1] == '-',
			Right: end.Val[0] == '-',
		},
	}, nil
}
