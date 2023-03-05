// This just wraps the minetest.LogLevels struct
package log

import (
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
)

func Default(str string) {
	minetest.Loggers.Default(str, 2)
}

func Print(str string) {
	minetest.Loggers.Default(str, 2)
}

func Println(str string) {
	minetest.Loggers.Default(str+"\n", 2)
}

func Defaultf(str string, v ...any) {
	minetest.Loggers.Default(fmt.Sprintf(str, v...), 2)
}

func Printf(str string, v ...any) {
	minetest.Loggers.Default(fmt.Sprintf(str, v...), 2)
}

func Info(str string) {
	minetest.Loggers.Info(str, 2)
}

func Infof(str string, v ...any) {
	minetest.Loggers.Info(fmt.Sprintf(str, v...), 2)
}

func Verbose(str string) {
	minetest.Loggers.Verbose(str, 2)
}

func Verbosef(str string, v ...any) {
	minetest.Loggers.Verbose(fmt.Sprintf(str, v...), 2)
}

func Warn(str string) {
	minetest.Loggers.Warn(str, 2)
}

func Warnf(str string, v ...any) {
	minetest.Loggers.Warn(fmt.Sprintf(str, v...), 2)
}

func Error(str string) {
	minetest.Loggers.Error(str, 2)
}

func Errorf(str string, v ...any) {
	minetest.Loggers.Error(fmt.Sprintf(str, v...), 2)
}
