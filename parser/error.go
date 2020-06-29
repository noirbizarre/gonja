package parser

import (
	"github.com/goph/emperror"
	"github.com/noirbizarre/gonja/tokens"
	"github.com/pkg/errors"
)

// Error produces a nice error message and returns an error-object.
// The 'token'-argument is optional. If provided, it will take
// the token's position information. If not provided, it will
// automatically use the CURRENT token's position information.
func (p *Parser) Error(msg string, token *tokens.Token) error {
	if token == nil {
		return errors.New(msg)
	}
	return emperror.With(
		errors.Errorf(`%s (Line: %d Col: %d, near "%s")`, msg, token.Line, token.Col, token.Val),
		"token", token,
	)
}
