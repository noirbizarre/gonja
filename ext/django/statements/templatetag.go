package statements

import (
	"fmt"

	"github.com/paradime-io/gonja/exec"
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/parser"
	"github.com/paradime-io/gonja/tokens"
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

func (node *TemplateTagStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	r.WriteString(node.content)
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
	All.Register("templatetag", templateTagParser)
}
