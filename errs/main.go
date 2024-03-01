package errs

import (
	"fmt"
	"runtime"
	"strings"
)

func NewFull(code int, err error, format string, a ...any) *Error {
	return &Error{Code: code, err: err, msg: fmt.Sprintf(format, a...), Stack: CallStack(3, 30)}
}

func NewMsg(code int, format string, a ...any) *Error {
	return &Error{Code: code, msg: fmt.Sprintf(format, a...), Stack: CallStack(3, 30)}
}

func New(code int, err error) *Error {
	return &Error{Code: code, err: err, Stack: CallStack(3, 30)}
}

func (e *Error) Short() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message())
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%d] %s\n%s", e.Code, e.Message(), e.Stack)
}

func (e *Error) Message() string {
	if e.err == nil {
		return e.msg
	}
	var errMsg string
	if PrintErr != nil {
		errMsg = PrintErr(e.err)
	} else {
		errMsg = e.err.Error()
	}
	if e.msg == "" {
		return errMsg
	}
	return fmt.Sprintf("%s %s", e.msg, errMsg)
}

func (e *Error) Unwrap() error {
	return e.err
}

func CallStack(skip, maxNum int) string {
	pc := make([]uintptr, maxNum)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	var texts = make([]string, 0, 16)
	for {
		f, more := frames.Next()
		texts = append(texts, fmt.Sprintf("  at %v:%v %v", f.File, f.Line, f.Function))
		if !more {
			break
		}
	}
	return strings.Join(texts, "\n")
}
