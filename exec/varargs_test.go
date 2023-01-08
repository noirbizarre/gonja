package exec_test

import (
	// "fmt"

	"testing"

	"github.com/paradime-io/gonja/exec"
	"github.com/stretchr/testify/assert"
)

func failsafe(t *testing.T) {
	if err := recover(); err != nil {
		t.Error(err)
	}
}

func TestVarArgs(t *testing.T) {
	t.Run("first", testVAFirst)
	t.Run("GetKwarg", testVAGetKwarg)
	t.Run("Expect", testVAExpect)
}

func testVAFirst(t *testing.T) {
	t.Run("nil if empty", func(t *testing.T) {
		defer failsafe(t)
		assert := assert.New(t)

		va := exec.VarArgs{}
		first := va.First()
		assert.True(first.IsNil())
	})
	t.Run("first value", func(t *testing.T) {
		defer failsafe(t)
		assert := assert.New(t)

		va := exec.VarArgs{Args: []*exec.Value{exec.AsValue(42)}}
		first := va.First()
		assert.Equal(42, first.Integer())
	})
}

func testVAGetKwarg(t *testing.T) {
	t.Run("value if found", func(t *testing.T) {
		defer failsafe(t)
		assert := assert.New(t)

		va := exec.VarArgs{KwArgs: map[string]*exec.Value{
			"key": exec.AsValue(42),
		}}
		kwarg := va.GetKwarg("key", "not found")
		assert.Equal(42, kwarg.Integer())
	})
	t.Run("default if missing", func(t *testing.T) {
		defer failsafe(t)
		assert := assert.New(t)

		va := exec.VarArgs{}
		kwarg := va.GetKwarg("missing", "not found")
		assert.Equal("not found", kwarg.String())
	})
}

var nothingCases = []struct {
	name  string
	va    *exec.VarArgs
	error string
}{
	{"got nothing", &exec.VarArgs{}, ""}, {
		"got an argument",
		&exec.VarArgs{Args: []*exec.Value{exec.AsValue(42)}},
		`Unexpected argument '42'`,
	}, {
		"got multiples arguments",
		&exec.VarArgs{Args: []*exec.Value{exec.AsValue(42), exec.AsValue(7)}},
		`Unexpected arguments '42, 7'`,
	}, {
		"got a keyword argument",
		&exec.VarArgs{KwArgs: map[string]*exec.Value{
			"key": exec.AsValue(42),
		}},
		`Unexpected keyword argument 'key=42'`,
	}, {
		"got multiple keyword arguments",
		&exec.VarArgs{KwArgs: map[string]*exec.Value{
			"key":   exec.AsValue(42),
			"other": exec.AsValue(7),
		}},
		`Unexpected keyword arguments 'key=42, other=7'`,
	}, {
		"got one of each",
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42)},
			KwArgs: map[string]*exec.Value{
				"key": exec.AsValue(42),
			},
		},
		`Unexpected arguments '42, key=42'`,
	},
}

var argsCases = []struct {
	name  string
	va    *exec.VarArgs
	args  int
	error string
}{
	{
		"got expected",
		&exec.VarArgs{Args: []*exec.Value{exec.AsValue(42), exec.AsValue(7)}},
		2, "",
	}, {
		"got less arguments",
		&exec.VarArgs{Args: []*exec.Value{exec.AsValue(42)}},
		2, `Expected 2 arguments, got 1`,
	}, {
		"got less arguments (singular)",
		&exec.VarArgs{},
		1, `Expected an argument, got 0`,
	}, {
		"got more arguments",
		&exec.VarArgs{Args: []*exec.Value{exec.AsValue(42), exec.AsValue(7)}},
		1, `Unexpected argument '7'`,
	}, {
		"got a keyword argument",
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42)},
			KwArgs: map[string]*exec.Value{
				"key": exec.AsValue(42),
			},
		},
		1, `Unexpected keyword argument 'key=42'`,
	},
}

var kwargsCases = []struct {
	name   string
	va     *exec.VarArgs
	kwargs []*exec.KwArg
	error  string
}{
	{
		"got expected",
		&exec.VarArgs{KwArgs: map[string]*exec.Value{
			"key":   exec.AsValue(42),
			"other": exec.AsValue(7),
		}},
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		"",
	}, {
		"got unexpected arguments",
		&exec.VarArgs{Args: []*exec.Value{exec.AsValue(42), exec.AsValue(7), exec.AsValue("unexpected")}},
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		`Unexpected argument 'unexpected'`,
	}, {
		"got an unexpected keyword argument",
		&exec.VarArgs{KwArgs: map[string]*exec.Value{
			"unknown": exec.AsValue(42),
		}},
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		`Unexpected keyword argument 'unknown=42'`,
	}, {
		"got multiple keyword arguments",
		&exec.VarArgs{KwArgs: map[string]*exec.Value{
			"unknown": exec.AsValue(42),
			"seven":   exec.AsValue(7),
		}},
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		`Unexpected keyword arguments 'seven=7, unknown=42'`,
	},
}

var mixedArgsKwargsCases = []struct {
	name     string
	va       *exec.VarArgs
	args     int
	kwargs   []*exec.KwArg
	expected *exec.VarArgs
	error    string
}{
	{
		"got expected",
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42)},
			KwArgs: map[string]*exec.Value{
				"key":   exec.AsValue(42),
				"other": exec.AsValue(7),
			},
		},
		1,
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42)},
			KwArgs: map[string]*exec.Value{
				"key":   exec.AsValue(42),
				"other": exec.AsValue(7),
			},
		},
		"",
	},
	{
		"fill with default",
		&exec.VarArgs{Args: []*exec.Value{exec.AsValue(42)}},
		1,
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42)},
			KwArgs: map[string]*exec.Value{
				"key":   exec.AsValue("default key"),
				"other": exec.AsValue("default other"),
			},
		},
		"",
	},
	{
		"keyword as argument",
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42), exec.AsValue(42)},
			KwArgs: map[string]*exec.Value{
				"other": exec.AsValue(7),
			},
		},
		1,
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42)},
			KwArgs: map[string]*exec.Value{
				"key":   exec.AsValue(42),
				"other": exec.AsValue(7),
			},
		},
		"",
	},
	{
		"keyword submitted twice",
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42), exec.AsValue(5)},
			KwArgs: map[string]*exec.Value{
				"key":   exec.AsValue(42),
				"other": exec.AsValue(7),
			},
		},
		1,
		[]*exec.KwArg{
			{"key", "default key"},
			{"other", "default other"},
		},
		&exec.VarArgs{
			Args: []*exec.Value{exec.AsValue(42), exec.AsValue(5)},
			KwArgs: map[string]*exec.Value{
				"key":   exec.AsValue(42),
				"other": exec.AsValue(7),
			},
		},
		`Keyword 'key' has been submitted twice`,
	},
}

func assertError(t *testing.T, rva *exec.ReducedVarArgs, expected string) {
	assert := assert.New(t)
	if len(expected) > 0 {
		if assert.True(rva.IsError(), "Should have returned an error") {
			assert.Equal(expected, rva.Error())
		}
	} else {
		assert.Falsef(rva.IsError(), "Unexpected error: %s", rva.Error())
	}
}

func testVAExpect(t *testing.T) {
	t.Run("nothing", func(t *testing.T) {
		for _, tc := range nothingCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				defer failsafe(t)
				rva := test.va.ExpectNothing()
				assertError(t, rva, test.error)
			})
		}
	})
	t.Run("arguments", func(t *testing.T) {
		for _, tc := range argsCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				defer failsafe(t)
				rva := test.va.ExpectArgs(test.args)
				assertError(t, rva, test.error)
			})
		}
	})
	t.Run("keyword arguments", func(t *testing.T) {
		for _, tc := range kwargsCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				defer failsafe(t)
				rva := test.va.Expect(0, test.kwargs)
				assertError(t, rva, test.error)
			})
		}
	})
	t.Run("mixed arguments", func(t *testing.T) {
		for _, tc := range mixedArgsKwargsCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				defer failsafe(t)
				assert := assert.New(t)
				rva := test.va.Expect(test.args, test.kwargs)
				assertError(t, rva, test.error)
				if assert.Equal(len(test.expected.Args), len(rva.Args)) {
					for idx, expected := range test.expected.Args {
						arg := rva.Args[idx]
						assert.Equalf(expected.Interface(), arg.Interface(),
							`Argument %d mismatch: expected '%s' got '%s'`,
							idx, expected.String(), arg.String(),
						)
					}
				}
				if assert.Equal(len(test.expected.KwArgs), len(rva.KwArgs)) {
					for key, expected := range test.expected.KwArgs {
						if assert.Contains(rva.KwArgs, key) {
							value := rva.KwArgs[key]
							assert.Equalf(expected.Interface(), value.Interface(),
								`Keyword argument %s mismatch: expected '%s' got '%s'`,
								key, expected.String(), value.String(),
							)

						}
					}
				}
			})
		}
	})
}
