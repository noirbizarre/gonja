package statements

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/paradime-io/gonja/exec"
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/parser"
	"github.com/paradime-io/gonja/tokens"
)

type BlockStmt struct {
	Location *tokens.Token
	Name     string
}

func (stmt *BlockStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *BlockStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("BlockStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *BlockStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	// root, block := r.Root.GetBlock(stmt.Name)
	blocks := r.Root.GetBlocks(stmt.Name)
	block, blocks := blocks[0], blocks[1:]

	if block == nil {
		return errors.Errorf(`Unable to find block "%s"`, stmt.Name)
	}

	sub := r.Inherit()
	infos := &BlockInfos{Block: stmt, Renderer: sub, Blocks: blocks}

	sub.Ctx.Set("super", infos.super)
	sub.Ctx.Set("self", exec.Self(sub))

	err := sub.ExecuteWrapper(block)
	if err != nil {
		return err
	}

	return nil
}

type BlockInfos struct {
	Block    *BlockStmt
	Renderer *exec.Renderer
	Blocks   []*nodes.Wrapper
	Root     *nodes.Template
}

func (bi *BlockInfos) super() string {
	if len(bi.Blocks) <= 0 {
		return ""
	}
	r := bi.Renderer
	block, blocks := bi.Blocks[0], bi.Blocks[1:]
	sub := r.Inherit()
	var out strings.Builder
	sub.Out = &out
	infos := &BlockInfos{
		Block:    bi.Block,
		Renderer: sub,
		Blocks:   blocks,
	}
	sub.Ctx.Set("self", exec.Self(sub))
	sub.Ctx.Set("super", infos.super)
	sub.ExecuteWrapper(block)
	return out.String()
}

func blockParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	block := &BlockStmt{
		Location: p.Current(),
	}
	if args.End() {
		return nil, errors.New("Tag 'block' requires an identifier.")
	}

	name := args.Match(tokens.Name)
	if name == nil {
		return nil, errors.New("First argument for tag 'block' must be an identifier.")
	}

	if !args.End() {
		return nil, errors.New("Tag 'block' takes exactly 1 argument (an identifier).")
	}

	wrapper, endArgs, err := p.WrapUntil("endblock")
	if err != nil {
		return nil, err
	}
	if !endArgs.End() {
		endName := endArgs.Match(tokens.Name)
		if endName != nil {
			if name.Val != endName.Val {
				return nil, errors.Errorf(`Name for 'endblock' must equal to 'block'-tag's name ('%s' != '%s').`,
					name.Val, endName.Val)
			}
		}

		if endName == nil || !endArgs.End() {
			return nil, errors.New("Either no or only one argument (identifier) allowed for 'endblock'.")
		}
	}

	if !p.Template.Blocks.Exists(name.Val) {
		p.Template.Blocks.Register(name.Val, wrapper)
	} else {
		return nil, args.Error(fmt.Sprintf("Block named '%s' already defined", name.Val), nil)
	}

	block.Name = name.Val
	return block, nil
}

func init() {
	All.Register("block", blockParser)
}
