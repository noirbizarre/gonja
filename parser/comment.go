package parser

import (
	"fmt"

	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/tokens"
	log "github.com/sirupsen/logrus"
)

func (p *Parser) ParseComment() (*nodes.Comment, error) {
	log.WithFields(log.Fields{
		"current": p.Current(),
	}).Trace("ParseComment")

	tok := p.Match(tokens.CommentBegin)
	if tok == nil {
		msg := fmt.Sprintf(`Expected '%s' , got %s`, p.Config.CommentStartString, p.Current())
		return nil, p.Error(msg, p.Current())
	}

	comment := &nodes.Comment{
		Start: tok,
		Trim:  &nodes.Trim{},
	}

	tok = p.Match(tokens.Data)
	if tok == nil {
		comment.Text = ""
	} else {
		comment.Text = tok.Val
	}

	tok = p.Match(tokens.CommentEnd)
	if tok == nil {
		msg := fmt.Sprintf(`Expected '%s' , got %s`, p.Config.CommentEndString, p.Current())
		return nil, p.Error(msg, p.Current())
	}
	comment.End = tok

	log.WithFields(log.Fields{
		"node": comment,
	}).Trace("ParseComment return")
	return comment, nil
}
