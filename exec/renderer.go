package exec

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/paradime-io/gonja/nodes"
)

// TrimState stores and apply trim policy
type TrimState struct {
	Should      bool
	ShouldBlock bool
	Buffer      *strings.Builder
}

func (ts *TrimState) TrimBlocks(r rune) bool {
	if ts.ShouldBlock {
		switch r {
		case '\n':
			ts.ShouldBlock = false
			return true
		case ' ', '\t':
			return true
		default:
			return false
		}
	}
	return false
}

// Renderer is a node visitor in charge of rendering
type Renderer struct {
	*EvalConfig
	Ctx      *Context
	Template *Template
	Root     *nodes.Template
	Out      *strings.Builder
	Trim     *TrimState
}

// NewRenderer initialize a new renderer
func NewRenderer(ctx *Context, out *strings.Builder, cfg *EvalConfig, tpl *Template) *Renderer {
	var buffer strings.Builder
	r := &Renderer{
		EvalConfig: cfg,
		Ctx:        ctx,
		Template:   tpl,
		Root:       tpl.Root,
		Out:        out,
		Trim:       &TrimState{Buffer: &buffer},
	}
	r.Ctx.Set("self", Self(r))
	return r
}

// Inherit creates a new sub renderer
func (r *Renderer) Inherit() *Renderer {
	sub := &Renderer{
		EvalConfig: r.EvalConfig.Inherit(),
		Ctx:        r.Ctx.Inherit(),
		Template:   r.Template,
		Root:       r.Root,
		Out:        r.Out,
		Trim:       r.Trim,
	}
	return sub
}

func (r *Renderer) Flush(lstrip bool) {
	r.FlushAndTrim(false, lstrip)
}

func (r *Renderer) FlushAndTrim(trim, lstrip bool) {
	txt := r.Trim.Buffer.String()
	if r.Config.LstripBlocks && !lstrip {
		lines := strings.Split(txt, "\n")
		last := lines[len(lines)-1]
		lines[len(lines)-1] = strings.TrimLeft(last, " \t")
		txt = strings.Join(lines, "\n")
	}
	if trim {
		txt = strings.TrimRight(txt, " \t\n")
	}
	r.Out.WriteString(txt)
	r.Trim.Buffer.Reset()
}

// WriteString wraps the trimming policy
func (r *Renderer) WriteString(txt string) (int, error) {
	if r.Config.TrimBlocks {
		txt = strings.TrimLeftFunc(txt, r.Trim.TrimBlocks)
	}
	if r.Trim.Should {
		txt = strings.TrimLeft(txt, " \t\n")
		if len(txt) > 0 {
			r.Trim.Should = false
		}
	}
	return r.Trim.Buffer.WriteString(txt)
}

// RenderValue properly render a value
func (r *Renderer) RenderValue(value *Value) {
	if r.Autoescape && value.IsString() && !value.Safe {
		r.WriteString(value.Escaped())
	} else {
		r.WriteString(value.String())
	}
}

func (r *Renderer) StartTag(trim *nodes.Trim, lstrip bool) {
	if trim == nil {
		r.Flush(lstrip)
	} else {
		r.FlushAndTrim(trim.Left, lstrip)
	}
	r.Trim.Should = false
}

func (r *Renderer) EndTag(trim *nodes.Trim) {
	if trim == nil {
		return
	}
	r.Trim.Should = trim.Right
}

func (r *Renderer) Tag(trim *nodes.Trim, lstrip bool) {
	r.StartTag(trim, lstrip)
	r.EndTag(trim)
}

// Visit implements the nodes.Visitor interface
func (r *Renderer) Visit(node nodes.Node) (nodes.Visitor, error) {
	switch n := node.(type) {
	case *nodes.Comment:
		r.Tag(n.Trim, false)
		return nil, nil
	case *nodes.Data:
		r.WriteString(n.Data.Val)
		return nil, nil
	case *nodes.Output:
		r.StartTag(n.Trim, false)
		value := r.Eval(n.Expression)
		if value.IsError() {
			return nil, errors.Wrapf(value, `Unable to render expression '%s'`, n.Expression)
		}
		r.RenderValue(value)
		r.EndTag(n.Trim)
		return nil, nil
	case *nodes.StatementBlock:
		r.Tag(n.Trim, n.LStrip)
		r.Trim.ShouldBlock = r.Config.TrimBlocks
		stmt, ok := n.Stmt.(Statement)
		if ok {
			// Silently ignore non executable statements
			// return nil, nil
			// return nil, errors.Errorf(`Unable to execute statement '%s'`, n.Stmt)
			if err := stmt.Execute(r, n); err != nil {
				return nil, errors.Wrapf(err, `Unable to execute statement '%s'`, n.Stmt)
			}
		}
		return nil, nil
	default:
		return r, nil
	}
}

// ExecuteWrapper wraps the nodes.Wrapper execution logic
func (r *Renderer) ExecuteWrapper(wrapper *nodes.Wrapper) error {
	sub := r.Inherit()
	err := nodes.Walk(sub, wrapper)
	sub.Tag(wrapper.Trim, wrapper.LStrip)
	r.Trim.ShouldBlock = r.Config.TrimBlocks
	return err
}

func (r *Renderer) LStrip() {
}

func (r *Renderer) Execute() error {
	// Determine the parent to be executed (for template inheritance)
	root := r.Root
	for root.Parent != nil {
		root = root.Parent
	}

	err := nodes.Walk(r, root)
	if err == nil {
		r.Flush(false)
	}
	return err
}

func (r *Renderer) String() string {
	r.Flush(false)
	out := r.Out.String()
	if !r.Config.KeepTrailingNewline {
		out = strings.TrimSuffix(out, "\n")
	}
	return out
}
