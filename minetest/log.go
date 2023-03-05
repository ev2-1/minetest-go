package minetest

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type lineParams struct {
	Level   string
	Time    time.Time
	File    string
	Content string
}

func (l lineParams) strS() []string {
	return []string{
		l.Level,
		l.Time.Format("2006-01-02 15:04:05"),
		l.File,
		l.Content,
	}
}

func (l lineParams) strRaw() []string {
	return []string{
		l.Level,
		l.Time.Format("2006-01-02T15:04:05"),
		l.File,
		strings.TrimSuffix(l.Content, "\n"),
	}
}

func logLine(line lineParams) {
	logWriterMu.Lock()
	defer logWriterMu.Unlock()

	err := logWriter.Write(line.strRaw())
	if err != nil {
		panic(fmt.Sprintf("Failed to log: %s\n", err))
	}

	logWriter.Flush()

	log := strings.Join(line.strS()[1:], " ")

	//make sure line has a newline at the end
	if !strings.HasSuffix(log, "\n") {
		log += "\n"
	}

	fmt.Fprintf(os.Stderr, "%s %s", strings.ToUpper(line.Level[:1]), log)
}

var (
	logWriter   *csv.Writer
	logWriterMu sync.Mutex
)

func initLog() {
	logWriterMu.Lock()
	defer logWriterMu.Unlock()
	if logWriter != nil {
		return
	}

	f, err := os.OpenFile(Path(GetConfigV("log-file", "latest.log")), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	logWriter = csv.NewWriter(f)
}

var Loggers interface {
	Default(s string, depth int)
	Defaultf(s string, depth int, v ...any)

	Info(s string, depth int)
	Infof(s string, depth int, v ...any)

	Verbose(s string, depth int)
	Verbosef(s string, depth int, v ...any)

	Warn(s string, depth int)
	Warnf(s string, depth int, v ...any)

	Error(s string, depth int)
	Errorf(s string, depth int, v ...any)

	ll()
} = logLevels{}

type logLevels struct{}

func (logLevels) ll() {} // Dummy func to prevent overwriting the variable

func (logLevels) Default(str string, depth int) {
	initLog()

	logLine(lineParams{
		Level:   "default",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: str,
	})
}

func (logLevels) Defaultf(str string, depth int, v ...any) {
	initLog()

	logLine(lineParams{
		Level:   "default",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: fmt.Sprintf(str, v...),
	})
}

func (logLevels) Info(str string, depth int) {
	initLog()

	logLine(lineParams{
		Level:   "info",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: str,
	})
}

func (logLevels) Infof(str string, depth int, v ...any) {
	initLog()

	logLine(lineParams{
		Level:   "info",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: fmt.Sprintf(str, v...),
	})
}

func (logLevels) Verbose(str string, depth int) {
	initLog()

	logLine(lineParams{
		Level:   "verbose",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: str,
	})
}

func (logLevels) Verbosef(str string, depth int, v ...any) {
	initLog()

	logLine(lineParams{
		Level:   "verbose",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: fmt.Sprintf(str, v...),
	})
}

func (logLevels) Warn(str string, depth int) {
	initLog()

	logLine(lineParams{
		Level:   "warn",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: str,
	})
}

func (logLevels) Warnf(str string, depth int, v ...any) {
	initLog()

	logLine(lineParams{
		Level:   "warn",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: fmt.Sprintf(str, v...),
	})
}

func (logLevels) Error(str string, depth int) {
	initLog()

	logLine(lineParams{
		Level:   "error",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: str,
	})
}

func (logLevels) Errorf(str string, depth int, v ...any) {
	initLog()

	logLine(lineParams{
		Level:   "error",
		Time:    time.Now(),
		File:    SCaller(depth),
		Content: fmt.Sprintf(str, v...),
	})
}
