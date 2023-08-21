package gnolang // {{{1

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	// "sort"
	j "github.com/grepsuzette/joeson"
	"github.com/grepsuzette/joeson/helpers"
	// "github.com/jaekwon/testify/assert"
)

func testExpectation(t *testing.T, expectation expectation) {
	t.Helper()
	ast := parseX(expectation.unparsedString)
	sAst := "nil"
	if ast != nil {
		sAst = ast.String()
	}
	allOk := true
	for _, predicate := range expectation.predicates {
		if err := predicate.satisfies(ast, expectation); err != nil {
			allOk = false
			t.Fatalf(
				"%s parsed as %s "+j.BoldRed("ERR")+" %s\n",
				helpers.Escape(expectation.unparsedString),
				sAst,
				err.Error(),
			)
		}
	}
	if allOk {
		var b strings.Builder
		first := true
		for _, v := range expectation.predicates {
			if !first {
				b.WriteString(", ")
			}
			b.WriteString(j.Magenta(strings.TrimPrefix(
				fmt.Sprintf("%#v", v),
				"gnolang.",
			)))
			first = false
		}
		fmt.Printf(
			"%s parsed as %s "+j.Green("✓")+" %s\n",
			j.Green(helpers.Escape(expectation.unparsedString)),
			j.Yellow(helpers.Escape(sAst)),
			"", // b.String(),
		)
	}
}

func doesntMatchError(expect, got string) bool {
	return !strings.HasPrefix(got, expect[len("ERROR"):])
}

type (
	predicate interface {
		satisfies(j.Ast, expectation) error
	}
	expectation struct {
		unparsedString string
		predicates     []predicate
	}
	parsesAs                  struct{ string } // strict string equality
	parsesAsChar              struct{ rune }   // strict string equality
	parsesAsNil               struct{}
	isBasicLit                struct{ kind Word }
	isSelectorExpr            struct{}
	isNameExpr                struct{}
	isCallExpr                struct{}
	errorIs                   struct{ string }
	errorContains             struct{ string }
	noError                   struct{}
	isType                    struct{ string }
	binaryExprEvaluatesAsInt  struct{ int }
	binaryExprEvaluatesAsBool struct{ bool }
	doom                      struct{}
)

var (
	_ predicate = parsesAs{}
	_ predicate = parsesAsChar{}
	_ predicate = parsesAsNil{}
	_ predicate = isBasicLit{}
	_ predicate = isSelectorExpr{}
	_ predicate = isNameExpr{}
	_ predicate = isCallExpr{}
	_ predicate = errorIs{}
	_ predicate = errorContains{}
	_ predicate = noError{}
	_ predicate = isType{}
	_ predicate = binaryExprEvaluatesAsInt{}
	_ predicate = binaryExprEvaluatesAsBool{}

	// doom = stop tests (useful to stop from the middle of the list of
	// tests to inspect one in particular)
	_ predicate = doom{}
)

// expect() is for non-error expectations (a noError{} predicate gets inserted)
// See expectError()
func expect(unparsedString string, preds ...predicate) expectation {
	// insert noError{} at the beginning
	a := make([]predicate, len(preds)+1)
	copy(a[1:], preds)
	a[0] = noError{}
	return expectation{unparsedString, a}
}

func expectError(unparsedString string, expectedError string) expectation {
	return expectation{unparsedString, []predicate{errorIs{expectedError}}}
}

func expectErrorContains(unparsedString string,
	expectedError string,
) expectation {
	return expectation{
		unparsedString,
		[]predicate{errorContains{expectedError}},
	}
}

// this is just a way to stop the program at a certain place
// from the array of tests
func expectDoom() expectation {
	return expectation{"", []predicate{doom{}}}
}

func (expectation expectation) brief() string {
	for _, pred := range expectation.predicates {
		switch v := pred.(type) {
		case parsesAs:
			// the best brief description there is
			return `"` + v.string + `"`
		default:
		}
	}
	return "it's a bit complicated"
}

func (v parsesAs) satisfies(ast j.Ast, expectation expectation) error {
	if basicLit, ok := ast.(*BasicLitExpr); ok {
		switch basicLit.Kind {
		case INT, FLOAT, IMAG:
		case STRING:
			// when it's a string,
			// we will need strconv.Unquote for things like
			// `"\u65e5本\U00008a9e"`
			// to be comparable to "日本語". We wouldn't in fact
			// necessarily need this conversion to be made,
			// but it helps make the tests more clear.
			// Also necessary for `parsesAsChar`.
			if s, err := strconv.Unquote(basicLit.Value); err == nil {
				if v.string == s {
					return nil // it's cool
				} else {
					return errors.New(fmt.Sprintf(
						"was expecting \"%s\", got \"%s\"", v.string, s))
				}
			} else {
				return errors.New(fmt.Sprintf(
					"%s did not successfully went thought strconv.Unquote: %s",
					basicLit.Value, err.Error()))
			}
		default:
			return errors.New(fmt.Sprintf(
				"expecting BasicLitExpr with Kind STRING, got %s",
				basicLit.Kind))
		}
	}
	// general case (binary expr etc)
	if ast == nil {
		return errors.New(fmt.Sprintf(
			"was expecting \"%s\", got %q", v.string, ast))
	} else if ast.String() != v.string {
		return errors.New(fmt.Sprintf(
			"was expecting \"%s\", got \"%s\"", v.string, ast.String()))
	}
	return nil
}

func (v parsesAsChar) satisfies(ast j.Ast, expectation expectation) error {
	if basicLit, ok := ast.(*BasicLitExpr); ok {
		if basicLit.Kind != CHAR {
			return errors.New(fmt.Sprintf(
				"expecting BasicLitExpr with Kind CHAR, got %s",
				basicLit.Kind))
		}
		if c, _, _, err := strconv.UnquoteChar(basicLit.Value, 0); err == nil {
			if v.rune == c {
				return nil // it's cool
			} else {
				return errors.New(fmt.Sprintf(
					"was expecting rune of hex \"%x\", got hex \"%x\"",
					v.rune, c))
			}
		} else {
			return errors.New(fmt.Sprintf(
				"%s did not successfully went through strconv.UnquoteChar: %s",
				basicLit.Value, err.Error()))
		}
	} else {
		return errors.New("expecting BasicLitExpr")
	}
}

func (v parsesAsNil) satisfies(ast j.Ast, expectation expectation) error {
	if ast != nil {
		return errors.New(fmt.Sprintf("was expecting nil, got %q", ast.String()))
	}
	return nil
}

func (v isBasicLit) satisfies(ast j.Ast, expectation expectation) error {
	if expr, ok := ast.(*BasicLitExpr); ok {
		if expr.Kind != v.kind {
			return errors.New(fmt.Sprintf(
				"was expecting Kind=%s for &BasicLitExpr, got %s",
				v.kind,
				expr.Kind,
			))
		}
	} else {
		return errors.New(fmt.Sprintf(
			"was expecting &BasicLitExpr (%v), got %s",
			v.kind,
			reflect.TypeOf(ast).String(),
		))
	}
	return nil
}

func (v isSelectorExpr) satisfies(ast j.Ast, expectation expectation) error {
	if _, ok := ast.(*SelectorExpr); !ok {
		return errors.New(fmt.Sprintf(
			"was expecting &SelectorExpr, got %s",
			reflect.TypeOf(ast).String(),
		))
	}
	return nil
}

func (v isNameExpr) satisfies(ast j.Ast, expectation expectation) error {
	if _, ok := ast.(*NameExpr); !ok {
		return errors.New(fmt.Sprintf(
			"was expecting &NameExpr, got %s",
			reflect.TypeOf(ast).String(),
		))
	}
	return nil
}

func (v isCallExpr) satisfies(ast j.Ast, expectation expectation) error {
	if _, ok := ast.(*CallExpr); !ok {
		return errors.New(fmt.Sprintf(
			"was expecting &CallExpr, got %s",
			reflect.TypeOf(ast).String(),
		))
	}
	return nil
}

func (v errorIs) satisfies(ast j.Ast, expectation expectation) error {
	if !j.IsParseError(ast) {
		return errors.New(fmt.Sprintf(
			"was expecting error %q, got result %q", v.string, ast.String()))
	}
	if v.string != "" && strings.TrimPrefix(ast.String(), "ERROR ") != v.string {
		return errors.New(fmt.Sprintf(
			"although we got a parse error as expected, were expecting %q"+
				", got %q", v.string, ast.String()))
	}
	return nil
}

func (v errorContains) satisfies(ast j.Ast, expectation expectation) error {
	if !j.IsParseError(ast) {
		return errors.New(fmt.Sprintf(
			"was expecting error %q, got %q", v.string, ast.String()))
	}
	if !strings.Contains(ast.String(), v.string) {
		return errors.New(fmt.Sprintf(
			"parse error as expected, but expecting error to contain \"%s\", "+
				"got \"%s\" instead", v.string, ast.String()))
	}
	return nil
}

func (noError) satisfies(ast j.Ast, expectation expectation) error {
	if j.IsParseError(ast) {
		return errors.New(fmt.Sprintf(
			"unexpected ParseError, was expecting %s", expectation.brief()))
	}
	return nil
}

func (t isType) satisfies(ast j.Ast, expectation expectation) error {
	theType := fmt.Sprintf("%T", ast)
	if !strings.HasSuffix(theType, t.string) {
		return errors.New(fmt.Sprintf("type should have been %s, not %s",
			t.string, theType))
	}
	return nil
}

// extract or eval left and right sides of BinaryExpr as int
// supposes limitations explained in evaluateBinaryExpr,
// and that BasicLitExpr of Kind INT be used as operands.
func vint(bx *BinaryExpr) (left int, right int) {
	if lx, ok := bx.Left.(*BinaryExpr); ok {
		left = evaluateBinaryExpr(lx).(int)
	} else {
		ble := bx.Left.(*BasicLitExpr)
		if ble.Kind != INT {
			panic("assert")
		}
		left, _ = strconv.Atoi(ble.Value)
	}
	if rx, ok := bx.Right.(*BinaryExpr); ok {
		right = evaluateBinaryExpr(rx).(int)
	} else {
		ble := bx.Right.(*BasicLitExpr)
		if ble.Kind != INT {
			panic("assert")
		}
		right, _ = strconv.Atoi(ble.Value)
	}
	return
}

// eval left and right sides of BinaryExpr as bool
// as in a && b and a || b. No extraction is done as in vint.
// This is not to be used outside of evaluateBinaryExpr.
func vbool(bx *BinaryExpr) (left bool, right bool) {
	if lx, ok := bx.Left.(*BinaryExpr); ok {
		left = evaluateBinaryExpr(lx).(bool)
	} else {
		panic("assert")
	}
	if rx, ok := bx.Right.(*BinaryExpr); ok {
		right = evaluateBinaryExpr(rx).(bool)
	} else {
		panic("assert")
	}
	return
}

// Used by binaryExprEvaluatesAsInt to write precedence tests.
// return: bool or int
// Severely limited binary expr evaluator!
//   - it can evaluate only * + - == != < <= > >= && ||
//   - only supports bool and int
//   - Left and Right MUST be either another limited BinaryExpr
//     or int or bool.
//   - There's no point at all using this other than from
//     very basic tests.
func evaluateBinaryExpr(bx *BinaryExpr) interface{} {
	switch bx.Op {
	case MUL:
		l, r := vint(bx)
		return l * r
	case ADD:
		l, r := vint(bx)
		return l + r
	case SUB:
		l, r := vint(bx)
		return l - r
	case EQL:
		l, r := vint(bx)
		return l == r
	case NEQ:
		l, r := vint(bx)
		return l != r
	case LSS:
		l, r := vint(bx)
		return l < r
	case LEQ:
		l, r := vint(bx)
		return l <= r
	case GEQ:
		l, r := vint(bx)
		return l >= r
	case GTR:
		l, r := vint(bx)
		return l > r
	case LAND:
		l, r := vbool(bx)
		return l && r
	case LOR:
		l, r := vbool(bx)
		return l || r
	default:
		panic(fmt.Sprintf("unsupported op in evaluateBinaryExpr: %d\n", bx.Op))
	}
}

func (e binaryExprEvaluatesAsInt) satisfies(ast j.Ast, expectation expectation) error {
	if x, ok := evaluateBinaryExpr(ast.(*BinaryExpr)).(int); !ok {
		return errors.New(fmt.Sprintf("was expecting binaryExpr %q to evaluate as int(%d), result is not int: %v",
			expectation.unparsedString,
			e.int,
			ok,
		))
	} else {
		if x == e.int {
			return nil
		} else {
			return errors.New(fmt.Sprintf("was expecting binaryExpr %q to evaluate as int(%d), got %d",
				expectation.unparsedString,
				e.int,
				x,
			))
		}
	}
}

func (e binaryExprEvaluatesAsBool) satisfies(ast j.Ast, expectation expectation) error {
	if x, ok := evaluateBinaryExpr(ast.(*BinaryExpr)).(bool); !ok {
		return errors.New(fmt.Sprintf("was expecting binaryExpr %q to evaluate as bool(%v), result is not bool: %v",
			expectation.unparsedString,
			e.bool,
			ok,
		))
	} else {
		if x == e.bool {
			return nil
		} else {
			return errors.New(fmt.Sprintf("was expecting binaryExpr %q to evaluate as bool(%v), got %v",
				expectation.unparsedString,
				e.bool,
				x,
			))
		}
	}
}

func (doom) satisfies(ast j.Ast, expectation expectation) error {
	fmt.Println("doom{} called")
	os.Exit(1)
	return nil
}

// }}}1

func TestJoeson(t *testing.T) {
	os.Setenv("TRACE", "stack")
	tests := []expectation{
		expect(``, parsesAsNil{}),
		// https://golang.google.com/ref/spec#Integer_literals
		expect(`2398`, parsesAs{"2398"}, isBasicLit{INT}),
		expect(`0`, parsesAs{"0"}, isBasicLit{INT}),
		expect(`0b0`, parsesAs{"0b0"}, isBasicLit{INT}),
		expect(`0B1`, parsesAs{"0b1"}, isBasicLit{INT}),
		expect(`0B_1`, parsesAs{"0b1"}, isBasicLit{INT}),
		expect(`0B_10`, parsesAs{"0b10"}, isBasicLit{INT}),
		expect(`0O777`, parsesAs{"0o777"}, isBasicLit{INT}),
		expect(`0o1`, parsesAs{"0o1"}, isBasicLit{INT}),
		expect(`0xBadFace`, parsesAs{"0xbadface"}, isBasicLit{INT}),
		expect(`0xBadAce`, parsesAs{"0xbadace"}, isBasicLit{INT}),
		expect(`0xdE_A_d_faC_e`, parsesAs{"0xdeadface"}, isBasicLit{INT}),
		expect(`0x_67_7a_2f_cc_40_c6`, parsesAs{"0x677a2fcc40c6"}, isBasicLit{INT}),
		expectErrorContains(`170141183460469231731687303715884105727`, "value out of range"),
		expectErrorContains(`170_141183_460469_231731_687303_715884_105727`, "value out of range"),
		expect(`_42`, parsesAs{"_42<VPUverse(0)>"}, isNameExpr{}), // an identifier, not an integer literal
		// expectError(`42_`, "invalid: _ must separate successive digits"),
		// 4__2        // invalid: only one _ at a time
		// 0_xBadFace  // invalid: _ must separate successive digits

		// https://golang.google.com/ref/spec#Floating-point_literals
		expect(`0.`, parsesAs{"0"}, isBasicLit{FLOAT}), // spec/FloatingPointsLiterals.txt
		expect(`72.40`, parsesAs{"72.4"}, isBasicLit{FLOAT}),
		expect(`072.40`, parsesAs{"72.4"}, isBasicLit{FLOAT}), // == 72.40
		expect(`2.71828`, parsesAs{"2.71828"}, isBasicLit{FLOAT}),
		expect(`1.e+0`, parsesAs{"1"}, isBasicLit{FLOAT}),
		expect(`6.67428e-11`, parsesAs{"6.67428e-11"}, isBasicLit{FLOAT}),
		expect(`1E6`, parsesAs{"1e+06"}, isBasicLit{FLOAT}),
		expect(`.25`, parsesAs{"0.25"}, isBasicLit{FLOAT}),
		expect(`.12345E+5`, parsesAs{"12345"}, isBasicLit{FLOAT}),
		expect(`1_5.`, parsesAs{"15"}, isBasicLit{FLOAT}),                 // == 15.0
		expect(`0.15e+0_2`, parsesAs{"15"}, isBasicLit{FLOAT}),            // == 15.0
		expect(`0x1p-2`, parsesAs{"0x1p-02"}, isBasicLit{FLOAT}),          // == 0.25
		expect(`0x2.p10`, parsesAs{"0x1p+11"}, isBasicLit{FLOAT}),         // == 2048.0
		expect(`0x1.Fp+0`, parsesAs{"0x1.fp+00"}, isBasicLit{FLOAT}),      // == 1.9375
		expect(`0X.8p-0`, parsesAs{"0x1p-01"}, isBasicLit{FLOAT}),         // == 0.5
		expect(`0X_1FFFP-16`, parsesAs{"0x1.fffp-04"}, isBasicLit{FLOAT}), // == 0.1249847412109375

		// https://golang.google.com/ref/spec#Imaginary_literals
		expect(`0i`, parsesAs{"0i"}, isBasicLit{IMAG}),
		expect(`0123i`, parsesAs{"0o123i"}, isBasicLit{IMAG}), // == 123i for backward-compatibility
		expect(`0.i`, parsesAs{"0i"}, isBasicLit{IMAG}),
		expect(`0o123i`, parsesAs{"0o123i"}, isBasicLit{IMAG}), // == 0o123 * 1i == 83i
		expect(`0xabci`, parsesAs{"0xabci"}, isBasicLit{IMAG}), // == 0xabc * 1i == 2748i
		expect(`2.71828i`, parsesAs{"2.71828i"}, isBasicLit{IMAG}),
		expect(`1.e+0i`, parsesAs{"1i"}, isBasicLit{IMAG}), // == (0+1i)
		expect(`6.67428e-11i`, parsesAs{"6.67428e-11i"}, isBasicLit{IMAG}),
		expect(`1E6i`, parsesAs{"1e+06i"}, isBasicLit{IMAG}), // == (0+1e+06i)
		expect(`.25i`, parsesAs{"0.25i"}, isBasicLit{IMAG}),
		expect(`.12345E+5i`, parsesAs{"12345i"}, isBasicLit{IMAG}),
		expect(`0x1p-2i`, parsesAs{"0x1p-02i"}, isBasicLit{IMAG}), // == 0x1p-2 * 1i == (0+0.25i)

		expect(`0x15e-2`, parsesAs{"0x15e - 2"}, isType{"BinaryExpr"}), // == 0x15e - 2 (integer subtraction)
		expect(`123 + 345`, parsesAs{"123 + 345"}, isType{"BinaryExpr"}),
		expect(`-1234`, parsesAs{"-1234"}, isType{"UnaryExpr"}),
		expect(`- 1234`, parsesAs{"-1234"}, isType{"UnaryExpr"}),
		expect(`+ 1234`, parsesAs{"+1234"}, isType{"UnaryExpr"}),
		expect(`!0`, parsesAs{"!0"}, isType{"UnaryExpr"}),
		expect(`^0`, parsesAs{"^0"}, isType{"UnaryExpr"}),
		expect(`<-s`, parsesAs{"<-s<VPUverse(0)>"}, isType{"UnaryExpr"}),
		expect(`*s`, parsesAs{"*(s<VPUverse(0)>)"}, isType{"StarExpr"}),
		expect(`&s`, parsesAs{"&(s<VPUverse(0)>)"}, isType{"RefExpr"}),
		expect(`-7 -2`, parsesAs{"-7 - 2"}, isType{"BinaryExpr"}),

		// {"0x.p1", "ERROR hexadecimal literal has no digits"},
		// expectError("0x.p1", "hexadecimal literal has no digits"),
		// 1p-2         // invalid: p exponent requires hexadecimal mantissa
		// 0x1.5e-2     // invalid: hexadecimal mantissa requires p exponent
		// 1_.5         // invalid: _ must separate successive digits
		// 1._5         // invalid: _ must separate successive digits
		// 1.5_e1       // invalid: _ must separate successive digits
		// 1.5e_1       // invalid: _ must separate successive digits
		// 1.5e1_       // invalid: _ must separate successive digits

		// https://golang.google.com/ref/spec#Rune_literals
		expect(`'\125'`, parsesAsChar{'U'}, isBasicLit{CHAR}),
		expectError(`'\0'`, "illegal: too few octal digits"),
		expectError(`'\12'`, "illegal: too few octal digits"),
		expectError(`'\400'`, "illegal: octal value over 255"),
		expectError(`'\1234'`, "illegal: too many octal digits"),
		expect(`'\x3d'`, parsesAsChar{'='}, isBasicLit{CHAR}),
		expect(`'\x3D'`, parsesAsChar{'='}, isBasicLit{CHAR}),
		expect(`'\a'`, parsesAsChar{'\a'}, isBasicLit{CHAR}), // alert or bell
		expect(`'\b'`, parsesAsChar{'\b'}, isBasicLit{CHAR}), // backspace
		expect(`'\f'`, parsesAsChar{'\f'}, isBasicLit{CHAR}), // form feed
		expect(`'\n'`, parsesAsChar{'\n'}, isBasicLit{CHAR}), // line feed or newline
		expect(`'\r'`, parsesAsChar{'\r'}, isBasicLit{CHAR}), // carriage return
		expect(`'\t'`, parsesAsChar{'\t'}, isBasicLit{CHAR}), // horizontal tab
		expect(`'\v'`, parsesAsChar{'\v'}, isBasicLit{CHAR}), // vertical tab
		expect(`'\\'`, parsesAsChar{'\\'}, isBasicLit{CHAR}), // backslash
		// expect(`'\''`, parsesAsChar{'\''}, isBasicLit{CHAR}),  // is this notation possible, HOW? See \u0027 below. single quote  (valid escape only within rune literals)
		expect(`'"'`, parsesAsChar{'"'}, isBasicLit{CHAR}),       // double quote  (valid escape only within string literals)
		expect(`'\u0007'`, parsesAsChar{'\a'}, isBasicLit{CHAR}), // alert or bell
		expect(`'\u0008'`, parsesAsChar{'\b'}, isBasicLit{CHAR}), // backspace
		expect(`'\u000C'`, parsesAsChar{'\f'}, isBasicLit{CHAR}), // form feed
		expect(`'\u000a'`, parsesAsChar{'\n'}, isBasicLit{CHAR}), // line feed or newline
		expect(`'\u000D'`, parsesAsChar{'\r'}, isBasicLit{CHAR}), // carriage return
		expect(`'\u0009'`, parsesAsChar{'\t'}, isBasicLit{CHAR}), // horizontal tab
		expect(`'\u000b'`, parsesAsChar{'\v'}, isBasicLit{CHAR}), // vertical tab
		expect(`'\u005c'`, parsesAsChar{'\\'}, isBasicLit{CHAR}), // backslash
		expect(`'\u0027'`, parsesAsChar{'\''}, isBasicLit{CHAR}), // single quote  (valid escape only within rune literals)
		expect(`'\u0022'`, parsesAsChar{'"'}, isBasicLit{CHAR}),  // double quote  (valid escape only within string literals)
		expect(`'\u13F8'`, parsesAsChar{'ᏸ'}, isBasicLit{CHAR}),
		expectError(`'\u13a'`, "little_u_value requires 4 hex"),
		expectError(`'\u1a248'`, "little_u_value requires 4 hex"),
		expect(`'\UFFeeFFee'`, isBasicLit{CHAR}),
		expectError(`'\UFFeeFFe'`, "big_u_value requires 8 hex"),
		expectError(`'\UFFeeFFeeA'`, "big_u_value requires 8 hex"),
		expect("'ä'", parsesAsChar{'ä'}, isBasicLit{CHAR}),
		expect("'本'", parsesAsChar{'本'}, isBasicLit{CHAR}),
		expect(`'\000'`, parsesAsChar{'\000'}, isBasicLit{CHAR}),
		expect(`'\007'`, parsesAsChar{'\007'}, isBasicLit{CHAR}),
		expect(`'''`, parsesAsChar{'\''}, isBasicLit{CHAR}), // rune literal containing single quote character
		// expectError("'aa'", "ERROR illegal: too many characters"),
		// expect("'\\k'",          "ERROR illegal: k is not recognized after a backslash",
		expectError(`'\xa'`, "illegal: too few hexadecimal digits"),
		// "'\\uDFFF'": "ERROR illegal: surrogate half", // TODO
		// "'\\U00110000'": "ERROR illegal: invalid Unicode code point", // TODO

		// tests from https://go.dev/ref/spec#String_literals
		expect("`abc`", parsesAs{"abc"}, isBasicLit{STRING}),
		expect("`"+`\n`+"`", parsesAs{"\\n"}, isBasicLit{STRING}), // original example is `\n<Actual CR>\n` // same as "\\n\n\\n". But's a bit hard to reproduce...
		expect(`"abc"`, parsesAs{"abc"}, isBasicLit{STRING}),
		expect(`"\\\""`, parsesAs{`"`}, isBasicLit{STRING}), // same as `"`
		expect(`"Hello, world!\\n"`, parsesAs{"Hello, world!\n"}, isBasicLit{STRING}),
		expect(`"\\xff\\u00FF"`, isBasicLit{STRING}),
		expect(`"日本語"`, parsesAs{"日本語"}, isBasicLit{STRING}), // this and the 3 next lines all represent the same string ("japanese")
		expect(`"\\u65e5本\\U00008a9e"`, parsesAs{"日本語"}, isBasicLit{STRING}),
		expect(`"\\U000065e5\\U0000672c\\U00008a9e"`, parsesAs{"日本語"}, isBasicLit{STRING}),             // the explicit Unicode code points
		expect(`"\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e"`, parsesAs{"日本語"}, isBasicLit{STRING}), // the explicit UTF-8 bytes

		// tests from https://golang.google.com/ref/spec#Identifiers
		expect(`a`, parsesAs{"a<VPUverse(0)>"}, isNameExpr{}),
		expect(`_x9`, parsesAs{"_x9<VPUverse(0)>"}, isNameExpr{}),
		expect(`ThisVariableIsExported`, parsesAs{"ThisVariableIsExported<VPUverse(0)>"}, isNameExpr{}),
		expect(`αβ`, parsesAs{"αβ<VPUverse(0)>"}, isNameExpr{}),
		expect(`nil`, parsesAs{"nil<VPUverse(0)>"}, isNameExpr{}),

		// tests from https://dev.to/flopp/golang-identifiers-vs-unicode-1fe7
		expect(`abc_123`, parsesAs{"abc_123<VPUverse(0)>"}, isNameExpr{}),
		expect(`_myidentifier`, parsesAs{"_myidentifier<VPUverse(0)>"}, isNameExpr{}),
		expect(`Σ`, parsesAs{"Σ<VPUverse(0)>"}, isNameExpr{}),             // (U+03A3 GREEK CAPITAL LETTER SIGMA),
		expect(`㭪`, parsesAs{"㭪<VPUverse(0)>"}, isNameExpr{}),             // (some CJK character from the Lo category),
		expect(`x٣३߃૩୩3`, parsesAs{"x٣३߃૩୩3<VPUverse(0)>"}, isNameExpr{}), // (x + decimal digits 3 from various scripts),
		expectError(`😀`, ""),  // (not a letter, but So / Symbol, other)
		expectError(`⽔`, ""),  // (not a letter, but So / Symbol, other)
		expectError(`x🌞`, ""), // (starts with a letter, but contains non-letter/digit characters)

		// expect(`package math`, parsesAs{"package math"}), // unsupported by X() AFAIK -> don't bother Stmt and scopes for now
		expect(`math.Sin`, parsesAs{"math<VPUverse(0)>.Sin"}, isSelectorExpr{}), // denotes the Sin function in package math

		expect(`math.Atan2(x, y)`, parsesAs{"math<VPUverse(0)>.Atan2(x<VPUverse(0)>, y<VPUverse(0)>)"}, isCallExpr{}), // function call

		expect(`h(x+y)`, parsesAs{"h<VPUverse(0)>(x<VPUverse(0)> + y<VPUverse(0)>)"}, isCallExpr{}),
		expect(`f.Close()`, parsesAs{"f<VPUverse(0)>.Close()"}, isCallExpr{}),
		expect(`<-ch`, parsesAs{"<-ch<VPUverse(0)>"}),
		expect(`(<-ch)`, parsesAs{"<-ch<VPUverse(0)>"}),
		expect(`len("foo")`, parsesAs{`len<VPUverse(0)>("foo")`}, isCallExpr{}), // marked "illegal if len is the built-in function" in gospec, I don't get why?

		// https://golang.google.com/ref/spec#Primary_expressions
		expect(`x`, parsesAs{"x<VPUverse(0)>"}, isNameExpr{}),
		expect(`2`, parsesAs{"2"}, isBasicLit{INT}),
		expect(`s + ".txt"`, parsesAs{`s<VPUverse(0)> + ".txt"`}, isType{"BinaryExpr"}),
		expect(`f(3.1415, true)`, parsesAs{`f<VPUverse(0)>(3.1415, true<VPUverse(0)>)`}, isCallExpr{}),
		// expect(`Point{1, 2}`), not supported yet
		expect(`m["foo"]`, parsesAs{`m<VPUverse(0)>["foo"]`}, isType{"IndexExpr"}),
		expect(`m[361]`, parsesAs{`m<VPUverse(0)>[361]`}, isType{"IndexExpr"}),
		expect(`s[i : j + 1]`, parsesAs{`s<VPUverse(0)>[i<VPUverse(0)>:j<VPUverse(0)> + 1]`}, isType{"SliceExpr"}),
		expect(`s[1:2:3]`, parsesAs{`s<VPUverse(0)>[1:2:3]`}, isType{"SliceExpr"}),
		expect(`s[:2:3]`, parsesAs{`s<VPUverse(0)>[:2:3]`}, isType{"SliceExpr"}),
		expect(`s[1:2]`, parsesAs{`s<VPUverse(0)>[1:2]`}, isType{"SliceExpr"}),
		expect(`s[:2]`, parsesAs{`s<VPUverse(0)>[:2]`}, isType{"SliceExpr"}),
		expect(`s[1:]`, parsesAs{`s<VPUverse(0)>[1:]`}, isType{"SliceExpr"}),
		expect(`s[: i : (314*10)-6]`, parsesAs{`s<VPUverse(0)>[:i<VPUverse(0)>:314 * 10 - 6]`}, isType{"SliceExpr"}),
		expect(`obj.color`, parsesAs{`obj<VPUverse(0)>.color`}, isType{"SelectorExpr"}),
		expect(`f.p[i].x()`, parsesAs{`f<VPUverse(0)>.p[i<VPUverse(0)>].x()`}, isCallExpr{}),

		// TypeAssertion using various types notation
		expect(`x.(int)`, parsesAs{`x<VPUverse(0)>.((const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`x.(float32)`, parsesAs{`x<VPUverse(0)>.((const-type float32))`}, isType{"TypeAssertExpr"}),
		// TODO support non primitive types as below
		// expect(`x.(*T)`, parsesAs{`x<VPUverse(0)>.([3](const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`x.([]int)`, parsesAs{`x<VPUverse(0)>.([](const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`x.([3]int)`, parsesAs{`x<VPUverse(0)>.([3](const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`x.(*int)`, parsesAs{`x<VPUverse(0)>.(*((const-type int)))`}, isType{"TypeAssertExpr"}),
		expect(`x.(map[string]bool)`, parsesAs{`x<VPUverse(0)>.(map[(const-type string)] (const-type bool))`}, isType{"TypeAssertExpr"}),
		expect(`x.(chan int)`, parsesAs{`x<VPUverse(0)>.(chan (const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`x.(chan<- float64)`, parsesAs{`x<VPUverse(0)>.(<-chan (const-type float64))`}, isType{"TypeAssertExpr"}),
		expect(`x.(<-chan string)`, parsesAs{`x<VPUverse(0)>.(chan<- (const-type string))`}, isType{"TypeAssertExpr"}),
		expect(`x.(<-chan []int)`, parsesAs{`x<VPUverse(0)>.(chan<- [](const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`x.(<-chan chan<- chan []<-chan int)`, parsesAs{`x<VPUverse(0)>.(chan<- <-chan chan []chan<- (const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a int))`, parsesAs{`f<VPUverse(0)>.(func(a (const-type int)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a int, b int))`, parsesAs{`f<VPUverse(0)>.(func(a (const-type int), b (const-type int)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a int, b int) int)`, parsesAs{`f<VPUverse(0)>.(func(a (const-type int), b (const-type int))  (const-type int))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a int, b int) (ok bool, sum int))`, parsesAs{`f<VPUverse(0)>.(func(a (const-type int), b (const-type int)) ok (const-type bool), sum (const-type int))`}, isType{"TypeAssertExpr"}),
		expectError(`f.(func(a int, b int) (ok ...bool)`, "function results can not be variadic"),
		expectError(`f.(func(a int, b int) (ok bool, sum ...int))`, "function results can not be variadic"),
		expect(`f.(func(a, b int))`, parsesAs{`f<VPUverse(0)>.(func(a (const-type int), b (const-type int)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a int, b ...int))`, parsesAs{`f<VPUverse(0)>.(func(a (const-type int), b ...(const-type int)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a, b int, c ...int))`, parsesAs{`f<VPUverse(0)>.(func(a (const-type int), b (const-type int), c ...(const-type int)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a func(s string) bool))`, parsesAs{`f<VPUverse(0)>.(func(a func(s (const-type string))  (const-type bool)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a func(s ...string) bool))`, parsesAs{`f<VPUverse(0)>.(func(a func(s ...(const-type string))  (const-type bool)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a func(s ...string) (a int, s string)))`, parsesAs{`f<VPUverse(0)>.(func(a func(s ...(const-type string)) a (const-type int), s (const-type string)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a func(s ...string) (a int, o <-chan int)))`, parsesAs{`f<VPUverse(0)>.(func(a func(s ...(const-type string)) a (const-type int), o chan<- (const-type int)))`}, isType{"TypeAssertExpr"}),
		expect(`f.(func(a func(s ...string) (a int, o <-chan func(s string) bool)))`, parsesAs{`f<VPUverse(0)>.(func(a func(s ...(const-type string)) a (const-type int), o chan<- func(s (const-type string))  (const-type bool)))`}, isType{"TypeAssertExpr"}),
		expect(`x.(struct { x, y float32 })`, parsesAs{`x<VPUverse(0)>.(struct { x (const-type float32) <nil>, y (const-type float32) <nil> })`}, isType{"TypeAssertExpr"}), // legit (albeit useless? can not think of any application) type assertion e.g. when `var x interface{}`.

		// https://go.dev/ref/spec#CompositeLit
		expect(`[]int{2,  3,5,7, 9, 2147483647}`, parsesAs{`[](const-type int){2, 3, 5, 7, 9, 2147483647}`}, isType{"CompositeLitExpr"}),
		expect(`[128]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true}`,
			parsesAs{`[128](const-type bool){a: true<VPUverse(0)>, e: true<VPUverse(0)>, i: true<VPUverse(0)>, o: true<VPUverse(0)>, u: true<VPUverse(0)>, y: true<VPUverse(0)>}`}, isType{"CompositeLitExpr"}),
		expect(`[10]float32{-1, 0, 0, 0, -0.1, -0.1, 0, 0, 0, -1}`,
			parsesAs{`[10](const-type float32){-1, 0, 0, 0, -0.1, -0.1, 0, 0, 0, -1}`}, isType{"CompositeLitExpr"}),
		expect(`[10]float32{-1, 4: -0.1, -0.1, 9: -1}`, parsesAs{`[10](const-type float32){-1, 4: -0.1, -0.1, 9: -1}`}, isType{"CompositeLitExpr"}),
		expect(`map[string]float32{"C0": 16.35, "D0": 18.35, "E0": 20.60, "F0": 21.83, "G0": 24.50, "A0": 27.50, "B0": 30.87, }`,
			parsesAs{`map[(const-type string)] (const-type float32){"C0": 16.35, "D0": 18.35, "E0": 20.6, "F0": 21.83, "G0": 24.5, "A0": 27.5, "B0": 30.87}`}, isType{"CompositeLitExpr"}),

		// no func until we parse statements
		// expect(`func(x, y int) int { x + y }`, parsesAs{`[]`}, isType{"FunctionLit"}), // FIXME using FunctionLit and SimpleStmt we can't express anything interesting yet

		// test precedence (from X())
		//	5             *  /  %  <<  >>  &  &^
		//	4             +  -  |  ^
		//	3             ==  !=  <  <=  >  >=
		//	2             &&
		//	1             ||
		expect(`a == d`, parsesAs{`a<VPUverse(0)> == d<VPUverse(0)>`}, isType{"BinaryExpr"}),
		expect(`1 + 7*2`, isType{"BinaryExpr"}, binaryExprEvaluatesAsInt{15}),
		// expect(`7*2 + 1`, isType{"BinaryExpr"}, binaryExprEvaluatesAsInt{15}),
		// expect(`7 + 1*2 == 7 + 1*2`, isType{"BinaryExpr"}, binaryExprEvaluatesAsBool{true}),
		// expect(`a * b + c == d - e / f && 4 >= 1+1 || 7/1 == 7`, parsesAs{`x`}),

		// "Binary operators of the same precedence associate from left to right."
	}
	for _, expectation := range tests {
		testExpectation(t, expectation)
	}
}

// vim: fdm=marker fdl=0