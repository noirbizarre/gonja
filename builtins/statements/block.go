package statements

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/noirbizarre/gonja/exec"
	"github.com/noirbizarre/gonja/nodes"
	"github.com/noirbizarre/gonja/parser"
	"github.com/noirbizarre/gonja/tokens"
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

func (bi *BlockInfos) super() (string, error) {
	if len(bi.Blocks) <= 0 {
		return "", errors.New("super() can only be used in child templates")
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
	if err := sub.ExecuteWrapper(block); err != nil {
		return "", errors.Wrap(err, "Unable to render parent block")
	}
	return out.String(), nil
}

func blockParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	block := &BlockStmt{
		Location: p.Current(),
	}
	if args.End() {
		return nil, errors.New("Tag 'block' requires an identifier")
	}

	name := args.Match(tokens.Name)
	if name == nil {
		return nil, errors.New("First argument for tag 'block' must be an identifier")
	}

	if !args.End() {
		return nil, errors.New("Tag 'block' takes exactly 1 argument (an identifier)")
	}

	wrapper, endargs, err := p.WrapUntil("endblock")
	if err != nil {
		return nil, err
	}
	if !endargs.End() {
		endName := endargs.Match(tokens.Name)
		if endName != nil {
			if endName.Val != name.Val {
				return nil, errors.Errorf(`Name for 'endblock' must equal to 'block'-tag's name ('%s' != '%s').`,
					name.Val, endName.Val)
			}
		}

		if endName == nil || !endargs.End() {
			return nil, errors.New("Either no or only one argument (identifier) allowed for 'endblock'")
		}
	}

	if !p.Template.Blocks.Exists(name.Val) {
		if err = p.Template.Blocks.Register(name.Val, wrapper); err != nil {
			msg := fmt.Sprintf("Error while registering block named '%s': %s", name.Val, err)
			return nil, args.Error(msg, block.Location)
		}
	} else {
		msg := fmt.Sprintf("Block named '%s' already defined", name.Val)
		return nil, args.Error(msg, block.Location)
	}

	block.Name = name.Val
	return block, nil
}

func init() {
	All.MustRegister("block", blockParser)
}
