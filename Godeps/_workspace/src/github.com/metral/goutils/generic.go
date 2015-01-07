package goutils

import (
	"fmt"
	"log"
	"runtime"
)

type ErrorParams struct {
	Err       error
	Stderr    string
	CallerNum int
}

// check for errors and panic, if found
func CheckForErrors(e ErrorParams) {
	if e.Err != nil {
		pc, fn, line, _ := runtime.Caller(e.CallerNum)
		msg := ""
		if e.Stderr != "" {
			msg = fmt.Sprintf("[Error] in %s[%s:%d] %v: %s",
				runtime.FuncForPC(pc).Name(), fn, line, e.Err, e.Stderr)
		} else {
			msg = fmt.Sprintf("[Error] in %s[%s:%d] %v",
				runtime.FuncForPC(pc).Name(), fn, line, e.Err)
		}
		log.Fatal(msg)
	}
}
