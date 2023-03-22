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
	// If set to True the XML/HTML autoescaping feature is enabled by default.
	// For more details about autoescaping see Markup.
	// This can also be a callable that is passed the template name
	// and has to return True or False depending on autoescape should be enabled by default.
	Autoescape bool
	// Whether to be strict about undefined attribute or item in an object and return error
	// or return a nil value on missing data and ignore it entirely
	StrictUndefined bool

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
		Autoescape:          false,
		StrictUndefined:     false,
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
		Autoescape:          cfg.Autoescape,
		StrictUndefined:     cfg.StrictUndefined,
		Ext:                 ext,
	}
}

// DefaultConfig is a configuration with default values
var DefaultConfig = NewConfig()
