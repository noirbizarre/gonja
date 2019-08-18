package django

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/noirbizarre/gonja/exec"
	u "github.com/noirbizarre/gonja/utils"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var Filters = exec.FilterSet{
	"escapejs":           filterEscapejs,
	"add":                filterAdd,
	"addslashes":         filterAddslashes,
	"capfirst":           filterCapfirst,
	"cut":                filterCut,
	"date":               filterDate,
	"default_if_none":    filterDefaultIfNone,
	"floatformat":        filterFloatformat,
	"get_digit":          filterGetdigit,
	"iriencode":          filterIriencode,
	"length_is":          filterLengthis,
	"linebreaks":         filterLinebreaks,
	"linebreaksbr":       filterLinebreaksbr,
	"linenumbers":        filterLinenumbers,
	"ljust":              filterLjust,
	"make_list":          filterMakelist,
	"phone2numeric":      filterPhone2numeric,
	"pluralize":          filterPluralize,
	"removetags":         filterRemovetags,
	"rjust":              filterRjust,
	"split":              filterSplit,
	"stringformat":       filterStringformat,
	"time":               filterDate, // time uses filterDate (same golang-format,
	"truncatechars":      filterTruncatechars,
	"truncatechars_html": filterTruncatecharsHTML,
	"truncatewords":      filterTruncatewords,
	"truncatewords_html": filterTruncatewordsHTML,
	"yesno":              filterYesno,
}

func filterTruncatecharsHelper(s string, newLen int) string {
	runes := []rune(s)
	if newLen < len(runes) {
		if newLen >= 3 {
			return fmt.Sprintf("%s...", string(runes[:newLen-3]))
		}
		// Not enough space for the ellipsis
		return string(runes[:newLen])
	}
	return string(runes)
}

func filterTruncateHTMLHelper(value string, newOutput *bytes.Buffer, cond func() bool, fn func(c rune, s int, idx int) int, finalize func()) {
	vLen := len(value)
	var tagStack []string
	idx := 0

	for idx < vLen && !cond() {
		c, s := utf8.DecodeRuneInString(value[idx:])
		if c == utf8.RuneError {
			idx += s
			continue
		}

		if c == '<' {
			newOutput.WriteRune(c)
			idx += s // consume "<"

			if idx+1 < vLen {
				if value[idx] == '/' {
					// Close tag

					newOutput.WriteString("/")

					tag := ""
					idx++ // consume "/"

					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}
						tag += string(c2)
						idx += size2
					}

					if len(tagStack) > 0 {
						// Ideally, the close tag is TOP of tag stack
						// In malformed HTML, it must not be, so iterate through the stack and remove the tag
						for i := len(tagStack) - 1; i >= 0; i-- {
							if tagStack[i] == tag {
								// Found the tag
								tagStack[i] = tagStack[len(tagStack)-1]
								tagStack = tagStack[:len(tagStack)-1]
								break
							}
						}
					}

					newOutput.WriteString(tag)
					newOutput.WriteString(">")
				} else {
					// Open tag

					tag := ""

					params := false
					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						newOutput.WriteRune(c2)

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}

						if !params {
							if c2 == ' ' {
								params = true
							} else {
								tag += string(c2)
							}
						}

						idx += size2
					}

					// Add tag to stack
					tagStack = append(tagStack, tag)
				}
			}
		} else {
			idx = fn(c, s, idx)
		}
	}

	finalize()

	for i := len(tagStack) - 1; i >= 0; i-- {
		tag := tagStack[i]
		// Close everything from the regular tag stack
		newOutput.WriteString(fmt.Sprintf("</%s>", tag))
	}
}

func filterTruncatechars(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	s := in.String()
	newLen := params.Args[0].Integer()
	return exec.AsValue(filterTruncatecharsHelper(s, newLen))
}

func filterTruncatecharsHTML(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	value := in.String()
	newLen := u.Max(params.Args[0].Integer()-3, 0)

	newOutput := bytes.NewBuffer(nil)

	textcounter := 0

	filterTruncateHTMLHelper(value, newOutput, func() bool {
		return textcounter >= newLen
	}, func(c rune, s int, idx int) int {
		textcounter++
		newOutput.WriteRune(c)

		return idx + s
	}, func() {
		if textcounter >= newLen && textcounter < len(value) {
			newOutput.WriteString("...")
		}
	})

	return exec.AsSafeValue(newOutput.String())
}

func filterTruncatewords(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	words := strings.Fields(in.String())
	n := params.Args[0].Integer()
	if n <= 0 {
		return exec.AsValue("")
	}
	nlen := u.Min(len(words), n)
	out := make([]string, 0, nlen)
	for i := 0; i < nlen; i++ {
		out = append(out, words[i])
	}

	if n < len(words) {
		out = append(out, "...")
	}

	return exec.AsValue(strings.Join(out, " "))
}

func filterTruncatewordsHTML(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	value := in.String()
	newLen := u.Max(params.Args[0].Integer(), 0)

	newOutput := bytes.NewBuffer(nil)

	wordcounter := 0

	filterTruncateHTMLHelper(value, newOutput, func() bool {
		return wordcounter >= newLen
	}, func(_ rune, _ int, idx int) int {
		// Get next word
		wordFound := false

		for idx < len(value) {
			c2, size2 := utf8.DecodeRuneInString(value[idx:])
			if c2 == utf8.RuneError {
				idx += size2
				continue
			}

			if c2 == '<' {
				// HTML tag start, don't consume it
				return idx
			}

			newOutput.WriteRune(c2)
			idx += size2

			if c2 == ' ' || c2 == '.' || c2 == ',' || c2 == ';' {
				// Word ends here, stop capturing it now
				break
			} else {
				wordFound = true
			}
		}

		if wordFound {
			wordcounter++
		}

		return idx
	}, func() {
		if wordcounter >= newLen {
			newOutput.WriteString("...")
		}
	})

	return exec.AsSafeValue(newOutput.String())
}

func filterEscapejs(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	sin := in.String()

	var b bytes.Buffer

	idx := 0
	for idx < len(sin) {
		c, size := utf8.DecodeRuneInString(sin[idx:])
		if c == utf8.RuneError {
			idx += size
			continue
		}

		if c == '\\' {
			// Escape seq?
			if idx+1 < len(sin) {
				switch sin[idx+1] {
				case 'r':
					b.WriteString(fmt.Sprintf(`\u%04X`, '\r'))
					idx += 2
					continue
				case 'n':
					b.WriteString(fmt.Sprintf(`\u%04X`, '\n'))
					idx += 2
					continue
					/*case '\'':
						b.WriteString(fmt.Sprintf(`\u%04X`, '\''))
						idx += 2
						continue
					case '"':
						b.WriteString(fmt.Sprintf(`\u%04X`, '"'))
						idx += 2
						continue*/
				}
			}
		}

		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == ' ' || c == '/' {
			b.WriteRune(c)
		} else {
			b.WriteString(fmt.Sprintf(`\u%04X`, c))
		}

		idx += size
	}

	return exec.AsValue(b.String())
}

func filterAdd(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	param := params.Args[0]
	if in.IsNumber() && param.IsNumber() {
		if in.IsFloat() || param.IsFloat() {
			return exec.AsValue(in.Float() + param.Float())
		}
		return exec.AsValue(in.Integer() + param.Integer())
	}
	// If in/param is not a number, we're relying on the
	// Value's String() conversion and just add them both together
	return exec.AsValue(in.String() + param.String())
}

func filterAddslashes(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	output := strings.Replace(in.String(), "\\", "\\\\", -1)
	output = strings.Replace(output, "\"", "\\\"", -1)
	output = strings.Replace(output, "'", "\\'", -1)
	return exec.AsValue(output)
}

func filterCut(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	return exec.AsValue(strings.Replace(in.String(), params.Args[0].String(), "", -1))
}

func filterLengthis(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	return exec.AsValue(in.Len() == params.Args[0].Integer())
}

func filterDefaultIfNone(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.IsError() || in.IsNil() {
		return params.Args[0]
	}
	return in
}

func filterFloatformat(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	val := in.Float()
	param := params.First()

	decimals := -1
	if !param.IsNil() {
		// Any argument provided?
		decimals = param.Integer()
	}

	// if the argument is not a number (e. g. empty), the default
	// behaviour is trim the result
	trim := !param.IsNumber()

	if decimals <= 0 {
		// argument is negative or zero, so we
		// want the output being trimmed
		decimals = -decimals
		trim = true
	}

	if trim {
		// Remove zeroes
		if float64(int(val)) == val {
			return exec.AsValue(in.Integer())
		}
	}

	return exec.AsValue(strconv.FormatFloat(val, 'f', decimals, 64))
}

func filterGetdigit(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if len(params.Args) > 1 {
		return exec.AsValue(errors.New("'getdigit' filter expect one and only one argument"))
		// return nil, &Error{
		// 	Sender:    "filter:getdigit",
		// 	OrigError: errors.New("'getdigit' filter expect one and only one argument"),
		// }
	}
	param := params.First()
	i := param.Integer()
	l := len(in.String()) // do NOT use in.Len() here!
	if i <= 0 || i > l {
		return in
	}
	return exec.AsValue(in.String()[l-i] - 48)
}

func filterIriencode(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	return exec.AsValue(u.IRIEncode(in.String()))
}

func filterMakelist(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	s := in.String()
	result := make([]string, 0, len(s))
	for _, c := range s {
		result = append(result, string(c))
	}
	return exec.AsValue(result)
}

func filterCapfirst(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.Len() <= 0 {
		return exec.AsValue("")
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return exec.AsValue(strings.ToUpper(string(r)) + t[size:])
}

func filterDate(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	t, isTime := in.Interface().(time.Time)
	if !isTime {
		return exec.AsValue(errors.New("filter input argument must be of type 'time.Time'"))
		// return nil, &Error{
		// 	Sender:    "filter:date",
		// 	OrigError: errors.New("filter input argument must be of type 'time.Time'"),
		// }
	}
	return exec.AsValue(t.Format(params.Args[0].String()))
}

func filterLinebreaks(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if in.Len() == 0 {
		return in
	}

	var b bytes.Buffer

	// Newline = <br />
	// Double newline = <p>...</p>
	lines := strings.Split(in.String(), "\n")
	lenlines := len(lines)

	opened := false

	for idx, line := range lines {

		if !opened {
			b.WriteString("<p>")
			opened = true
		}

		b.WriteString(line)

		if idx < lenlines-1 && strings.TrimSpace(lines[idx]) != "" {
			// We've not reached the end
			if strings.TrimSpace(lines[idx+1]) == "" {
				// Next line is empty
				if opened {
					b.WriteString("</p>")
					opened = false
				}
			} else {
				b.WriteString("<br />")
			}
		}
	}

	if opened {
		b.WriteString("</p>")
	}

	return exec.AsValue(b.String())
}

func filterSplit(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	chunks := strings.Split(in.String(), params.Args[0].String())

	return exec.AsValue(chunks)
}

func filterLinebreaksbr(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	return exec.AsValue(strings.Replace(in.String(), "\n", "<br />", -1))
}

func filterLinenumbers(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	lines := strings.Split(in.String(), "\n")
	output := make([]string, 0, len(lines))
	for idx, line := range lines {
		output = append(output, fmt.Sprintf("%d. %s", idx+1, line))
	}
	return exec.AsValue(strings.Join(output, "\n"))
}

func filterLjust(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	times := params.Args[0].Integer() - in.Len()
	if times < 0 {
		times = 0
	}
	return exec.AsValue(fmt.Sprintf("%s%s", in.String(), strings.Repeat(" ", times)))
}

func filterStringformat(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	return exec.AsValue(fmt.Sprintf(params.Args[0].String(), in.Interface()))
}

// https://en.wikipedia.org/wiki/Phoneword
var filterPhone2numericMap = map[string]string{
	"a": "2", "b": "2", "c": "2", "d": "3", "e": "3", "f": "3", "g": "4", "h": "4", "i": "4", "j": "5", "k": "5",
	"l": "5", "m": "6", "n": "6", "o": "6", "p": "7", "q": "7", "r": "7", "s": "7", "t": "8", "u": "8", "v": "8",
	"w": "9", "x": "9", "y": "9", "z": "9",
}

func filterPhone2numeric(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	sin := in.String()
	for k, v := range filterPhone2numericMap {
		sin = strings.Replace(sin, k, v, -1)
		sin = strings.Replace(sin, strings.ToUpper(k), v, -1)
	}
	return exec.AsValue(sin)
}

func filterPluralize(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	param := params.First()
	if in.IsNumber() {
		// Works only on numbers
		if param.Len() > 0 {
			endings := strings.Split(param.String(), ",")
			if len(endings) > 2 {
				return exec.AsValue(errors.New("you cannot pass more than 2 arguments to filter 'pluralize'"))
				// return nil, &Error{
				// 	Sender:    "filter:pluralize",
				// 	OrigError: errors.New("you cannot pass more than 2 arguments to filter 'pluralize'"),
				// }
			}
			if len(endings) == 1 {
				// 1 argument
				if in.Integer() != 1 {
					return exec.AsValue(endings[0])
				}
			} else {
				if in.Integer() != 1 {
					// 2 arguments
					return exec.AsValue(endings[1])
				}
				return exec.AsValue(endings[0])
			}
		} else {
			if in.Integer() != 1 {
				// return default 's'
				return exec.AsValue("s")
			}
		}

		return exec.AsValue("")
	}
	// return nil, &Error{
	// 	Sender:    "filter:pluralize",
	// 	OrigError: errors.New("filter 'pluralize' does only work on numbers"),
	// }
	return exec.AsValue(errors.New("filter 'pluralize' does only work on numbers"))
}

func filterRemovetags(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	s := in.String()
	tags := strings.Split(params.Args[0].String(), ",")

	// Strip only specific tags
	for _, tag := range tags {
		re := regexp.MustCompile(fmt.Sprintf("</?%s/?>", tag))
		s = re.ReplaceAllString(s, "")
	}

	return exec.AsValue(strings.TrimSpace(s))
}

func filterRjust(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	return exec.AsValue(fmt.Sprintf(fmt.Sprintf("%%%ds", params.Args[0].Integer()), in.String()))
}

func filterYesno(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if len(params.Args) > 1 {
		// return nil, &Error{
		// 	Sender:    "filter:getdigit",
		// 	OrigError: errors.New("'getdigit' filter expect one and only one argument"),
		// }
		return exec.AsValue(errors.New("'getdigit' filter expect one and only one argument"))
	}
	choices := map[int]string{
		0: "yes",
		1: "no",
		2: "maybe",
	}
	param := params.First()
	paramString := param.String()
	customChoices := strings.Split(paramString, ",")
	if len(paramString) > 0 {
		if len(customChoices) > 3 {
			return exec.AsValue(errors.Errorf("You cannot pass more than 3 options to the 'yesno'-filter (got: '%s').", paramString))
			// return nil, &Error{
			// 	Sender:    "filter:yesno",
			// 	OrigError: errors.Errorf("You cannot pass more than 3 options to the 'yesno'-filter (got: '%s').", paramString),
			// }
		}
		if len(customChoices) < 2 {
			// return nil, &Error{
			// 	Sender:    "filter:yesno",
			// 	OrigError: errors.Errorf("You must pass either no or at least 2 arguments to the 'yesno'-filter (got: '%s').", paramString),
			// }
			return exec.AsValue(errors.Errorf("You must pass either no or at least 2 arguments to the 'yesno'-filter (got: '%s').", paramString))
		}

		// Map to the options now
		choices[0] = customChoices[0]
		choices[1] = customChoices[1]
		if len(customChoices) == 3 {
			choices[2] = customChoices[2]
		}
	}

	// maybe
	if in.IsNil() {
		return exec.AsValue(choices[2])
	}

	// yes
	if in.IsTrue() {
		return exec.AsValue(choices[0])
	}

	// no
	return exec.AsValue(choices[1])
}
