package common

import (
	"context"
	"runtime/debug"
	"strings"

	"github.com/rhysd/actionlint"
	log "github.com/sirupsen/logrus"
	"github.com/wayneashleyberry/truecolor/pkg/color"
)

type loggerContextKey string

const loggerContextKeyVal = loggerContextKey("logrus.FieldLogger")

// Logger returns the appropriate logger for current context
func Logger(ctx context.Context) log.Ext1FieldLogger {
	val := ctx.Value(loggerContextKeyVal)
	if val != nil {
		if logger, ok := val.(log.Ext1FieldLogger); ok {
			return logger
		}
	}
	return log.StandardLogger()
}

// WithLogger adds a value to the context for the logger
func WithLogger(ctx context.Context, logger log.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerContextKeyVal, logger)
}

var logHeader = func(ctx context.Context, msg string) {
	Logger(ctx).Debugf("%s: %s", color.Color(255, 0, 89).Sprint("stage"), color.Color(89, 255, 0).Sprint(msg))
}

var logEmpty = func(ctx context.Context) {
	Logger(ctx).Debugf("\t%s\u001B[0m\u001B[22m", color.Color(255, 255, 0).Bold().Sprint("empty"))
}

var logWithColour = func(ctx context.Context, msg, k string, v interface{}) {
	if k == "PATH" {
		Logger(ctx).Debugf("\t%s %s", color.Color(255, 165, 00).Sprint(k), color.Color(186, 218, 85).Sprint("="))
		for _, v := range strings.Split(v.(string), `:`) {
			Logger(ctx).Debugf("\t\t%s", color.Color(0, 165, 255).Sprint(v))
		}
	} else {
		if v == "" {
			v = color.Color(255, 255, 0).Bold().Sprint("empty")
		} else {
			v = color.Color(255, 165, 255).Sprint(v)
		}
		Logger(ctx).Debugf("\t%s %s %s\u001B[0m\u001B[22m", color.Color(252, 58, 61).Sprint(k), color.Color(186, 218, 85).Sprint("="), v)
	}
}

func LogString(ctx context.Context, preMsg, postMsg, obj string) {
	Logger(ctx).Debugf("%s '%s' %s\u001B[0m\u001B[22m", preMsg, color.Color(186, 218, 85).Sprint(obj), postMsg)
}

func LogMap(ctx context.Context, msg string, env map[string]string) {
	logHeader(ctx, msg)
	if len(env) > 0 {
		for k, v := range env {
			logWithColour(ctx, msg, k, v)
		}
	} else {
		logEmpty(ctx)
	}
}

func LogMapInterface(ctx context.Context, msg string, env map[string]actionlint.RawYAMLValue) {
	logHeader(ctx, msg)
	if len(env) > 0 {
		for k, v := range env {
			logWithColour(ctx, msg, k, v)
		}
	} else {
		logEmpty(ctx)
	}
}

func LogSlice(ctx context.Context, msg string, env []string) {
	logHeader(ctx, msg)
	if len(env) > 0 {
		for _, v := range env {
			split := strings.SplitN(v, `=`, 2)
			k := split[0]
			v = split[1]
			logWithColour(ctx, msg, k, v)
		}
	} else {
		logEmpty(ctx)
	}
}

func LogMatrixRows(mr map[string]*actionlint.MatrixRow) {
	for k, v := range mr {
		log.Debugf("\t%s:", color.Color(255, 165, 00).Sprint(k))
		for k2, v2 := range v.Values {
			log.Infof("\t\t%s: %s",
				color.Color(255, 255, 0).Sprint(k2),
				color.Color(255, 165, 255).Sprint(v2.String()),
			)
		}
	}
}

func HandlePanic() {
	if err := recover(); err != nil {
		debug.PrintStack()
		log.Fatalf("program panicked during run time: '%#v'", err)
	}
}
