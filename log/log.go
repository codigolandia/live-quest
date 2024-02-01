package log

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Level int

const (
	All Level = iota
	Debug
	Info
	Warning
	Error
)

func (l Level) String() string {
	switch l {
	case All:
		return " ALL"
	case Debug:
		return "DEBG"
	case Info:
		return "INFO"
	case Warning:
		return "WARN"
	case Error:
		return "ERRR"
	}
	return "INVALID"
}

func (l *Level) Set(v string) error {
	if l == nil {
		l = new(Level)
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return err
	}
	*l = Level(i)
	return nil
}

var LogLevel Level = Info

func init() {
	flag.Var(&LogLevel, "log-level", "O nível de log mínimo")
}

func LOG(l Level, msg string, args ...any) {
	if l < LogLevel {
		return
	}
	fmt.Fprintf(os.Stderr, l.String()+": "+msg+"\n", args...)
}

func D(msg string, args ...any) {
	LOG(Debug, msg, args...)
}

func I(msg string, args ...any) {
	LOG(Info, msg, args...)
}

func W(msg string, args ...any) {
	LOG(Warning, msg, args...)
}

func E(msg string, args ...any) {
	LOG(Error, msg, args...)
}
