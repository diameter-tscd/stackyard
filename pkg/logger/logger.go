package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps the zerolog logger
type Logger struct {
	z     zerolog.Logger
	quiet bool
}

// LoggerConfig contains configuration for the logger
type LoggerConfig struct {
	Debug       bool
	Quiet       bool // suppress console output (logs still go to broadcaster)
	Broadcaster io.Writer
}

// New creates a new fancy logger
func New(debug bool, broadcaster io.Writer) *Logger {
	return NewWithConfig(LoggerConfig{
		Debug:       debug,
		Quiet:       false,
		Broadcaster: broadcaster,
	})
}

// NewQuiet creates a new logger with console output suppressed
func NewQuiet(debug bool, broadcaster io.Writer) *Logger {
	return NewWithConfig(LoggerConfig{
		Debug:       debug,
		Quiet:       true,
		Broadcaster: broadcaster,
	})
}

// NewWithConfig creates a new logger with full configuration
func NewWithConfig(cfg LoggerConfig) *Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	// Console Output (Fancy)
	consoleOutput := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	consoleOutput.FormatLevel = func(i interface{}) string {
		var l string
		if ll, ok := i.(string); ok {
			switch ll {
			case "debug":
				l = "\x1b[38;2;139;233;253m[ DEBUG ]\x1b[0m" // Pastel Cyan
			case "info":
				l = "\x1b[38;2;189;147;249m[ INFO  ]\x1b[0m" // Pastel Purple
			case "warn":
				l = "\x1b[38;2;241;250;140m[ WARN  ]\x1b[0m" // Pastel Yellow
			case "error":
				l = "\x1b[38;2;255;121;198m[ ERROR ]\x1b[0m" // Pastel Pink
			case "fatal":
				l = "\x1b[38;2;255;85;85m[ FATAL ]\x1b[0m" // Pastel Red
			case "panic":
				l = "\x1b[38;2;255;85;85m[ PANIC ]\x1b[0m" // Pastel Red
			default:
				l = strings.ToUpper(ll)
			}
		} else {
			if i == nil {
				l = strings.ToUpper(fmt.Sprintf("%s", i))
			} else {
				l = strings.ToUpper(fmt.Sprintf("%s", i))
			}
		}
		return l
	}
	consoleOutput.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("\x1b[1m%s\x1b[0m", i)
	}

	var multi zerolog.LevelWriter

	if cfg.Quiet {
		// Quiet mode: only write to broadcaster (if available), not to console
		if cfg.Broadcaster != nil {
			// Create a simple console writer for the broadcaster (without stdout)
			broadcasterOutput := zerolog.ConsoleWriter{Out: cfg.Broadcaster, TimeFormat: "15:04:05", NoColor: true}
			multi = zerolog.MultiLevelWriter(broadcasterOutput)
		} else {
			// No broadcaster and quiet mode = discard all logs
			multi = zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: io.Discard})
		}
	} else {
		// Normal mode: write to console and broadcaster
		if cfg.Broadcaster != nil {
			multi = zerolog.MultiLevelWriter(consoleOutput, cfg.Broadcaster)
		} else {
			multi = zerolog.MultiLevelWriter(consoleOutput)
		}
	}

	logLevel := zerolog.InfoLevel
	if cfg.Debug {
		logLevel = zerolog.DebugLevel
	}

	z := zerolog.New(multi).Level(logLevel).With().Timestamp().Logger()

	return &Logger{z: z, quiet: cfg.Quiet}
}

// IsQuiet returns whether the logger is in quiet mode
func (l *Logger) IsQuiet() bool {
	return l.quiet
}

// Info logs an info message
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log(l.z.Info(), msg, keyvals...)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, keyvals ...interface{}) {
	if err != nil {
		l.z.Error().Err(err).Fields(keyvals).Msg(msg)
	} else {
		l.log(l.z.Error(), msg, keyvals...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log(l.z.Debug(), msg, keyvals...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log(l.z.Warn(), msg, keyvals...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error) {
	if err != nil {
		l.z.Fatal().Err(err).Msg(msg)
	} else {
		l.z.Fatal().Msg(msg)
	}
}

func (l *Logger) log(e *zerolog.Event, msg string, keyvals ...interface{}) {
	if len(keyvals)%2 != 0 {
		e.Msg(msg + " (odd number of keyvals caused metadata drop)")
		return
	}
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		e.Interface(key, keyvals[i+1])
	}
	e.Msg(msg)
}
