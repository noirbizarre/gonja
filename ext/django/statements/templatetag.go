package statements

import (
	"fmt"

	"github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/parser"
	"github.com/noirbizarre/gonja/tokens"
	"github.com/pkg/errors"
)

type TemplateTagStmt struct {
	Location *tokens.Token
	content  string
}

func (stmt *TemplateTagStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *TemplateTagStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("TemplateTagStmt(Line=%d Col=%d)", t.Line, t.Col)
}

var templateTagMapping = map[string]string{
	"openblock":     "{%",
	"closeblock":    "%}",
	"openvariable":  "{{",
	"closevariable": "}}",
	"openbrace":     "{",
	"closebrace":    "}",
	"opencomment":   "{#",
	"closecomment":  "#}",
}

func (stmt *TemplateTagStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	if _, err := r.WriteString(stmt.content); err != nil {
		return errors.Wrap(err, `Unable to execute 'templatetag' statement`)
	}
	return nil
}

func templateTagParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &TemplateTagStmt{}

	if argToken := args.Match(tokens.Name); argToken != nil {
		output, found := templateTagMapping[argToken.Val]
		if !found {
			return nil, args.Error("Argument not found", argToken)
		}
		stmt.content = output
	} else {
		return nil, args.Error("Identifier expected.", nil)
	}

	if !args.End() {
		return nil, args.Error("Malformed templatetag-tag argument.", nil)
	}

	return stmt, nil
}

func init() {
	All.MustRegister("templatetag", templateTagParser)
}
