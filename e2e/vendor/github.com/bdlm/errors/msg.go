package errors

/*
ErrMsg defines the interface to error message data.
*/
type ErrMsg interface {
	Caller() Caller
	Code() Code
	Error() string
	Msg() string
	SetCode(Code) ErrMsg
}

/*
Msg defines a single error message.
*/
type Msg struct {
	err    error
	caller Caller
	code   Code
	msg    string
}

/*
Caller implements ErrMsg
*/
func (msg Msg) Caller() Caller {
	return msg.caller
}

/*
Code implements ErrMsg
*/
func (msg Msg) Code() Code {
	return msg.code
}

/*
Error implements error
*/
func (msg Msg) Error() string {
	return msg.String()
}

/*
Msg implements ErrMsg
*/
func (msg Msg) Msg() string {
	return msg.msg
}

/*
SetCode implements ErrMsg
*/
func (msg Msg) SetCode(code Code) ErrMsg {
	msg.code = code
	return msg
}

/*
String implements Stringer
*/
func (msg Msg) String() string {
	if nil == msg.err {
		return msg.msg
	}
	return msg.err.Error()
}
