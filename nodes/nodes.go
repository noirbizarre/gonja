package nodes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/noirbizarre/gonja/tokens"
	u "github.com/noirbizarre/gonja/utils"
)

// ----------------------------------------------------------------------------
// Interfaces
//
// There are 3 main classes of nodes: Expressions and type nodes,
// statement nodes, and declaration nodes. The node names usually
// match the corresponding Go spec production names to which they
// correspond. The node fields correspond to the individual parts
// of the respective productions.
//
// All nodes contain position information marking the beginning of
// the corresponding source text segment; it is accessible via the
// Pos accessor method. Nodes may contain additional position info
// for language constructs where comments may be found between parts
// of the construct (typically any larger, parenthesized subpart).
// That position information is needed to properly position comments
// when printing the construct.

// All node types implement the Node interface.
type Node interface {
	fmt.Stringer
	Position() *tokens.Token
}

// Expression represents an evaluable expression part
type Expression interface {
	Node
}

// Statement represents a statement block "{% %}"
type Statement interface {
	Node
}

// Template is the root node of any template
type Template struct {
	Name   string
	Nodes  []Node
	Blocks BlockSet
	Macros map[string]*Macro
	Parent *Template
}

func (tpl *Template) Position() *tokens.Token { return tpl.Nodes[0].Position() }
func (tpl *Template) String() string {
	tok := tpl.Position()
	return fmt.Sprintf("Template(Name=%s Line=%d Col=%d)", tpl.Name, tok.Line, tok.Col)
}

func (tpl *Template) GetBlocks(name string) []*Wrapper {
	var blocks []*Wrapper
	if tpl.Parent != nil {
		blocks = tpl.Parent.GetBlocks(name)
	} else {
		blocks = []*Wrapper{}
	}
	block, exists := tpl.Blocks[name]
	if exists {
		blocks = append([]*Wrapper{block}, blocks...)
	}
	return blocks
}

type Trim struct {
	Left  bool
	Right bool
}

type Data struct {
	Data *tokens.Token // data token
}

func (d *Data) Position() *tokens.Token { return d.Data }

func (d *Data) String() string {
	return fmt.Sprintf("Data(text=%s Line=%d Col=%d)",
		u.Ellipsis(d.Data.Val, 20), d.Data.Line, d.Data.Col)
}

// A Comment node represents a single {# #} comment.
type Comment struct {
	Start *tokens.Token // Opening token
	Text  string        // Comment text
	End   *tokens.Token // Closing token
	Trim  *Trim
}

func (c *Comment) Position() *tokens.Token { return c.Start }

// func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }
func (c *Comment) String() string {
	return fmt.Sprintf("Comment(text=%s Line=%d Col=%d)",
		u.Ellipsis(c.Text, 20), c.Start.Line, c.Start.Col)
}

// Ouput represents a printable expression node {{ }}
type Output struct {
	Start      *tokens.Token
	Expression Expression
	End        *tokens.Token
	Trim       *Trim
}

func (o *Output) Position() *tokens.Token { return o.Start }
func (o *Output) String() string {
	return fmt.Sprintf("Output(Expression=%s Line=%d Col=%d)",
		o.Expression, o.Start.Line, o.End.Col)
}

type FilteredExpression struct {
	Expression Expression
	Filters    []*FilterCall
}

func (expr *FilteredExpression) Position() *tokens.Token {
	return expr.Expression.Position()
}
func (expr *FilteredExpression) String() string {
	t := expr.Expression.Position()

	return fmt.Sprintf("FilteredExpression(Expression=%s Line=%d Col=%d)",
		expr.Expression, t.Line, t.Col)
	// return fmt.Sprintf("<FilteredExpression Expression=%s", expr.Expression)
}

type FilterCall struct {
	Token *tokens.Token

	Name   string
	Args   []Expression
	Kwargs map[string]Expression

	// filterFunc FilterFunction
}

type TestExpression struct {
	Expression Expression
	Test       *TestCall
}

func (expr *TestExpression) String() string {
	t := expr.Position()

	return fmt.Sprintf("TestExpression(Expression=%s Test=%s Line=%d Col=%d)",
		expr.Expression, expr.Test, t.Line, t.Col)
	// return fmt.Sprintf("TestExpression(Expression=%s Test=%s)",
	// 	expr.Expression, expr.Test)
}
func (expr *TestExpression) Position() *tokens.Token {
	return expr.Expression.Position()
}

type TestCall struct {
	Token *tokens.Token

	Name   string
	Args   []Expression
	Kwargs map[string]Expression

	// testFunc TestFunction
}

func (tc *TestCall) String() string {
	return fmt.Sprintf("TestCall(name=%s Line=%d Col=%d)",
		tc.Name, tc.Token.Line, tc.Token.Col)
}

type String struct {
	Location *tokens.Token
	Val      string
}

func (s *String) Position() *tokens.Token { return s.Location }
func (s *String) String() string          { return s.Location.Val }

type Integer struct {
	Location *tokens.Token
	Val      int
}

func (i *Integer) Position() *tokens.Token { return i.Location }
func (i *Integer) String() string          { return i.Location.Val }

type Float struct {
	Location *tokens.Token
	Val      float64
}

func (f *Float) Position() *tokens.Token { return f.Location }
func (f *Float) String() string          { return f.Location.Val }

type Bool struct {
	Location *tokens.Token
	Val      bool
}

func (b *Bool) Position() *tokens.Token { return b.Location }
func (b *Bool) String() string          { return b.Location.Val }

type Name struct {
	Name *tokens.Token
}

func (n *Name) Position() *tokens.Token { return n.Name }
func (n *Name) String() string {
	t := n.Position()
	return fmt.Sprintf("Name(Val=%s Line=%d Col=%d)", t.Val, t.Line, t.Col)
}

type List struct {
	Location *tokens.Token
	Val      []Expression
}

func (l *List) Position() *tokens.Token { return l.Location }
func (l *List) String() string          { return l.Location.Val }

type Tuple struct {
	Location *tokens.Token
	Val      []Expression
}

func (t *Tuple) Position() *tokens.Token { return t.Location }
func (t *Tuple) String() string          { return t.Location.Val }

type Dict struct {
	Token *tokens.Token
	Pairs []*Pair
}

func (d *Dict) Position() *tokens.Token { return d.Token }
func (d *Dict) String() string          { return d.Token.Val }

type Pair struct {
	Key   Expression
	Value Expression
}

func (p *Pair) Position() *tokens.Token { return p.Key.Position() }
func (p *Pair) String() string {
	t := p.Position()
	return fmt.Sprintf("Pair(Key=%s Value=%s Line=%d Col=%d)", p.Key, p.Value, t.Line, t.Col)
}

type Variable struct {
	Location *tokens.Token

	Parts []*VariablePart
}

func (v *Variable) Position() *tokens.Token { return v.Location }
func (v *Variable) String() string {
	parts := make([]string, 0, len(v.Parts))
	for _, p := range v.Parts {
		switch p.Type {
		case VarTypeInt:
			parts = append(parts, strconv.Itoa(p.I))
		case VarTypeIdent:
			parts = append(parts, p.S)
		default:
			panic("unimplemented")
		}
	}
	return strings.Join(parts, ".")
}

const (
	VarTypeInt = iota
	VarTypeIdent
)

type VariablePart struct {
	Type int
	S    string
	I    int

	IsFunctionCall bool
	// callingArgs    []functionCallArgument // needed for a function call, represents all argument nodes (Node supports nested function calls)
	Args   []Expression
	Kwargs map[string]Expression
}

func (vp *VariablePart) String() string {
	return fmt.Sprintf("VariablePart(S=%s I=%d)", vp.S, vp.I)
}

type Call struct {
	Location *tokens.Token
	Func     Node
	Args     []Expression
	Kwargs   map[string]Expression
}

func (c *Call) Position() *tokens.Token { return c.Location }
func (c *Call) String() string {
	t := c.Position()
	return fmt.Sprintf("Call(Args=%s Kwargs=%s Line=%d Col=%d)", c.Args, c.Kwargs, t.Line, t.Col)
}

type Getitem struct {
	Location *tokens.Token
	Node     Node
	Arg      string
	Index    int
}

func (g *Getitem) Position() *tokens.Token { return g.Location }
func (g *Getitem) String() string {
	t := g.Position()
	var param string
	if g.Arg != "" {
		param = fmt.Sprintf(`Arg=%s`, g.Arg)
	} else {
		param = fmt.Sprintf(`Index=%s`, strconv.Itoa(g.Index))
	}
	return fmt.Sprintf("Getitem(Node=%s %s Line=%d Col=%d)", g.Node, param, t.Line, t.Col)
}

type Getattr struct {
	Location *tokens.Token
	Node     Node
	Attr     string
	Index    int
}

func (g *Getattr) Position() *tokens.Token { return g.Location }
func (g *Getattr) String() string {
	t := g.Position()
	var param string
	if g.Attr != "" {
		param = fmt.Sprintf(`Attr=%s`, g.Attr)
	} else {
		param = fmt.Sprintf(`Index=%s`, strconv.Itoa(g.Index))
	}
	return fmt.Sprintf("Getattr(Node=%s %s Line=%d Col=%d)", g.Node, param, t.Line, t.Col)
}

type Negation struct {
	Term     Expression
	Operator *tokens.Token
}

func (n *Negation) Position() *tokens.Token { return n.Operator }
func (n *Negation) String() string {
	t := n.Operator
	return fmt.Sprintf("Negation(term=%s Line=%d Col=%d)", n.Term, t.Line, t.Col)
}

type UnaryExpression struct {
	Negative bool
	Term     Expression
	Operator *tokens.Token
}

func (u *UnaryExpression) Position() *tokens.Token { return u.Operator }
func (u *UnaryExpression) String() string {
	t := u.Operator

	return fmt.Sprintf("UnaryExpression(sign=%s term=%s Line=%d Col=%d)",
		t.Val, u.Term, t.Line, t.Col)
}

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Operator *BinOperator
}

func (expr *BinaryExpression) Position() *tokens.Token { return expr.Left.Position() }
func (expr *BinaryExpression) String() string {
	t := expr.Position()

	return fmt.Sprintf("BinaryExpression(operator=%s left=%s right=%s Line=%d Col=%d)",
		expr.Operator.Token.Val, expr.Left, expr.Right, t.Line, t.Col)
}

type BinOperator struct {
	Token *tokens.Token
}

func (op BinOperator) Position() *tokens.Token { return op.Token }
func (op BinOperator) String() string          { return op.Token.String() }

type StatementBlock struct {
	Location *tokens.Token
	Name     string
	Stmt     Statement
	Trim     *Trim
	LStrip   bool
}

func (s StatementBlock) Position() *tokens.Token { return s.Location }
func (s StatementBlock) String() string {
	t := s.Position()

	return fmt.Sprintf("StatementBlock(Name=%s Impl=%s Line=%d Col=%d)",
		s.Name, s.Stmt, t.Line, t.Col)
}

type Wrapper struct {
	Location *tokens.Token
	Nodes    []Node
	EndTag   string
	Trim     *Trim
	LStrip   bool
}

func (w Wrapper) Position() *tokens.Token { return w.Location }
func (w Wrapper) String() string {
	t := w.Position()

	return fmt.Sprintf("Wrapper(Nodes=%s EndTag=%s Line=%d Col=%d)",
		w.Nodes, w.EndTag, t.Line, t.Col)
}

type Macro struct {
	Location *tokens.Token
	Name     string
	Args     []string
	Kwargs   []*Pair
	Wrapper  *Wrapper
}

func (m *Macro) Position() *tokens.Token { return m.Location }
func (m *Macro) String() string {
	t := m.Position()
	return fmt.Sprintf("Macro(Name=%s Args=%s Kwargs=%s Line=%d Col=%d)", m.Name, m.Args, m.Kwargs, t.Line, t.Col)
}
