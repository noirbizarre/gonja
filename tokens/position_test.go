package tokens_test

import (
	"testing"

	"github.com/paradime-io/gonja/tokens"
	"github.com/stretchr/testify/assert"
)

const multilineSample = `Hello
{# 
    Multiline comment
#}
World
`

var readablePositionsCases = []struct {
	name string
	pos  int
	line int
	col  int
	char byte
}{
	{"First char", 0, 1, 1, 'H'},
	{"Last char", len(multilineSample) - 1, 5, 6, '\n'},
	{"Anywhere", 14, 3, 5, 'M'},
}

func TestReadablePosition(t *testing.T) {
	for _, rp := range readablePositionsCases {
		test := rp
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equalf(test.char, multilineSample[test.pos],
				`Invalid test, expected "%#U" rune at pos %d, got "%#U"`,
				test.char, test.pos, multilineSample[test.pos])
			line, col := tokens.ReadablePosition(test.pos, multilineSample)
			assert.Equalf(test.line, line, "Expected line %d, got %d", test.line, line)
			assert.Equalf(test.col, col, "Expected col %d, got %d", test.col, col)
		})
	}
}
