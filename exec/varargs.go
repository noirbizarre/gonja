package exec

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// VarArgs represents pythonic variadic args/kwargs
type VarArgs struct {
	Args   []*Value
	KwArgs map[string]*Value
}

func NewVarArgs() *VarArgs {
	return &VarArgs{
		Args:   []*Value{},
		KwArgs: map[string]*Value{},
	}
}

// First returns the first argument or nil AsValue
func (va *VarArgs) First() *Value {
	if len(va.Args) > 0 {
		return va.Args[0]
	}
	return AsValue(nil)
}

// GetKwarg gets a keyword arguments with fallback on default value
func (va *VarArgs) GetKwarg(key string, fallback interface{}) *Value {
	value, ok := va.KwArgs[key]
	if ok {
		return value
	}
	return AsValue(fallback)
}

type KwArg struct {
	Name    string
	Default interface{}
}

// Expect validates VarArgs against an expected signature
func (va *VarArgs) Expect(args int, kwargs []*KwArg) *ReducedVarArgs {
	rva := &ReducedVarArgs{VarArgs: va}
	reduced := &VarArgs{
		Args:   va.Args,
		KwArgs: map[string]*Value{},
	}
	reduceIdx := -1
	unexpectedArgs := []string{}
	if len(va.Args) < args {
		// Priority on missing arguments
		if args > 1 {
			rva.error = errors.Errorf(`Expected %d arguments, got %d`, args, len(va.Args))
		} else {
			rva.error = errors.Errorf(`Expected an argument, got %d`, len(va.Args))
		}
		return rva
	} else if len(va.Args) > args {
		reduced.Args = va.Args[:args]
		for idx, arg := range va.Args[args:] {
			if len(kwargs) > idx {
				reduced.KwArgs[kwargs[idx].Name] = arg
				reduceIdx = idx + 1
			} else {
				unexpectedArgs = append(unexpectedArgs, arg.String())
			}
		}
	}

	unexpectedKwArgs := []string{}
Loop:
	for key, value := range va.KwArgs {
		for idx, kwarg := range kwargs {
			if key == kwarg.Name {
				if reduceIdx < 0 || idx >= reduceIdx {
					reduced.KwArgs[key] = value
					continue Loop
				} else {
					rva.error = errors.Errorf(`Keyword '%s' has been submitted twice`, key)
					break Loop
				}
			}
		}
		kv := strings.Join([]string{key, value.String()}, "=")
		unexpectedKwArgs = append(unexpectedKwArgs, kv)
	}
	sort.Strings(unexpectedKwArgs)

	if rva.error != nil {
		return rva
	}

	switch {
	case len(unexpectedArgs) == 0 && len(unexpectedKwArgs) == 0:
	case len(unexpectedArgs) == 1 && len(unexpectedKwArgs) == 0:
		rva.error = errors.Errorf(`Unexpected argument '%s'`, unexpectedArgs[0])
	case len(unexpectedArgs) > 1 && len(unexpectedKwArgs) == 0:
		rva.error = errors.Errorf(`Unexpected arguments '%s'`, strings.Join(unexpectedArgs, ", "))
	case len(unexpectedArgs) == 0 && len(unexpectedKwArgs) == 1:
		rva.error = errors.Errorf(`Unexpected keyword argument '%s'`, unexpectedKwArgs[0])
	case len(unexpectedArgs) == 0 && len(unexpectedKwArgs) > 0:
		rva.error = errors.Errorf(`Unexpected keyword arguments '%s'`, strings.Join(unexpectedKwArgs, ", "))
	default:
		rva.error = errors.Errorf(`Unexpected arguments '%s, %s'`,
			strings.Join(unexpectedArgs, ", "),
			strings.Join(unexpectedKwArgs, ", "),
		)
	}

	if rva.error != nil {
		return rva
	}
	// fill defaults
	for _, kwarg := range kwargs {
		_, exists := reduced.KwArgs[kwarg.Name]
		if !exists {
			reduced.KwArgs[kwarg.Name] = AsValue(kwarg.Default)
		}
	}
	rva.VarArgs = reduced
	return rva
}

// ExpectArgs ensures VarArgs receive only arguments
func (va *VarArgs) ExpectArgs(args int) *ReducedVarArgs {
	return va.Expect(args, []*KwArg{})
}

// ExpectNothing ensures VarArgs does not receive any argument
func (va *VarArgs) ExpectNothing() *ReducedVarArgs {
	return va.ExpectArgs(0)
}

// ExpectKwArgs allow to specify optionnaly expected KwArgs
func (va *VarArgs) ExpectKwArgs(kwargs []*KwArg) *ReducedVarArgs {
	return va.Expect(0, kwargs)
}

// ReducedVarArgs represents pythonic variadic args/kwargs
// but values are reduced (ie. kwargs given as args are accessible by name)
type ReducedVarArgs struct {
	*VarArgs
	error error
}

// IsError returns true if there was an error on Expect call
func (rva *ReducedVarArgs) IsError() bool {
	return rva.error != nil
}

func (rva *ReducedVarArgs) Error() string {
	if rva.IsError() {
		return rva.error.Error()
	}
	return ""
}
