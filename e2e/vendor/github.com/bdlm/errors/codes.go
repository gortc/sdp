package errors

/*
Internal errors
*/
const (
	// ErrUnknown - 0: An unknown error occurred.
	ErrUnknown Code = iota
	// ErrFatal - 1: An fatal error occurred.
	ErrFatal
)

/*
I/O errors
*/
const (
	// ErrEOF - 100: An invalid HTTP method was requested.
	ErrEOF Code = iota + 100
	ErrReader
)

/*
Encoding errors
*/
const (
	// ErrDecodingFailed - Decoding failed due to an error with the data.
	ErrDecodingFailed Code = iota + 200
	// ErrDecodingJSON - JSON data could not be decoded.
	ErrDecodingJSON
	// ErrDecodingToml - Toml data could not be decoded.
	ErrDecodingToml
	// ErrDecodingYaml - Yaml data could not be decoded.
	ErrDecodingYaml
	// ErrEncodingFailed - Encoding failed due to an error with the data.
	ErrEncodingFailed
	// ErrEncodingJSON - JSON data could not be encoded.
	ErrEncodingJSON
	// ErrEncodingToml - Toml data could not be encoded.
	ErrEncodingToml
	// ErrEncodingYaml - Yaml data could not be encoded.
	ErrEncodingYaml
	// ErrInvalidJSON - Data is not valid JSON.
	ErrInvalidJSON
	// ErrInvalidToml - Data is not valid Toml.
	ErrInvalidToml
	// ErrInvalidYaml - Data is not valid Yaml.
	ErrInvalidYaml
	// ErrTypeConversionFailed - Data type conversion failed.
	ErrTypeConversionFailed
)

/*
Server errors
*/
const (
	// ErrInvalidHTTPMethod - 300: An invalid HTTP method was requested.
	ErrInvalidHTTPMethod Code = iota + 300
)

func init() {
	// Internal errors
	Codes[ErrUnknown] = ErrCode{"An unknown error occurred", "", 0}
	Codes[ErrFatal] = ErrCode{"A fatal error occurred", "a fatal error occurred", 0}

	// I/O errors
	Codes[ErrEOF] = ErrCode{"End of input", "unexpected EOF", 0}
	Codes[ErrReader] = ErrCode{"Read failed", "read failed", 0}

	// Encoding errors
	Codes[ErrDecodingJSON] = ErrCode{"JSON data could not be decoded", "JSON data could not be decoded", 0}
	Codes[ErrDecodingToml] = ErrCode{"TOML data could not be decoded", "TOML data could not be decoded", 0}
	Codes[ErrDecodingYaml] = ErrCode{"YAML data could not be decoded", "YAML data could not be decoded", 0}
	Codes[ErrEncodingJSON] = ErrCode{"JSON data could not be encoded", "JSON data could not be encoded", 0}
	Codes[ErrEncodingToml] = ErrCode{"TOML data could not be encoded", "TOML data could not be encoded", 0}
	Codes[ErrEncodingYaml] = ErrCode{"YAML data could not be encoded", "YAML data could not be encoded", 0}
	Codes[ErrTypeConversionFailed] = ErrCode{"Data type conversion failed", "data type conversion failed", 0}

	// Server errors
	Codes[ErrInvalidHTTPMethod] = ErrCode{"Invalid HTTP Method", "an invalid HTTP method was requested", 0}
}

/*
Coder defines an interface for an error code.
*/
type Coder interface {
	// Internal only (logs) error text.
	Detail() string
	// HTTP status that should be used for the associated error code.
	HTTPStatus() int
	// External (user) facing error text.
	String() string
}

/*
Code defines an error code type.
*/
type Code int

/*
Codes contains a map of error codes to metadata
*/
var Codes = map[Code]Coder{}

/*
ErrCode implements coder
*/
type ErrCode struct {
	// External (user) facing error text.
	Ext string
	// Internal only (logs) error text.
	Int string
	// HTTP status that should be used for the associated error code.
	HTTP int
}

/*
Detail returns the internal error message, if any.
*/
func (code ErrCode) Detail() string {
	return code.Int
}

/*
String implements stringer. String returns the external error message,
if any.
*/
func (code ErrCode) String() string {
	return code.Ext
}

/*
HTTPStatus returns the associated HTTP status code, if any. Otherwise,
returns 200.
*/
func (code ErrCode) HTTPStatus() int {
	if 0 == code.HTTP {
		return 200
	}
	return code.HTTP
}
