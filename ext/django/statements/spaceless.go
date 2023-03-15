package statements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nikolalohinski/gonja/exec"
	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
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
	err := sub.ExecuteWrapper(stmt.wrapper)
	if err != nil {
		return err
	}

	s := out.String()
	// Repeat this recursively
	changed := true
	for changed {
		s2 := spacelessRegexp.ReplaceAllString(s, "$1$3")
		changed = s != s2
		s = s2
	}

	r.WriteString(s)

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
	All.Register("spaceless", spacelessParser)
}
