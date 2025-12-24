package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level  Level
	logger *log.Logger
	json   bool
}

func New(level string, output io.Writer, format string) *Logger {
	if output == nil {
		output = os.Stdout
	}

	// change flag logera
	return &Logger{
		level:  parseLevel(level),
		logger: log.New(output, "", log.Lshortfile),
		json:   strings.ToLower(format) == "json",
	}
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.log("DEBUG", msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.log("INFO", msg, args...)
	}
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.log("WARN", msg, args...)
	}
}

func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= LevelError {
		l.log("ERROR", msg, args...)
	}
}

func (l *Logger) log(level, msg string, args ...interface{}) {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	if l.json {
		fields := map[string]interface{}{
			"time":  timestamp,
			"level": level,
			"msg":   msg,
		}

		for i := 0; i+1 < len(args); i += 2 {
			k := fmt.Sprint(args[i])
			fields[k] = args[i+1]
		}

		b, _ := json.Marshal(fields)
		l.logger.Println(string(b))
		return
	}

	var kv strings.Builder
	for i := 0; i+1 < len(args); i += 2 {
		kv.WriteString(fmt.Sprintf(" %v=%v", args[i], args[i+1]))
	}

	l.logger.Printf("[%s] %s: %s%s", timestamp, level, msg, kv.String())
}

func parseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
