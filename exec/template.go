package exec

import (
	"bytes"
	"io"
	"strings"

	"github.com/pkg/errors"

	"github.com/nikolalohinski/gonja/nodes"
	"github.com/nikolalohinski/gonja/parser"
	"github.com/nikolalohinski/gonja/tokens"
)

type TemplateLoader interface {
	GetTemplate(string) (*Template, error)
	Path(string) (string, error)
}

type Template struct {
	Name   string
	Reader io.Reader
	Source string

	Env    *EvalConfig
	Loader TemplateLoader

	Tokens *tokens.Stream
	Parser *parser.Parser

	Root   *nodes.Template
	Macros MacroSet
}

func NewTemplate(name string, source string, cfg *EvalConfig) (*Template, error) {
	// Create the template
	t := &Template{
		Env:    cfg,
		Name:   name,
		Source: source,
		Tokens: tokens.Lex(source),
	}

	// Parse it
	t.Parser = parser.NewParser(name, cfg.Config, t.Tokens)
	t.Parser.Statements = *t.Env.Statements
	t.Parser.TemplateParser = t.Env.GetTemplate
	root, err := t.Parser.Parse()
	if err != nil {
		return nil, err
	}
	t.Root = root

	return t, nil
}

func (tpl *Template) execute(ctx map[string]interface{}, out io.StringWriter) error {
	exCtx := tpl.Env.Globals.Inherit()
	exCtx.Update(ctx)

	var builder strings.Builder
	renderer := NewRenderer(exCtx, &builder, tpl.Env, tpl)

	err := renderer.Execute()
	if err != nil {
		return errors.Wrap(err, `Unable to Execute template`)
	}
	out.WriteString(renderer.String())

	return nil
}

func (tpl *Template) newBufferAndExecute(ctx map[string]interface{}) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	// Create output buffer
	// We assume that the rendered template will be 30% larger
	// buffer := bytes.NewBuffer(make([]byte, 0, int(float64(tpl.size)*1.3)))
	if err := tpl.execute(ctx, &buffer); err != nil {
		return nil, err
	}
	return &buffer, nil
}

// // Executes the template with the given context and writes to writer (io.Writer)
// // on success. Context can be nil. Nothing is written on error; instead the error
// // is being returned.
// func (tpl *Template) ExecuteWriter(ctx *Context, writer io.Writer) error {
// 	buf, err := tpl.newBufferAndExecute(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = buf.WriteTo(writer)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // // Same as ExecuteWriter. The only difference between both functions is that
// // // this function might already have written parts of the generated template in the
// // // case of an execution error because there's no intermediate buffer involved for
// // // performance reasons. This is handy if you need high performance template
// // // generation or if you want to manage your own pool of buffers.
// // func (tpl *Template) ExecuteWriterUnbuffered(ctx *Context, writer io.Writer) error {
// // 	return tpl.newTemplateWriterAndExecute(ctx, writer)
// // }

// Executes the template and returns the rendered template as a []byte
func (tpl *Template) ExecuteBytes(ctx map[string]interface{}) ([]byte, error) {
	buffer, err := tpl.newBufferAndExecute(ctx)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Executes the template and returns the rendered template as a string
func (tpl *Template) Execute(ctx map[string]interface{}) (string, error) {
	var b strings.Builder
	err := tpl.execute(ctx, &b)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
