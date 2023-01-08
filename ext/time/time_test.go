package time_test

import (
	"testing"

	"github.com/bmuller/arrow"

	"github.com/paradime-io/gonja"
	"github.com/paradime-io/gonja/ext/time"
	tu "github.com/paradime-io/gonja/testutils"
)

func Env(root string) *gonja.Environment {
	env := tu.TestEnv(root)
	env.Statements.Update(time.Statements)
	cfg := time.NewConfig()
	parsed, _ := arrow.CParse("%Y-%m-%d %H:%M:%S", "1984-06-07 16:40:00")
	cfg.Now = &parsed
	env.Config.Ext["time"] = cfg
	return env
}

func TestTimeStatement(t *testing.T) {
	root := "./testData"
	env := Env(root)
	tu.GlobTemplateTests(t, root, env)
}
