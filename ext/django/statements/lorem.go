package statements

import (
	"fmt"
	"math/rand"
	"time"

	// "github.com/pkg/errors"

	"github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/parser"
	"github.com/noirbizarre/gonja/tokens"
	"github.com/noirbizarre/gonja/utils"
	"github.com/pkg/errors"
)

type LoremStmt struct {
	Location *tokens.Token
	count    int    // number of paragraphs
	method   string // w = words, p = HTML paragraphs, b = plain-text (default is b)
	random   bool   // does not use the default paragraph "Lorem ipsum dolor sit amet, ..."
}

func (stmt *LoremStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *LoremStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("LoremStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *LoremStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	lorem, err := utils.Lorem(stmt.count, stmt.method)
	if err != nil {
		return errors.Wrap(err, `Unable to execute 'lorem' statement`)
	}
	if _, err = r.WriteString(lorem); err != nil {
		return errors.Wrap(err, `Unable to execute 'lorem' statement`)
	}

	return nil
}

func loremParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &LoremStmt{
		Location: p.Current(),
		count:    1,
		method:   "b",
	}

	if countToken := args.Match(tokens.Integer); countToken != nil {
		stmt.count = exec.AsValue(countToken.Val).Integer()
	}

	if methodToken := args.Match(tokens.Name); methodToken != nil {
		if methodToken.Val != "w" && methodToken.Val != "p" && methodToken.Val != "b" {
			return nil, args.Error("lorem-method must be either 'w', 'p' or 'b'.", nil)
		}

		stmt.method = methodToken.Val
	}

	if args.MatchName("random") != nil {
		stmt.random = true
	}

	if !args.End() {
		return nil, args.Error("Malformed lorem-tag args.", nil)
	}

	return stmt, nil
}

func init() {
	rand.Seed(time.Now().Unix())

	All.MustRegister("lorem", loremParser)
}
