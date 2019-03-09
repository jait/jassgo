/* vim: set sts=4 sw=4 ts=4 noet: */
/**
 * jass - just another sudoku solver
 * (C) 2005-2019 Jari Tenhunen <jait@iki.fi>
 *
 * Go version copyright 2019
 */

package jass

import "fmt"

type LogLevel int

const (
	LogNone   LogLevel = 0
	LogNormal LogLevel = 1
	LogDebug  LogLevel = 2
)

var logLevel LogLevel

func SetLogLevel(level LogLevel) {
	logLevel = level
}

func Debug(s string, a ...interface{}) {
	if logLevel < LogDebug {
		return
	}
	fmt.Printf(s, a...)
	fmt.Printf("\n")
}

func Explain(s string, a ...interface{}) {
	Debug(s, a...)
}

func Info(s string, a ...interface{}) {
	fmt.Printf(s, a...)
	fmt.Printf("\n")
}
