package gonja_test

import (
	"flag"
	"os"
	"testing"

	. "github.com/go-check/check"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

var (
	logLevel = flag.String("log.level", "", "Log Level")
)

func TestMain(m *testing.M) {
	flag.Parse()

	log.SetFormatter(&prefixed.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
		ForceFormatting:  true,
	})

	switch *logLevel {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.PanicLevel)
	}
	code := m.Run()
	os.Exit(code)
}
