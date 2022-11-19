package sub

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Render(t *testing.T) {
	fIn, err := os.CreateTemp("", "goja-cli-test")
	assert.NoError(t, err)
	defer os.Remove(fIn.Name())
	fIn.WriteString("{{ FOO }}")
	fIn.Sync()
	os.Setenv("FOO", "bar")
	fOut, err := os.CreateTemp("", "goja-cli-out")
	assert.NoError(t, err)
	defer os.Remove(fOut.Name())
	rCmd(nil, []string{fIn.Name(), fOut.Name()})
	assert.NoError(t, err)
	assert.FileExists(t, fOut.Name())
	file, err := ioutil.ReadFile(fOut.Name())
	assert.NoError(t, err)
	assert.Equal(t, "bar\n", string(file))
}
