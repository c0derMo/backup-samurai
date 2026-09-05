package lib

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SamuraiLogger struct {
	File    *os.File
	Verbose bool
}

// Taken from zerolog sourcecode and adapted
func consoleDefaultFormatLevel(noColor bool) zerolog.Formatter {
	if noColor {
		return func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("%-6s |", i))
		}
	}
	return func(i interface{}) string {
		l := strings.ToUpper(fmt.Sprintf("%-6s", i))
		if ll, ok := i.(string); ok {
			level, _ := zerolog.ParseLevel(ll)
			color := zerolog.LevelColors[level]
			l = fmt.Sprintf("\x1b[%dm%v\x1b[0m", color, l)
		}
		return fmt.Sprintf("%s |", l)
	}
}

func InitializeLogging(verbose bool, file string) SamuraiLogger {
	sl := SamuraiLogger{
		File:    nil,
		Verbose: verbose,
	}

	sl.setLoggingLevel()
	if file == "-" {
		sl.createConsoleLogger()
	} else {
		sl.createFileLogger(file)
	}

	return sl
}

func (sl SamuraiLogger) Cleanup() {
	if sl.File != nil {
		sl.File.Close()
	}
}

func (sl SamuraiLogger) createConsoleWriter(out io.Writer, noColor bool) zerolog.ConsoleWriter {
	logger := zerolog.ConsoleWriter{Out: out}
	logger.FormatLevel = consoleDefaultFormatLevel(noColor)
	return logger
}

func (sl SamuraiLogger) createConsoleLogger() {
	log.Logger = log.Output(sl.createConsoleWriter(os.Stdout, false))
}

func (sl SamuraiLogger) createFileLogger(file string) {
	f, err := os.Create(file)
	if err != nil {
		sl.createConsoleLogger()
		log.Err(err).Msg("An error occurred while opening the logging file. Logs will be sent to stdout instead.")
		return
	}
	sl.File = f
	logger := sl.createConsoleWriter(f, true)
	logger.NoColor = true
	log.Logger = log.Output(logger)
}

func (sl SamuraiLogger) setLoggingLevel() {
	if sl.Verbose {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
