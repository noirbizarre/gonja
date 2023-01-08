package time

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/bmuller/arrow"

	"github.com/paradime-io/gonja/exec"
	"github.com/paradime-io/gonja/nodes"
	"github.com/paradime-io/gonja/parser"
	"github.com/paradime-io/gonja/tokens"
)

type TimeOffset struct {
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

type NowStmt struct {
	Location *tokens.Token
	TZ       string
	Format   string
	Offset   *TimeOffset
}

func (stmt *NowStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *NowStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("NowStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *NowStmt) Execute(r *exec.Renderer, tag *nodes.StatementBlock) error {
	var now arrow.Arrow

	cfg := r.Config.Ext["time"].(*Config)
	format := cfg.DatetimeFormat

	if cfg.Now != nil {
		now = *cfg.Now
	} else {
		now = arrow.Now()
	}

	if stmt.Format != "" {
		format = stmt.Format
	}

	now = now.InTimezone(stmt.TZ)

	if stmt.Offset != nil {
		offset := stmt.Offset
		if offset.Years != 0 || offset.Months != 0 || offset.Days != 0 {
			now = arrow.New(now.AddDate(offset.Years, offset.Months, offset.Days))
		}
		if offset.Hours != 0 {
			now = now.AddHours(offset.Hours)
		}
		if offset.Minutes != 0 {
			now = now.AddMinutes(offset.Minutes)
		}
		if offset.Seconds != 0 {
			now = now.AddSeconds(offset.Seconds)
		}
	}

	r.WriteString(now.CFormat(format))

	return nil
}

func nowParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	stmt := &NowStmt{
		Location: p.Current(),
	}

	// Timezone
	tz := args.Match(tokens.String)
	if tz == nil {
		return nil, args.Error(`now expect a timezone as first argument`, args.Current())
	}
	stmt.TZ = tz.Val

	// Offset
	if sign := args.Match(tokens.Add, tokens.Sub); sign != nil {
		offset := args.Match(tokens.String)
		if offset == nil {
			return nil, args.Error("Expected an time offset.", args.Current())
		}
		timeOffset, err := parseTimeOffset(offset.Val, sign.Val == "+")
		if err != nil {
			return nil, errors.Wrapf(err, `Unable to parse time offset '%s'`, offset.Val)
		}
		stmt.Offset = timeOffset
	}

	// Format
	if args.Match(tokens.Comma) != nil {
		format := args.Match(tokens.String)
		if format == nil {
			return nil, args.Error("Expected a format string.", args.Current())
		}
		stmt.Format = format.Val
	}

	if !args.End() {
		return nil, args.Error("Malformed now-tag args.", nil)
	}

	return stmt, nil
}

func parseTimeOffset(offset string, add bool) (*TimeOffset, error) {
	pairs := strings.Split(offset, ",")
	specs := map[string]int{}
	for _, pair := range pairs {
		splitted := strings.Split(pair, "=")
		if len(splitted) != 2 {
			return nil, errors.Errorf(`Expected a key=value pair, got '%s'`, pair)
		}
		unit := strings.TrimSpace(splitted[0])
		value, err := strconv.Atoi(strings.TrimSpace(splitted[1]))
		if err != nil {
			return nil, errors.Wrap(err, `Unable to parse int`)
		}
		specs[unit] = value
	}
	to := &TimeOffset{}
	for unit, value := range specs {
		if !add {
			value = -value
		}
		switch strings.ToLower(unit) {
		case "year", "years":
			to.Years = value
		case "month", "months":
			to.Months = value
		case "day", "days":
			to.Days = value
		case "hour", "hours":
			to.Hours = value
		case "minute", "minutes":
			to.Minutes = value
		case "second", "seconds":
			to.Seconds = value
		default:
			return nil, errors.Errorf(`Unknown unit '%s`, unit)
		}
	}
	return to, nil
}

func init() {
	Statements.Register("now", nowParser)
}
