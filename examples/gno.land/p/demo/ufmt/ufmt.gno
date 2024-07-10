// Package ufmt provides utility functions for formatting strings, similarly
// to the Go package "fmt", of which only a subset is currently supported
// (hence the name µfmt - micro fmt).
package ufmt

import (
	"errors"
	"strconv"
	"strings"
)

// Println formats using the default formats for its operands and writes to standard output.
// Println writes the given arguments to standard output with spaces between arguments
// and a newline at the end.
func Println(args ...interface{}) {
	var strs []string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			strs = append(strs, v)
		case (interface{ String() string }):
			strs = append(strs, v.String())
		case error:
			strs = append(strs, v.Error())
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			strs = append(strs, Sprintf("%d", v))
		case bool:
			if v {
				strs = append(strs, "true")
			} else {
				strs = append(strs, "false")
			}
		case nil:
			strs = append(strs, "<nil>")
		default:
			strs = append(strs, "(unhandled)")
		}
	}

	// TODO: remove println after gno supports os.Stdout
	println(strings.Join(strs, " "))
}

// Sprintf offers similar functionality to Go's fmt.Sprintf, or the sprintf
// equivalent available in many languages, including C/C++.
// The number of args passed must exactly match the arguments consumed by the format.
// A limited number of formatting verbs and features are currently supported,
// hence the name ufmt (µfmt, micro-fmt).
//
// The currently formatted verbs are the following:
//
//	%s: places a string value directly.
//	    If the value implements the interface interface{ String() string },
//	    the String() method is called to retrieve the value. Same about Error()
//	    string.
//	%c: formats the character represented by Unicode code point
//	%d: formats an integer value using package "strconv".
//	    Currently supports only uint, uint64, int, int64.
//	%t: formats a boolean value to "true" or "false".
//	%%: outputs a literal %. Does not consume an argument.
func Sprintf(format string, args ...interface{}) string {
	argNum := 0
	argLen := len(args)
	var buf strings.Builder
	verb := "" // a string like "%s". Single "%" is incompletely read
	for _, curRune := range format {
		if verb == "" {
			if curRune == '%' {
				verb = "%"
				continue
			} else {
				buf.WriteRune(curRune)
			}
		} else {
			// complete verb
			var arg interface{}
			if curRune != '%' { // don't consume/use arg upon "%%"
				if argNum > argLen {
					panic("invalid number of arguments to ufmt.Sprintf: format=" + format + " nbArgs=" + string(argNum))
				}
				arg = args[argNum]
				argNum++
			}
			verb = verb + string(curRune)
			switch curRune {
			case '%':
				buf.WriteRune('%')
			case 's':
				switch v := arg.(type) {
				case interface{ String() string }:
					buf.WriteString(v.String())
				case error:
					buf.WriteString(v.Error())
				case string:
					buf.WriteString(v)
				default:
					buf.WriteString(fallback(verb, v))
				}
			case 'c':
				switch v := arg.(type) {
				case rune:
					buf.WriteRune(v)
				case int:
					buf.WriteString(string(v))
				case int8:
					buf.WriteString(string(v))
				case int16:
					buf.WriteString(string(v))
				case uint:
					buf.WriteString(string(v))
				case uint8:
					buf.WriteString(string(v))
				case uint16:
					buf.WriteString(string(v))
				default:
					buf.WriteString(fallback(verb, v))
				}
			case 'd':
				switch v := arg.(type) {
				case int:
					buf.WriteString(strconv.Itoa(v))
				case int8:
					buf.WriteString(strconv.Itoa(int(v)))
				case int16:
					buf.WriteString(strconv.Itoa(int(v)))
				case int32:
					buf.WriteString(strconv.Itoa(int(v)))
				case int64:
					buf.WriteString(strconv.Itoa(int(v)))
				case uint:
					buf.WriteString(strconv.FormatUint(uint64(v), 10))
				case uint8:
					buf.WriteString(strconv.FormatUint(uint64(v), 10))
				case uint16:
					buf.WriteString(strconv.FormatUint(uint64(v), 10))
				case uint32:
					buf.WriteString(strconv.FormatUint(uint64(v), 10))
				case uint64:
					buf.WriteString(strconv.FormatUint(v, 10))
				default:
					buf.WriteString(fallback(verb, v))
				}
			case 't':
				switch v := arg.(type) {
				case bool:
					if v {
						buf.WriteString("true")
					} else {
						buf.WriteString("false")
					}
				default:
					buf.WriteString(fallback(verb, v))
				}
			default:
				buf.WriteString("(unhandled verb: %" + verb + ")")
			}
			verb = ""
		}
	}
	if verb != "" {
		buf.WriteString("%!(NOVERB)") // as in Go fmt.Sprintf("%")
	}
	if argNum < argLen {
		panic("too many arguments to ufmt.Sprintf: format=" + format + " nbArgs=" + string(argNum))
	}
	return buf.String()
}

// This function is used to mimic Go's fmt.Sprintf
// specific behaviour of showing verb/type mismatches,
// where for example:
//
//	fmt.Sprintf("%d", "foo") gives "%!d(string=foo)"
//
// Here:
//
//	fallback("%s", 8) -> "%!s(int=8)"
//	fallback("%d", nil) -> "%!d(<nil>)", and so on.
func fallback(verb string, arg interface{}) string {
	if !strings.HasPrefix(verb, "%") {
		panic("Invalid Sprintf verb " + verb)
	}
	verbLetter := verb[1:]
	var s string
	switch v := arg.(type) {
	case string:
		s = "string=" + v
	case (interface{ String() string }):
		s = "string=" + v.String()
	case error:
		// note: also "string=" in Go fmt
		s = "string=" + v.Error()
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// note: rune, byte would be dups, being aliases
		if typename, e := typeToString(v); e != nil {
			panic("should not happen")
		} else {
			s = typename + "=" + Sprintf("%d", v)
		}
	case bool:
		if v {
			s = "bool=true"
		} else {
			s = "bool=false"
		}
	case nil:
		s = "<nil>"
	default:
		s = "(unhandled)"
	}
	return "%!" + verbLetter + "(" + s + ")"
}

// Get the name of the type of `v` as a string.
// The recognized type of v is currently limited to native non-composite types.
// An error is returned otherwise.
func typeToString(v interface{}) (string, error) {
	switch v.(type) {
	case string:
		return "string", nil
	case int:
		return "int", nil
	case int8:
		return "int8", nil
	case int16:
		return "int16", nil
	case int32:
		return "int32", nil
	case int64:
		return "int64", nil
	case uint:
		return "uint", nil
	case uint8:
		return "uint8", nil
	case uint16:
		return "uint16", nil
	case uint32:
		return "uint32", nil
	case uint64:
		return "uint64", nil
	case float32:
		return "float32", nil
	case float64:
		return "float64", nil
	case bool:
		return "bool", nil
	default:
		return "", errors.New("(unsupported type)")
	}
}

// This function is used to mimic Go's fmt.Sprintf
// specific behaviour of showing verb/type mismatches,
// where for example:
//
//	fmt.Sprintf("%d", "foo") gives "%!d(string=foo)"
//
// Here:
//
//	fallback("s", 8) -> "%!s(int=8)"
//	fallback("d", nil) -> "%!d(<nil>)", and so on.
func fallback(verb string, arg interface{}) string {
	var s string
	switch v := arg.(type) {
	case string:
		s = "string=" + v
	case (interface{ String() string }):
		s = "string=" + v.String()
	case error:
		// note: also "string=" in Go fmt
		s = "string=" + v.Error()
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// note: rune, byte would be dups, being aliases
		if typename, e := typeToString(v); e != nil {
			panic("should not happen")
		} else {
			s = typename + "=" + Sprintf("%d", v)
		}
	case bool:
		if v {
			s = "bool=true"
		} else {
			s = "bool=false"
		}
	case nil:
		s = "<nil>"
	default:
		s = "(unhandled)"
	}
	return "%!" + verb + "(" + s + ")"
}

// Get the name of the type of `v` as a string.
// The recognized type of v is currently limited to native non-composite types.
// An error is returned otherwise.
func typeToString(v interface{}) (string, error) {
	switch v.(type) {
	case string:
		return "string", nil
	case int:
		return "int", nil
	case int8:
		return "int8", nil
	case int16:
		return "int16", nil
	case int32:
		return "int32", nil
	case int64:
		return "int64", nil
	case uint:
		return "uint", nil
	case uint8:
		return "uint8", nil
	case uint16:
		return "uint16", nil
	case uint32:
		return "uint32", nil
	case uint64:
		return "uint64", nil
	case float32:
		return "float32", nil
	case float64:
		return "float64", nil
	case bool:
		return "bool", nil
	default:
		return "", errors.New("(unsupported type)")
	}
}

// errMsg implements the error interface.
type errMsg struct {
	msg string
}

// Error defines the requirements of the error interface.
// It functions similarly to Go's errors.New()
func (e *errMsg) Error() string {
	return e.msg
}

// Errorf is a function that mirrors the functionality of fmt.Errorf.
//
// It takes a format string and arguments to create a formatted string,
// then sets this string as the 'msg' field of an errMsg struct and returns a pointer to this struct.
//
// This function operates in a similar manner to Go's fmt.Errorf,
// providing a way to create formatted error messages.
//
// The currently formatted verbs are the following:
//
//	%s: places a string value directly.
//	    If the value implements the interface interface{ String() string },
//	    the String() method is called to retrieve the value. Same for error.
//	%c: formats the character represented by Unicode code point
//	%d: formats an integer value using package "strconv".
//	    Currently supports only uint, uint64, int, int64.
//	%t: formats a boolean value to "true" or "false".
//	%%: outputs a literal %. Does not consume an argument.
func Errorf(format string, args ...interface{}) error {
	return &errMsg{Sprintf(format, args...)}
}
