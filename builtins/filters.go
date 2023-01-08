package builtins

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/paradime-io/gonja/exec"
	u "github.com/paradime-io/gonja/utils"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Filters export all builtin filters
var Filters = exec.FilterSet{
	"abs":            filterAbs,
	"attr":           filterAttr,
	"batch":          filterBatch,
	"capitalize":     filterCapitalize,
	"center":         filterCenter,
	"d":              filterDefault,
	"default":        filterDefault,
	"dictsort":       filterDictSort,
	"e":              filterEscape,
	"escape":         filterEscape,
	"filesizeformat": filterFileSize,
	"first":          filterFirst,
	"float":          filterFloat,
	"forceescape":    filterForceEscape,
	"format":         filterFormat,
	"groupby":        filterGroupBy,
	"indent":         filterIndent,
	"int":            filterInteger,
	"join":           filterJoin,
	"last":           filterLast,
	"length":         filterLength,
	"list":           filterList,
	"lower":          filterLower,
	"map":            filterMap,
	"max":            filterMax,
	"min":            filterMin,
	"pprint":         filterPPrint,
	"random":         filterRandom,
	"reject":         filterReject,
	"rejectattr":     filterRejectAttr,
	"replace":        filterReplace,
	"reverse":        filterReverse,
	"round":          filterRound,
	"safe":           filterSafe,
	"select":         filterSelect,
	"selectattr":     filterSelectAttr,
	"slice":          filterSlice,
	"sort":           filterSort,
	"string":         filterString,
	"striptags":      filterStriptags,
	"sum":            filterSum,
	"title":          filterTitle,
	"tojson":         filterToJSON,
	"trim":           filterTrim,
	"truncate":       filterTruncate,
	"unique":         filterUnique,
	"upper":          filterUpper,
	"urlencode":      filterUrlencode,
	"urlize":         filterUrlize,
	"wordcount":      filterWordcount,
	"wordwrap":       filterWordwrap,
	"xmlattr":        filterXMLAttr,
}

func filterAbs(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'abs'"))
	}
	if in.IsInteger() {
		asInt := in.Integer()
		if asInt < 0 {
			return exec.AsValue(-asInt)
		}
		return in
	} else if in.IsFloat() {
		return exec.AsValue(math.Abs(in.Float()))
	}
	return exec.AsValue(math.Abs(in.Float())) // nothing to do here, just to keep track of the safe application
}

func filterAttr(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.ExpectArgs(1)
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'attr'"))
	}
	attr := p.First().String()
	value, _ := in.GetAttr(attr)
	return value
}

func filterBatch(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(1, []*exec.KwArg{{"fill_with", nil}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'batch'"))
	}
	size := p.First().Integer()
	out := []*exec.Value{}
	var row []*exec.Value
	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		if math.Mod(float64(idx), float64(size)) == 0 {
			if row != nil {
				out = append(out, exec.AsValue(row))
			}
			row = []*exec.Value{}
		}
		row = append(row, key)
		return true
	}, func() {})
	if len(row) > 0 {
		fillWith := p.KwArgs["fill_with"]
		if !fillWith.IsNil() {
			for len(row) < size {
				row = append(row, fillWith)
			}
		}
		out = append(out, exec.AsValue(row))
	}
	return exec.AsValue(out)
}

func filterCapitalize(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'capitalize'"))
	}
	if in.Len() <= 0 {
		return exec.AsValue("")
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return exec.AsValue(strings.ToUpper(string(r)) + strings.ToLower(t[size:]))
}

func filterCenter(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.ExpectArgs(1)
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'center'"))
	}
	width := p.First().Integer()
	sLen := in.Len()
	if width <= sLen {
		return in
	}

	spaces := width - sLen
	left := spaces/2 + spaces%2
	right := spaces / 2

	return exec.AsValue(fmt.Sprintf("%s%s%s", strings.Repeat(" ", left),
		in.String(), strings.Repeat(" ", right)))
}

func filterDefault(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(1, []*exec.KwArg{{"boolean", false}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'default'"))
	}
	defaultVal := p.First()
	falsy := p.KwArgs["boolean"]
	if falsy.Bool() && (in.IsError() || !in.IsTrue()) {
		return defaultVal
	} else if in.IsError() || in.IsNil() {
		return defaultVal
	}
	return in
}

func sortByKey(in *exec.Value, caseSensitive bool, reverse bool) [][2]*exec.Value {
	out := [][2]*exec.Value{}
	in.IterateOrder(func(idx, count int, key, value *exec.Value) bool {
		out = append(out, [2]*exec.Value{key, value})
		return true
	}, func() {}, reverse, true, caseSensitive)
	return out
}

func sortByValue(in *exec.Value, caseSensitive, reverse bool) [][2]*exec.Value {
	out := [][2]*exec.Value{}
	items := in.Items()
	var sorter func(i, j int) bool
	switch {
	case caseSensitive && reverse:
		sorter = func(i, j int) bool {
			return items[i].Value.String() > items[j].Value.String()
		}
	case caseSensitive && !reverse:
		sorter = func(i, j int) bool {
			return items[i].Value.String() < items[j].Value.String()
		}
	case !caseSensitive && reverse:
		sorter = func(i, j int) bool {
			return strings.ToLower(items[i].Value.String()) > strings.ToLower(items[j].Value.String())
		}
	case !caseSensitive && !reverse:
		sorter = func(i, j int) bool {
			return strings.ToLower(items[i].Value.String()) < strings.ToLower(items[j].Value.String())
		}
	}
	sort.Slice(items, sorter)
	for _, item := range items {
		out = append(out, [2]*exec.Value{item.Key, item.Value})
	}
	return out
}

func filterDictSort(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"case_sensitive", false},
		{"by", "key"},
		{"reverse", false},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'dictsort'"))
	}

	caseSensitive := p.KwArgs["case_sensitive"].Bool()
	by := p.KwArgs["by"].String()
	reverse := p.KwArgs["reverse"].Bool()

	switch by {
	case "key":
		return exec.AsValue(sortByKey(in, caseSensitive, reverse))
	case "value":
		return exec.AsValue(sortByValue(in, caseSensitive, reverse))
	default:
		return exec.AsValue(errors.New(`by should be either 'key' or 'value`))
	}
}

func filterEscape(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'escape'"))
	}
	if in.Safe {
		return in
	}
	return exec.AsSafeValue(in.Escaped())
}

var (
	bytesPrefixes  = []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	binaryPrefixes = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
)

func filterFileSize(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{{"binary", false}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'filesizeformat'"))
	}
	bytes := in.Float()
	binary := p.KwArgs["binary"].Bool()
	var base float64
	var prefixes []string
	if binary {
		base = 1024.0
		prefixes = binaryPrefixes
	} else {
		base = 1000.0
		prefixes = bytesPrefixes
	}
	if bytes == 1.0 {
		return exec.AsValue("1 Byte")
	} else if bytes < base {
		return exec.AsValue(fmt.Sprintf("%.0f Bytes", bytes))
	} else {
		var i int
		var unit float64
		var prefix string
		for i, prefix = range prefixes {
			unit = math.Pow(base, float64(i+2))
			if bytes < unit {
				return exec.AsValue(fmt.Sprintf("%.1f %s", (base * bytes / unit), prefix))
			}
		}
		return exec.AsValue(fmt.Sprintf("%.1f %s", (base * bytes / unit), prefix))
	}
}

func filterFirst(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'first'"))
	}
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(0)
	}
	return exec.AsValue("")
}

func filterFloat(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'float'"))
	}
	return exec.AsValue(in.Float())
}

func filterForceEscape(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'forceescape'"))
	}
	return exec.AsSafeValue(in.Escaped())
}

func filterFormat(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	args := []interface{}{}
	for _, arg := range params.Args {
		args = append(args, arg.Interface())
	}
	return exec.AsValue(fmt.Sprintf(in.String(), args...))
}

func filterGroupBy(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.ExpectArgs(1)
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'groupby"))
	}
	field := p.First().String()
	groups := map[interface{}][]*exec.Value{}
	groupers := []interface{}{}

	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		attr, found := key.Get(field)
		if !found {
			return true
		}
		lst, exists := groups[attr.Interface()]
		if !exists {
			lst = []*exec.Value{}
			groupers = append(groupers, attr.Interface())
		}
		lst = append(lst, key)
		groups[attr.Interface()] = lst
		return true
	}, func() {})

	out := []map[string]*exec.Value{}
	for _, grouper := range groupers {
		out = append(out, map[string]*exec.Value{
			"grouper": exec.AsValue(grouper),
			"list":    exec.AsValue(groups[grouper]),
		})
	}
	return exec.AsValue(out)
}

func filterIndent(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"width", 4},
		{"first", false},
		{"blank", false},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'indent'"))
	}
	width := p.KwArgs["width"].Integer()
	first := p.KwArgs["first"].Bool()
	blank := p.KwArgs["blank"].Bool()
	indent := strings.Repeat(" ", width)
	lines := strings.Split(in.String(), "\n")
	// start := 1
	// if first {start = 0}
	var out strings.Builder
	for idx, line := range lines {
		if line == "" && !blank {
			out.WriteByte('\n')
			continue
		}
		if idx > 0 || first {
			out.WriteString(indent)
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return exec.AsValue(out.String())
}

func filterInteger(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'int'"))
	}
	return exec.AsValue(in.Integer())
}

func filterJoin(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"d", ""},
		{"attribute", nil},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'join'"))
	}
	if !in.CanSlice() {
		return in
	}
	sep := p.KwArgs["d"].String()
	sl := make([]string, 0, in.Len())
	for i := 0; i < in.Len(); i++ {
		sl = append(sl, in.Index(i).String())
	}
	return exec.AsValue(strings.Join(sl, sep))
}

func filterLast(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'last'"))
	}
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(in.Len() - 1)
	}
	return exec.AsValue("")
}

func filterLength(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'length'"))
	}
	return exec.AsValue(in.Len())
}

func filterList(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'list'"))
	}
	if in.IsString() {
		out := []string{}
		for _, r := range in.String() {
			out = append(out, string(r))
		}
		return exec.AsValue(out)
	}
	out := []*exec.Value{}
	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		out = append(out, key)
		return true
	}, func() {})
	return exec.AsValue(out)
}

func filterLower(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'lower'"))
	}
	return exec.AsValue(strings.ToLower(in.String()))
}

func filterMap(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"filter", ""},
		{"attribute", nil},
		{"default", nil},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'map'"))
	}
	filter := p.KwArgs["filter"].String()
	attribute := p.KwArgs["attribute"].String()
	defaultVal := p.KwArgs["default"]
	out := []*exec.Value{}
	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		val := key
		if len(attribute) > 0 {
			attr, found := val.Get(attribute)
			if found {
				val = attr
			} else if defaultVal != nil {
				val = defaultVal
			} else {
				return true
			}
		}
		if len(filter) > 0 {
			val = e.ExecuteFilterByName(filter, val, exec.NewVarArgs())
		}
		out = append(out, val)
		return true
	}, func() {})
	return exec.AsValue(out)
}

func filterMax(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"case_sensitive", false},
		{"attribute", nil},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'max'"))
	}
	caseSensitive := p.KwArgs["case_sensitive"].Bool()
	attribute := p.KwArgs["attribute"].String()

	var max *exec.Value
	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		val := key
		if len(attribute) > 0 {
			attr, found := val.Get(attribute)
			if found {
				val = attr
			} else {
				val = nil
			}
		}
		if max == nil {
			max = val
			return true
		}
		if val == nil || max == nil {
			return true
		}
		switch {
		case max.IsFloat() || max.IsInteger() && val.IsFloat() || val.IsInteger():
			if val.Float() > max.Float() {
				max = val
			}
		case max.IsString() && val.IsString():
			if !caseSensitive && strings.ToLower(val.String()) > strings.ToLower(max.String()) {
				max = val
			} else if caseSensitive && val.String() > max.String() {
				max = val
			}
		default:
			max = exec.AsValue(errors.Errorf(`%s and %s are not comparable`, max.Val.Type(), val.Val.Type()))
		}
		return true
	}, func() {})

	if max == nil {
		return exec.AsValue("")
	}
	return max
}

func filterMin(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"case_sensitive", false},
		{"attribute", nil},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'min'"))
	}
	caseSensitive := p.KwArgs["case_sensitive"].Bool()
	attribute := p.KwArgs["attribute"].String()

	var min *exec.Value
	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		val := key
		if len(attribute) > 0 {
			attr, found := val.Get(attribute)
			if found {
				val = attr
			} else {
				val = nil
			}
		}
		if min == nil {
			min = val
			return true
		}
		if val == nil || min == nil {
			return true
		}
		switch {
		case min.IsFloat() || min.IsInteger() && val.IsFloat() || val.IsInteger():
			if val.Float() < min.Float() {
				min = val
			}
		case min.IsString() && val.IsString():
			if !caseSensitive && strings.ToLower(val.String()) < strings.ToLower(min.String()) {
				min = val
			} else if caseSensitive && val.String() < min.String() {
				min = val
			}
		default:
			min = exec.AsValue(errors.Errorf(`%s and %s are not comparable`, min.Val.Type(), val.Val.Type()))
		}
		return true
	}, func() {})

	if min == nil {
		return exec.AsValue("")
	}
	return min
}

func filterPPrint(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{{"verbose", false}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'pprint'"))
	}
	b, err := json.MarshalIndent(in.Interface(), "", "  ")
	if err != nil {
		return exec.AsValue(errors.Wrapf(err, `Unable to pretty print '%s'`, in.String()))
	}
	return exec.AsSafeValue(string(b))
}

func filterRandom(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'random'"))
	}
	if !in.CanSlice() || in.Len() <= 0 {
		return in
	}
	i := rand.Intn(in.Len())
	return in.Index(i)
}

func filterReject(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	var test func(*exec.Value) bool
	if len(params.Args) == 0 {
		// Reject truthy value
		test = func(in *exec.Value) bool {
			return in.IsTrue()
		}
	} else {
		name := params.First().String()
		testParams := &exec.VarArgs{
			Args:   params.Args[1:],
			KwArgs: params.KwArgs,
		}
		test = func(in *exec.Value) bool {
			out := e.ExecuteTestByName(name, in, testParams)
			return out.IsTrue()
		}
	}

	out := []*exec.Value{}

	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		if !test(key) {
			out = append(out, key)
		}
		return true
	}, func() {})

	return exec.AsValue(out)
}

func filterRejectAttr(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	var test func(*exec.Value) *exec.Value
	if len(params.Args) < 1 {
		return exec.AsValue(errors.New("Wrong signature for 'rejectattr', expect at least an attribute name as argument"))
	}
	attribute := params.First().String()
	if len(params.Args) == 1 {
		// Reject truthy value
		test = func(in *exec.Value) *exec.Value {
			attr, found := in.Get(attribute)
			if !found {
				return exec.AsValue(errors.Errorf(`%s has no attribute '%s'`, in.String(), attribute))
			}
			return attr
		}
	} else {
		name := params.Args[1].String()
		testParams := &exec.VarArgs{
			Args:   params.Args[2:],
			KwArgs: params.KwArgs,
		}
		test = func(in *exec.Value) *exec.Value {
			attr, found := in.Get(attribute)
			if !found {
				return exec.AsValue(errors.Errorf(`%s has no attribute '%s'`, in.String(), attribute))
			}
			out := e.ExecuteTestByName(name, attr, testParams)
			return out
		}
	}

	out := []*exec.Value{}
	var err *exec.Value

	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		result := test(key)
		if result.IsError() {
			err = result
			return false
		}
		if !result.IsTrue() {
			out = append(out, key)
		}
		return true
	}, func() {})

	if err != nil {
		return err
	}
	return exec.AsValue(out)
}

func filterReplace(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(2, []*exec.KwArg{{"count", nil}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'replace'"))
	}
	old := p.Args[0].String()
	new := p.Args[1].String()
	count := p.KwArgs["count"]
	if count.IsNil() {
		return exec.AsValue(strings.ReplaceAll(in.String(), old, new))
	}
	return exec.AsValue(strings.Replace(in.String(), old, new, count.Integer()))
}

func filterReverse(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'safe'"))
	}
	if in.IsString() {
		var out strings.Builder
		in.IterateOrder(func(idx, count int, key, value *exec.Value) bool {
			out.WriteString(key.String())
			return true
		}, func() {}, true, false, false)
		return exec.AsValue(out.String())
	}
	out := []*exec.Value{}
	in.IterateOrder(func(idx, count int, key, value *exec.Value) bool {
		out = append(out, key)
		return true
	}, func() {}, true, true, false)
	return exec.AsValue(out)
}

func filterRound(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{{"precision", 0}, {"method", "common"}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'round'"))
	}
	method := p.KwArgs["method"].String()
	var op func(float64) float64
	switch method {
	case "common":
		op = math.Round
	case "floor":
		op = math.Floor
	case "ceil":
		op = math.Ceil
	default:
		return exec.AsValue(errors.Errorf(`Unknown method '%s', mush be one of 'common, 'floor', 'ceil`, method))
	}
	value := in.Float()
	factor := float64(10 * p.KwArgs["precision"].Integer())
	if factor > 0 {
		value = value * factor
	}
	value = op(value)
	if factor > 0 {
		value = value / factor
	}
	return exec.AsValue(value)
}

func filterSafe(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'safe'"))
	}
	in.Safe = true
	return in // nothing to do here, just to keep track of the safe application
}

func filterSelect(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	var test func(*exec.Value) bool
	if len(params.Args) == 0 {
		// Reject truthy value
		test = func(in *exec.Value) bool {
			return in.IsTrue()
		}
	} else {
		name := params.First().String()
		testParams := &exec.VarArgs{
			Args:   params.Args[1:],
			KwArgs: params.KwArgs,
		}
		test = func(in *exec.Value) bool {
			out := e.ExecuteTestByName(name, in, testParams)
			return out.IsTrue()
		}
	}

	out := []*exec.Value{}

	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		if test(key) {
			out = append(out, key)
		}
		return true
	}, func() {})

	return exec.AsValue(out)
}

func filterSelectAttr(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	var test func(*exec.Value) *exec.Value
	if len(params.Args) < 1 {
		return exec.AsValue(errors.New("Wrong signature for 'selectattr', expect at least an attribute name as argument"))
	}
	attribute := params.First().String()
	if len(params.Args) == 1 {
		// Reject truthy value
		test = func(in *exec.Value) *exec.Value {
			attr, found := in.Get(attribute)
			if !found {
				return exec.AsValue(errors.Errorf(`%s has no attribute '%s'`, in.String(), attribute))
			}
			return attr
		}
	} else {
		name := params.Args[1].String()
		testParams := &exec.VarArgs{
			Args:   params.Args[2:],
			KwArgs: params.KwArgs,
		}
		test = func(in *exec.Value) *exec.Value {
			attr, found := in.Get(attribute)
			if !found {
				return exec.AsValue(errors.Errorf(`%s has no attribute '%s'`, in.String(), attribute))
			}
			out := e.ExecuteTestByName(name, attr, testParams)
			return out
		}
	}

	out := []*exec.Value{}
	var err *exec.Value

	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		result := test(key)
		if result.IsError() {
			err = result
			return false
		}
		if result.IsTrue() {
			out = append(out, key)
		}
		return true
	}, func() {})

	if err != nil {
		return err
	}
	return exec.AsValue(out)
}

func filterSlice(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	comp := strings.Split(params.Args[0].String(), ":")
	if len(comp) != 2 {
		return exec.AsValue(errors.New("Slice string must have the format 'from:to' [from/to can be omitted, but the ':' is required]"))
	}

	if !in.CanSlice() {
		return in
	}

	from := exec.AsValue(comp[0]).Integer()
	to := in.Len()

	if from > to {
		from = to
	}

	vto := exec.AsValue(comp[1]).Integer()
	if vto >= from && vto <= in.Len() {
		to = vto
	}

	return in.Slice(from, to)
}

func filterSort(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{{"reverse", false}, {"case_sensitive", false}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'sort'"))
	}
	reverse := p.KwArgs["reverse"].Bool()
	caseSensitive := p.KwArgs["case_sensitive"].Bool()
	out := []*exec.Value{}
	in.IterateOrder(func(idx, count int, key, value *exec.Value) bool {
		out = append(out, key)
		return true
	}, func() {}, reverse, true, caseSensitive)
	return exec.AsValue(out)
}

func filterString(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'string'"))
	}
	return exec.AsValue(in.String())
}

var reStriptags = regexp.MustCompile("<[^>]*?>")

func filterStriptags(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'striptags'"))
	}
	s := in.String()

	// Strip all tags
	s = reStriptags.ReplaceAllString(s, "")

	return exec.AsValue(strings.TrimSpace(s))
}

func filterSum(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{{"attribute", nil}, {"start", 0}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'sum'"))
	}

	attribute := p.KwArgs["attribute"]
	sum := p.KwArgs["start"].Float()
	var err error

	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		if attribute.IsString() {
			val := key
			found := true
			for _, attr := range strings.Split(attribute.String(), ".") {
				val, found = val.Get(attr)
				if !found {
					err = errors.Errorf("'%s' has no attribute '%s'", key.String(), attribute.String())
					return false
				}
			}
			if found && val.IsNumber() {
				sum += val.Float()
			}
		} else if attribute.IsInteger() {
			value, found := key.GetItem(attribute.Integer())
			if found {
				sum += value.Float()
			}
		} else {
			sum += key.Float()
		}
		return true
	}, func() {})

	if err != nil {
		return exec.AsValue(err)
	} else if sum == math.Trunc(sum) {
		return exec.AsValue(int64(sum))
	}
	return exec.AsValue(sum)
}

func filterTitle(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'title'"))
	}
	if !in.IsString() {
		return exec.AsValue("")
	}
	caser := cases.Title(language.Und)
	return exec.AsValue(caser.String(in.String()))
}

func filterTrim(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'trim'"))
	}
	return exec.AsValue(strings.TrimSpace(in.String()))
}

func filterToJSON(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{{"indent", nil}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'tojson'"))
	}

	indent := p.KwArgs["indent"]
	var out string
	if indent.IsNil() {
		b, err := json.Marshal(in.Interface())
		if err != nil {
			return exec.AsValue(errors.Wrap(err, "Unable to marshall to json"))
		}
		out = string(b)
	} else if indent.IsInteger() {
		b, err := json.MarshalIndent(in.Interface(), "", strings.Repeat(" ", indent.Integer()))
		if err != nil {
			return exec.AsValue(errors.Wrap(err, "Unable to marshall to json"))
		}
		out = string(b)
	} else {
		return exec.AsValue(errors.Errorf("Expected an integer for 'indent', got %s", indent.String()))
	}
	return exec.AsSafeValue(out)
}

func filterTruncate(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"length", 255},
		{"killwords", false},
		{"end", "..."},
		{"leeway", 0},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'truncate'"))
	}

	source := in.String()
	length := p.KwArgs["length"].Integer()
	leeway := p.KwArgs["leeway"].Integer()
	killwords := p.KwArgs["killwords"].Bool()
	end := p.KwArgs["end"].String()
	rEnd := []rune(end)
	fullLength := length + leeway
	runes := []rune(source)

	if length < len(rEnd) {
		return exec.AsValue(errors.Errorf(`expected length >= %d, got %d`, len(rEnd), length))
	}

	if len(runes) <= fullLength {
		return exec.AsValue(source)
	}

	atLength := string(runes[:length-len(rEnd)])
	if !killwords {
		atLength = strings.TrimRightFunc(atLength, func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		atLength = strings.TrimRight(atLength, " \n\t")
	}
	return exec.AsValue(fmt.Sprintf("%s%s", atLength, end))
}

func filterUnique(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{{"case_sensitive", false}, {"attribute", nil}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'unique'"))
	}

	caseSensitive := p.KwArgs["case_sensitive"].Bool()
	attribute := p.KwArgs["attribute"]

	out := exec.ValuesList{}
	tracker := map[interface{}]bool{}
	var err error

	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		val := key
		if attribute.IsString() {
			attr := attribute.String()
			nested, found := key.Get(attr)
			if !found {
				err = errors.Errorf(`%s has no attribute %s`, key.String(), attr)
				return false
			}
			val = nested
		}
		tracked := val.Interface()
		if !caseSensitive && val.IsString() {
			tracked = strings.ToLower(val.String())
		}
		if _, contains := tracker[tracked]; !contains {
			tracker[tracked] = true
			out = append(out, key)
		}
		return true
	}, func() {})

	if err != nil {
		return exec.AsValue(err)
	}
	return exec.AsValue(out)
}

func filterUpper(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'upper'"))
	}
	return exec.AsValue(strings.ToUpper(in.String()))
}

func filterUrlencode(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'urlencode'"))
	}
	return exec.AsValue(url.QueryEscape(in.String()))
}

// TODO: This regexp could do some work
var filterUrlizeURLRegexp = regexp.MustCompile(`((((http|https)://)|www\.|((^|[ ])[0-9A-Za-z_\-]+(\.com|\.net|\.org|\.info|\.biz|\.de))))(?U:.*)([ ]+|$)`)
var filterUrlizeEmailRegexp = regexp.MustCompile(`(\w+@\w+\.\w{2,4})`)

func filterUrlizeHelper(input string, trunc int, rel string, target string) (string, error) {
	var sOutErr error
	sOut := filterUrlizeURLRegexp.ReplaceAllStringFunc(input, func(raw_url string) string {
		var prefix string
		var suffix string
		if strings.HasPrefix(raw_url, " ") {
			prefix = " "
		}
		if strings.HasSuffix(raw_url, " ") {
			suffix = " "
		}

		raw_url = strings.TrimSpace(raw_url)

		url := u.IRIEncode(raw_url)

		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("http://%s", url)
		}

		title := raw_url

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		title = u.Escape(title)

		attrs := ""
		if len(target) > 0 {
			attrs = fmt.Sprintf(` target="%s"`, target)
		}

		rels := []string{}
		cleanedRel := strings.Trim(strings.Replace(rel, "noopener", "", -1), " ")
		if len(cleanedRel) > 0 {
			rels = append(rels, cleanedRel)
		}
		rels = append(rels, "noopener")
		rel = strings.Join(rels, " ")

		return fmt.Sprintf(`%s<a href="%s" rel="%s"%s>%s</a>%s`, prefix, url, rel, attrs, title, suffix)
	})
	if sOutErr != nil {
		return "", sOutErr
	}

	sOut = filterUrlizeEmailRegexp.ReplaceAllStringFunc(sOut, func(mail string) string {
		title := mail

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		return fmt.Sprintf(`<a href="mailto:%s">%s</a>`, mail, title)
	})
	return sOut, nil
}

func filterUrlize(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.Expect(0, []*exec.KwArg{
		{"trim_url_limit", nil},
		{"nofollow", false},
		{"target", nil},
		{"rel", nil},
	})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'urlize'"))
	}
	truncate := -1
	if param := p.KwArgs["trim_url_limit"]; param.IsInteger() {
		truncate = param.Integer()
	}
	rel := p.KwArgs["rel"]
	target := p.KwArgs["target"]

	s, err := filterUrlizeHelper(in.String(), truncate, rel.String(), target.String())
	if err != nil {
		return exec.AsValue(err)
	}

	return exec.AsValue(s)
}

func filterWordcount(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if p := params.ExpectNothing(); p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'wordcount'"))
	}
	return exec.AsValue(len(strings.Fields(in.String())))
}

func filterWordwrap(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	words := strings.Fields(in.String())
	wordsLen := len(words)
	wrapAt := params.Args[0].Integer()
	if wrapAt <= 0 {
		return in
	}

	linecount := wordsLen/wrapAt + wordsLen%wrapAt
	lines := make([]string, 0, linecount)
	for i := 0; i < linecount; i++ {
		lines = append(lines, strings.Join(words[wrapAt*i:u.Min(wrapAt*(i+1), wordsLen)], " "))
	}
	return exec.AsValue(strings.Join(lines, "\n"))
}

func filterXMLAttr(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	p := params.ExpectKwArgs([]*exec.KwArg{{"autospace", true}})
	if p.IsError() {
		return exec.AsValue(errors.Wrap(p, "Wrong signature for 'xmlattr'"))
	}
	autospace := p.KwArgs["autospace"].Bool()
	kvs := []string{}
	in.Iterate(func(idx, count int, key, value *exec.Value) bool {
		if !value.IsTrue() {
			return true
		}
		kv := fmt.Sprintf(`%s="%s"`, key.Escaped(), value.Escaped())
		kvs = append(kvs, kv)
		return true
	}, func() {})
	out := strings.Join(kvs, " ")
	if autospace {
		out = " " + out
	}
	return exec.AsValue(out)
}
