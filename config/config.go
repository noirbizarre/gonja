package config

type Inheritable interface {
	Inherit() Inheritable
}

// Config holds plexer and parser parameters
type Config struct {
	Debug bool
	// The string marking the beginning of a block. Defaults to '{%'
	BlockStartString string
	// The string marking the end of a block. Defaults to '%}'.
	BlockEndString string
	// The string marking the beginning of a print statement. Defaults to '{{'.
	VariableStartString string
	// The string marking the end of a print statement. Defaults to '}}'.
	VariableEndString string
	// The string marking the beginning of a comment. Defaults to '{#'.
	CommentStartString string
	// The string marking the end of a comment. Defaults to '#}'.
	CommentEndString string
	// If given and a string, this will be used as prefix for line based statements.
	// See also Line Statements.
	LineStatementPrefix string
	// If given and a string, this will be used as prefix for line based comments.
	// See also Line Statements.
	LineCommentPrefix string
	// If this is set to True the first newline after a block is removed (block, not variable tag!).
	// Defaults to False.
	TrimBlocks bool
	// If this is set to True leading spaces and tabs are stripped from the start of a line to a block.
	// Defaults to False.
	LstripBlocks bool
	// The sequence that starts a newline.
	// Must be one of '\r', '\n' or '\r\n'.
	// The default is '\n' which is a useful default for Linux and OS X systems as well as web applications.
	NewlineSequence string
	// Preserve the trailing newline when rendering templates.
	// The default is False, which causes a single newline,
	// if present, to be stripped from the end of the template.
	KeepTrailingNewline bool
	// If set to True the XML/HTML autoescaping feature is enabled by default.
	// For more details about autoescaping see Markup.
	// This can also be a callable that is passed the template name
	// and has to return True or False depending on autoescape should be enabled by default.
	Autoescape bool

	// Allow extensions to store some config
	Ext map[string]Inheritable
}

func NewConfig() *Config {
	return &Config{
		Debug:               false,
		BlockStartString:    "{%",
		BlockEndString:      "%}",
		VariableStartString: "{{",
		VariableEndString:   "}}",
		CommentStartString:  "{#",
		CommentEndString:    "#}",
		TrimBlocks:          false,
		LstripBlocks:        false,
		NewlineSequence:     "\n",
		KeepTrailingNewline: false,
		Autoescape:          false,
		Ext:                 map[string]Inheritable{},
	}
}

func (cfg *Config) Inherit() *Config {
	ext := map[string]Inheritable{}
	for key, cfg := range cfg.Ext {
		ext[key] = cfg.Inherit()
	}
	return &Config{
		Debug:               cfg.Debug,
		BlockStartString:    cfg.BlockStartString,
		BlockEndString:      cfg.BlockEndString,
		VariableStartString: cfg.VariableStartString,
		VariableEndString:   cfg.VariableEndString,
		CommentStartString:  cfg.CommentStartString,
		CommentEndString:    cfg.CommentEndString,
		TrimBlocks:          cfg.TrimBlocks,
		LstripBlocks:        cfg.LstripBlocks,
		NewlineSequence:     cfg.NewlineSequence,
		KeepTrailingNewline: cfg.KeepTrailingNewline,
		Autoescape:          cfg.Autoescape,
		Ext:                 ext,
	}
}

// DefaultConfig is a configuration with default values
var DefaultConfig = NewConfig()
