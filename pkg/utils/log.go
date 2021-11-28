package utils

import (
	"fmt"
	"io"
	stdlog "log"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func Colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func SetupGlobalLogger(outW io.Writer) zerolog.Logger {
	out := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = outW
		w.PartsExclude = []string{zerolog.CallerFieldName, zerolog.TimestampFieldName}
		w.PartsOrder = []string{zerolog.LevelFieldName, zerolog.MessageFieldName}
		w.FormatMessage = func(v interface{}) string {
			return Colorize(v, 35, false)
		}
		w.FormatFieldValue = func(v interface{}) string {
			return Colorize(v, 32, false)
		}
	})

	log.Logger = log.Output(out)
	stdlog.SetOutput(log.Logger)
	stdlog.SetFlags(0)
	return log.Logger
}
