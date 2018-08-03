package errors

import (
	"fmt"
	"runtime"
)

/*
Caller defines an interface to runtime caller results.
*/
type Caller interface {
	File() string
	Line() int
	Ok() bool
	Pc() uintptr
	String() string
}

/*
Call holds runtime.Caller data
*/
type Call struct {
	loaded bool
	file   string
	line   int
	ok     bool
	pc     uintptr
}

/*
File returns the caller file name.
*/
func (call Call) File() string {
	return call.file
}

/*
Line returns the caller line number.
*/
func (call Call) Line() int {
	return call.line
}

/*
Ok returns whether the caller data was successfully recovered.
*/
func (call Call) Ok() bool {
	return call.ok
}

/*
Pc returns the caller program counter.
*/
func (call Call) Pc() uintptr {
	return call.pc
}

/*
String implements the Stringer interface
*/
func (call Call) String() string {
	return fmt.Sprintf(
		"%s:%d %s",
		call.file,
		call.line,
		runtime.FuncForPC(call.pc).Name(),
	)
}

func getCaller() Caller {
	call := Call{}
	call.pc, call.file, call.line, call.ok = runtime.Caller(2)
	return call
}
