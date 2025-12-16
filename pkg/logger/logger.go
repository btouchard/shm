// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"fmt"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func log(level Level, color, context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if context != "" {
		fmt.Printf("%s[%s]%s %s[%s]%s %s[%s]%s %s\n",
			Gray, timestamp(), Reset,
			color, level, Reset,
			Cyan, context, Reset,
			message)
	} else {
		fmt.Printf("%s[%s]%s %s[%s]%s %s\n",
			Gray, timestamp(), Reset,
			color, level, Reset,
			message)
	}
}

// Debug log en mode debug
func Debug(format string, args ...interface{}) {
	log(LevelDebug, Gray, "", format, args...)
}

// DebugCtx log debug avec contexte
func DebugCtx(context, format string, args ...interface{}) {
	log(LevelDebug, Gray, context, format, args...)
}

// Info log une information
func Info(format string, args ...interface{}) {
	log(LevelInfo, Green, "", format, args...)
}

// InfoCtx log info avec contexte
func InfoCtx(context, format string, args ...interface{}) {
	log(LevelInfo, Green, context, format, args...)
}

// Warn log un avertissement
func Warn(format string, args ...interface{}) {
	log(LevelWarn, Yellow, "", format, args...)
}

// WarnCtx log warning avec contexte
func WarnCtx(context, format string, args ...interface{}) {
	log(LevelWarn, Yellow, context, format, args...)
}

// Error log une erreur
func Error(format string, args ...interface{}) {
	log(LevelError, Red, "", format, args...)
}

// ErrorCtx log error avec contexte
func ErrorCtx(context, format string, args ...interface{}) {
	log(LevelError, Red, context, format, args...)
}
