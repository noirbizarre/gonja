package statements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/parser"
	"github.com/noirbizarre/gonja/tokens"
	"github.com/pkg/errors"
)

type SpacelessStmt struct {
	Location *tokens.Token
	wrapper  *nodes.Wrapper
}

func (stmt *SpacelessStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *SpacelessStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("SpacelessStmt(Line=%d Col=%d)", t.Line, t.Col)
}

var spacelessRegexp = regexp.MustCompile(`(?U:(<.*>))([\t\n\v\f\r ]+)(?U:(<.*>))`)

func (stmt *SpacelessStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	var out strings.Builder

	sub := r.Inherit()
	sub.Out = &out
	if err := sub.ExecuteWrapper(stmt.wrapper); err != nil {
		return errors.Wrap(err, `Unable to execute 'spaceless' statement`)
	}

	s := out.String()
	// Repeat this recursively
	changed := true
	for changed {
		s2 := spacelessRegexp.ReplaceAllString(s, "$1$3")
		changed = s != s2
		s = s2
	}

	if _, err := r.WriteString(s); err != nil {
		return errors.Wrap(err, `Unable to execute 'spaceless' statement`)
	}

	return nil
}

func spacelessParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &SpacelessStmt{
		Location: p.Current(),
	}

	wrapper, _, err := p.WrapUntil("endspaceless")
	if err != nil {
		return nil, err
	}
	stmt.wrapper = wrapper

	if !args.End() {
		return nil, args.Error("Malformed spaceless-tag args.", nil)
	}

	return stmt, nil
}

func init() {
	All.MustRegister("spaceless", spacelessParser)
}
