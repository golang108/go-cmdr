// Package logz is a tiny wrapper to logg/slog.
package logz

import (
	"context"
	"fmt"
	"sync"

	stdlog "log/slog"

	"github.com/hedzr/is"
	logz "github.com/hedzr/logg/slog"
)

var Logger *stdlog.Logger
var log logz.Logger
var onceLog sync.Once

func Info(msg string, args ...any) {
	// Logger.Info(msg, args...) // NOTE, std log/slog cannot ignore extra stack frame(s)
	log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	log.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

func Trace(msg string, args ...any) {
	log.Trace(msg, args...)
}

func Verbose(msg string, args ...any) {
	log.Verbose(msg, args...)
}

func Panic(msg string, args ...any) {
	log.Panic(msg, args...)
}

func Fatal(msg string, args ...any) {
	log.Fatal(msg, args...)
}

func Print(msg string, args ...any) {
	log.Print(msg, args...)
}

func Println(args ...any) {
	log.Println(args...)
}

func Printf(msg string, args ...any) {
	log.Println(fmt.Sprintf(msg, args...))
}

func OK(msg string, args ...any) {
	log.OK(msg, args...)
}

func Fail(msg string, args ...any) {
	log.Fail(msg, args...)
}

func Success(msg string, args ...any) {
	log.Success(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	log.InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	log.WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	log.ErrorContext(ctx, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	log.DebugContext(ctx, msg, args...)
}

func TraceContext(ctx context.Context, msg string, args ...any) {
	// if is.Tracing() {
	// 	log.DebugContext(ctx, msg, args...)
	// }
	log.TraceContext(ctx, msg, args...)
}

func VerboseContext(ctx context.Context, msg string, args ...any) {
	// if is.VerboseBuild() {
	// 	log.DebugContext(ctx, msg, args...)
	// }
	log.VerboseContext(ctx, msg, args...)
}

func PanicContext(ctx context.Context, msg string, args ...any) {
	log.PanicContext(ctx, msg, args...)
}

func FatalContext(ctx context.Context, msg string, args ...any) {
	log.FatalContext(ctx, msg, args...)
}

func PrintContext(ctx context.Context, msg string, args ...any) {
	log.PrintContext(ctx, msg, args...)
}

func PrintlnContext(ctx context.Context, msg string, args ...any) {
	log.PrintlnContext(ctx, msg, args...)
}

func OKContext(ctx context.Context, msg string, args ...any) {
	log.OKContext(ctx, msg, args...)
}

func FailContext(ctx context.Context, msg string, args ...any) {
	log.FailContext(ctx, msg, args...)
}

func SuccessContext(ctx context.Context, msg string, args ...any) {
	log.SuccessContext(ctx, msg, args...)
}

func SetLevel(level logz.Level) {
	log.SetLevel(level)
	logz.SetLevel(level)
}

func GetLevel() logz.Level { return log.Level() }

func SetJSONMode(mode bool) {
	log.SetJSONMode(mode)
}

func Skip(nFrames int) logz.Logger {
	return log.WithSkip(nFrames)
}

//

//

// WrappedLogger returns a reference to *slog.Logger which was
// wrapped to hedzr/logg/slog.
//
// In most cases, you'd better use dbglog.Info/... directly because
// these forms can locate the preferred stack frame(s) of the caller.
func WrappedLogger() *stdlog.Logger { return Logger }

func ifn(lvl logz.Level) {
	sll := logz.NewSlogHandler(log, &logz.HandlerOptions{
		NoColor:  false,
		NoSource: false,
		JSON:     false,
		Level:    logz.InfoLevel,
	})

	Logger = stdlog.New(sll)
}

func init() {
	onceLog.Do(func() {
		log00 := logz.New("[cmdr]") // .SetLevel(logz.DebugLevel)
		log = log00.                // make a child logger here.
						WithSkip(1) // extra stack frame(s) shall be ignored for dbglog.Info/...
		log00.Verbose("init logz")

		ifn(logz.InfoLevel)
		is.SetOnDebugChanged(func(mod bool, level int) {
			if mod {
				lvl := logz.DebugLevel
				ifn(lvl)
				log00.SetLevel(lvl)
				log.SetLevel(lvl)
				Debug("[cmdr.v2.logz][onDebugChanged] debug mode changed", "mode", mod, "level", level, "log-level", lvl)
			}
		})
		is.SetOnTraceChanged(func(mod bool, level int) {
			if mod {
				lvl := logz.TraceLevel
				ifn(lvl)
				log00.SetLevel(lvl)
				log.SetLevel(lvl)
				Trace("[cmdr.v2.logz][onTraceChanged] trace mode changed", "mode", mod, "level", level, "log-level", lvl)
			}
		})

		// ctx := context.Background()
		// InfoContext(ctx, "hello, world")
		// DebugContext(ctx, "hello, world")
	})
}
