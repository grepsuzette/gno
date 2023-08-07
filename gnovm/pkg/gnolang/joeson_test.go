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

const (
	asExpr parseAsKind = 0 // expressions
	asFile             = 1 // statements wrapped in implicit FileBlock
)

type (
	parseAsKind int
	predicate   interface {
		satisfies(j.Ast, expectation) error
	}
	expectation struct {
		parseAsKind
		unparsedString string
		predicates     []predicate
	}
	parsesAs      struct{ string } // strict string equality
	parsesAsChar  struct{ rune }   // strict string equality
	isBasicLit    struct{ kind Word }
	errorIs       struct{ string }
	errorContains struct{ string }
	noError       struct{}
	isType        struct{ string }
	doom          struct{}
)

var (
	_ predicate = parsesAs{}
	_ predicate = parsesAsChar{}
	_ predicate = isBasicLit{}
	_ predicate = errorIs{}
	_ predicate = errorContains{}
	_ predicate = noError{}
	_ predicate = isType{}
	_ predicate = doom{} // doom = stop tests (a, b, doom, c, d will stop after b, for quick debug)
)

func testExpectation(t *testing.T, expectation expectation) {
	t.Helper()
	ast := parseX(expectation.unparsedString)
	allOk := true
	for _, predicate := range expectation.predicates {
		if err := predicate.satisfies(ast, expectation); err != nil {
			allOk = false
			t.Fatalf(
				"%s parsed as %s "+j.BoldRed("ERR")+" %s\n",
				helpers.Escape(expectation.unparsedString),
				ast.String(),
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
			j.Yellow(helpers.Escape(ast.String())),
			"", // b.String(),
		)
	}
}

func doesntMatchError(expect, got string) bool {
	return !strings.HasPrefix(got, expect[len("ERROR"):])
}

// expect() parses `unparsedString`, requiring it is not parsed as a ParseError and all other
// conditions that are indicated as `...predicate`.  parseAsKind allow to parse it in different
// contexts.
func expect(parseAsKind parseAsKind, unparsedString string, preds ...predicate) expectation {
	// insert noError{} at the beginning
	a := make([]predicate, len(preds)+1)
	copy(a[1:], preds)
	a[0] = noError{}
	return expectation{parseAsKind, unparsedString, a}
}

func expectError(parseAsKind parseAsKind, unparsedString string, expectedError string) expectation {
	return expectation{
		parseAsKind,
		unparsedString,
		[]predicate{errorIs{expectedError}},
	}
}

func expectErrorContains(parseAsKind parseAsKind, unparsedString string,
	expectedError string,
) expectation {
	return expectation{
		parseAsKind,
		unparsedString,
		[]predicate{errorContains{expectedError}},
	}
}

// this is just a way to stop the program at a certain place
// from the array of tests
func expectDoom() expectation {
	return expectation{asExpr, "", []predicate{doom{}}}
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
	if ast.String() != v.string {
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

func (doom) satisfies(ast j.Ast, expectation expectation) error {
	fmt.Println("doom{} called")
	os.Exit(1)
	return nil
}

// }}}1

// TODO remaining some more custom errors to do, the tests to enable are commented below
func TestJoesonExpressions(t *testing.T) { // {{{1
	// os.Setenv("TRACE", "grammar,stack")
	tests := []expectation{
		expect(asExpr, `2398`, parsesAs{"2398"}, isBasicLit{INT}),
		expect(asExpr, `0`, parsesAs{"0"}, isBasicLit{INT}),
		expect(asExpr, `0b0`, parsesAs{"0b0"}, isBasicLit{INT}),
		expect(asExpr, `0B1`, parsesAs{"0b1"}, isBasicLit{INT}),
		expect(asExpr, `0B_1`, parsesAs{"0b1"}, isBasicLit{INT}),
		expect(asExpr, `0B_10`, parsesAs{"0b10"}, isBasicLit{INT}),
		expect(asExpr, `0O777`, parsesAs{"0o777"}, isBasicLit{INT}),
		expect(asExpr, `0o1`, parsesAs{"0o1"}, isBasicLit{INT}),
		expect(asExpr, `0xBadFace`, parsesAs{"0xbadface"}, isBasicLit{INT}),
		expect(asExpr, `0xBadAce`, parsesAs{"0xbadace"}, isBasicLit{INT}),
		expect(asExpr, `0xdE_A_d_faC_e`, parsesAs{"0xdeadface"}, isBasicLit{INT}),
		expect(asExpr, `0x_67_7a_2f_cc_40_c6`, parsesAs{"0x677a2fcc40c6"}, isBasicLit{INT}),
		expectErrorContains(asExpr, `170141183460469231731687303715884105727`, "value out of range"),
		expectErrorContains(asExpr, `170_141183_460469_231731_687303_715884_105727`, "value out of range"),
		// _42         // an identifier, not an integer literal
		// 42_         // invalid: _ must separate successive digits
		// 4__2        // invalid: only one _ at a time
		// 0_xBadFace  // invalid: _ must separate successive digits

		expect(asExpr, `0.`, parsesAs{"0"}, isBasicLit{FLOAT}), // spec/FloatingPointsLiterals.txt
		expect(asExpr, `72.40`, parsesAs{"72.4"}, isBasicLit{FLOAT}),
		expect(asExpr, `072.40`, parsesAs{"72.4"}, isBasicLit{FLOAT}), // == 72.40
		expect(asExpr, `2.71828`, parsesAs{"2.71828"}, isBasicLit{FLOAT}),
		expect(asExpr, `1.e+0`, parsesAs{"1"}, isBasicLit{FLOAT}),
		expect(asExpr, `6.67428e-11`, parsesAs{"6.67428e-11"}, isBasicLit{FLOAT}),
		expect(asExpr, `1E6`, parsesAs{"1e+06"}, isBasicLit{FLOAT}),
		expect(asExpr, `.25`, parsesAs{"0.25"}, isBasicLit{FLOAT}),
		expect(asExpr, `.12345E+5`, parsesAs{"12345"}, isBasicLit{FLOAT}),
		expect(asExpr, `1_5.`, parsesAs{"15"}, isBasicLit{FLOAT}),                 // == 15.0
		expect(asExpr, `0.15e+0_2`, parsesAs{"15"}, isBasicLit{FLOAT}),            // == 15.0
		expect(asExpr, `0x1p-2`, parsesAs{"0x1p-02"}, isBasicLit{FLOAT}),          // == 0.25
		expect(asExpr, `0x2.p10`, parsesAs{"0x1p+11"}, isBasicLit{FLOAT}),         // == 2048.0
		expect(asExpr, `0x1.Fp+0`, parsesAs{"0x1.fp+00"}, isBasicLit{FLOAT}),      // == 1.9375
		expect(asExpr, `0X.8p-0`, parsesAs{"0x1p-01"}, isBasicLit{FLOAT}),         // == 0.5
		expect(asExpr, `0X_1FFFP-16`, parsesAs{"0x1.fffp-04"}, isBasicLit{FLOAT}), // == 0.1249847412109375

		expect(asExpr, `0i`, parsesAs{"0i"}, isBasicLit{IMAG}),
		expect(asExpr, `0123i`, parsesAs{"0o123i"}, isBasicLit{IMAG}), // == 123i for backward-compatibility
		expect(asExpr, `0.i`, parsesAs{"0i"}, isBasicLit{IMAG}),
		expect(asExpr, `0o123i`, parsesAs{"0o123i"}, isBasicLit{IMAG}), // == 0o123 * 1i == 83i
		expect(asExpr, `0xabci`, parsesAs{"0xabci"}, isBasicLit{IMAG}), // == 0xabc * 1i == 2748i
		expect(asExpr, `2.71828i`, parsesAs{"2.71828i"}, isBasicLit{IMAG}),
		expect(asExpr, `1.e+0i`, parsesAs{"1i"}, isBasicLit{IMAG}), // == (0+1i)
		expect(asExpr, `6.67428e-11i`, parsesAs{"6.67428e-11i"}, isBasicLit{IMAG}),
		expect(asExpr, `1E6i`, parsesAs{"1e+06i"}, isBasicLit{IMAG}), // == (0+1e+06i)
		expect(asExpr, `.25i`, parsesAs{"0.25i"}, isBasicLit{IMAG}),
		expect(asExpr, `.12345E+5i`, parsesAs{"12345i"}, isBasicLit{IMAG}),
		expect(asExpr, `0x1p-2i`, parsesAs{"0x1p-02i"}, isBasicLit{IMAG}), // == 0x1p-2 * 1i == (0+0.25i)

		expect(asExpr, `0x15e-2`, parsesAs{"0x15e - 2"}, isType{"BinaryExpr"}), // == 0x15e - 2 (integer subtraction)
		expect(asExpr, `123 + 345`, parsesAs{"123 + 345"}, isType{"BinaryExpr"}),
		expect(asExpr, `-1234`, parsesAs{"-1234"}, isType{"UnaryExpr"}),
		expect(asExpr, `- 1234`, parsesAs{"-1234"}, isType{"UnaryExpr"}),
		expect(asExpr, `+ 1234`, parsesAs{"+1234"}, isType{"UnaryExpr"}),
		expect(asExpr, `!0`, parsesAs{"!0"}, isType{"UnaryExpr"}),
		expect(asExpr, `^0`, parsesAs{"^0"}, isType{"UnaryExpr"}),
		expect(asExpr, `-7 -2`, parsesAs{"-7 - 2"}, isType{"BinaryExpr"}),

		// {"0x.p1", "ERROR hexadecimal literal has no digits"},
		// expectError("0x.p1", "hexadecimal literal has no digits"),
		// 1p-2         // invalid: p exponent requires hexadecimal mantissa
		// 0x1.5e-2     // invalid: hexadecimal mantissa requires p exponent
		// 1_.5         // invalid: _ must separate successive digits
		// 1._5         // invalid: _ must separate successive digits
		// 1.5_e1       // invalid: _ must separate successive digits
		// 1.5e_1       // invalid: _ must separate successive digits
		// 1.5e1_       // invalid: _ must separate successive digits

		expect(asExpr, `'\125'`, parsesAsChar{'U'}, isBasicLit{CHAR}),
		expectError(asExpr, `'\0'`, "illegal: too few octal digits"),
		expectError(asExpr, `'\12'`, "illegal: too few octal digits"),
		expectError(asExpr, `'\400'`, "illegal: octal value over 255"),
		expectError(asExpr, `'\1234'`, "illegal: too many octal digits"),
		expect(asExpr, `'\x3d'`, parsesAsChar{'='}, isBasicLit{CHAR}),
		expect(asExpr, `'\x3D'`, parsesAsChar{'='}, isBasicLit{CHAR}),
		expect(asExpr, `'\a'`, parsesAsChar{'\a'}, isBasicLit{CHAR}),
		expect(asExpr, `'\b'`, parsesAsChar{'\b'}, isBasicLit{CHAR}),
		expect(asExpr, `'\f'`, parsesAsChar{'\f'}, isBasicLit{CHAR}),
		expect(asExpr, `'\n'`, parsesAsChar{'\n'}, isBasicLit{CHAR}),
		expect(asExpr, `'\r'`, parsesAsChar{'\r'}, isBasicLit{CHAR}),
		expect(asExpr, `'\t'`, parsesAsChar{'\t'}, isBasicLit{CHAR}),
		expect(asExpr, `'\v'`, parsesAsChar{'\v'}, isBasicLit{CHAR}),
		expect(asExpr, `'\u13F8'`, parsesAsChar{'ᏸ'}, isBasicLit{CHAR}),
		expectError(asExpr, `'\u13a'`, "little_u_value requires 4 hex"),
		expectError(asExpr, `'\u1a248'`, "little_u_value requires 4 hex"),
		expect(asExpr, `'\UFFeeFFee'`, isBasicLit{CHAR}),
		expectError(asExpr, `'\UFFeeFFe'`, "big_u_value requires 8 hex"),
		expectError(asExpr, `'\UFFeeFFeeA'`, "big_u_value requires 8 hex"),
		expect(asExpr, "'ä'", parsesAsChar{'ä'}, isBasicLit{CHAR}),
		expect(asExpr, "'本'", parsesAsChar{'本'}, isBasicLit{CHAR}),
		expect(asExpr, `'\000'`, parsesAsChar{'\000'}, isBasicLit{CHAR}),
		expect(asExpr, `'\007'`, parsesAsChar{'\007'}, isBasicLit{CHAR}),
		expect(asExpr, `'''`, parsesAsChar{'\''}, isBasicLit{CHAR}), // rune literal containing single quote character
		// expectError("'aa'", "ERROR illegal: too many characters"),
		// expect(asExpression, "'\\k'",          "ERROR illegal: k is not recognized after a backslash",
		expectError(asExpr, `'\xa'`, "illegal: too few hexadecimal digits"),
		// "'\\uDFFF'": "ERROR illegal: surrogate half", // TODO
		// "'\\U00110000'": "ERROR illegal: invalid Unicode code point", // TODO

		expect(asExpr, "`abc`", parsesAs{"abc"}, isBasicLit{STRING}),      // https://go.dev/ref/spec#String_literals
		expect(asExpr, "`"+`\n`+"`", parsesAs{"\\n"}, isBasicLit{STRING}), // original example is `\n<Actual CR>\n` // same as "\\n\n\\n". But's a bit hard to reproduce...
		expect(asExpr, `"abc"`, parsesAs{"abc"}, isBasicLit{STRING}),
		expect(asExpr, `"\\\""`, parsesAs{`"`}, isBasicLit{STRING}), // same as `"`
		expect(asExpr, `"Hello, world!\\n"`, parsesAs{"Hello, world!\n"}, isBasicLit{STRING}),
		expect(asExpr, `"\\xff\\u00FF"`, isBasicLit{STRING}),
		// these 4 examples all represent the same string ("japanese language" written in japanese)
		expect(asExpr, `"日本語"`, parsesAs{"日本語"}, isBasicLit{STRING}),
		expect(asExpr, `"\\u65e5本\\U00008a9e"`, parsesAs{"日本語"}, isBasicLit{STRING}),
		expect(asExpr, `"\\U000065e5\\U0000672c\\U00008a9e"`, parsesAs{"日本語"}, isBasicLit{STRING}),             // the explicit Unicode code points
		expect(asExpr, `"\\xe6\\x97\\xa5\\xe6\\x9c\\xac\\xe8\\xaa\\x9e"`, parsesAs{"日本語"}, isBasicLit{STRING}), // the explicit UTF-8 bytes
	}
	for _, expectation := range tests {
		testExpectation(t, expectation)
	}
} // }}}1

// statements inside an implicit blockfile
func TestJoesonBlockFile(t *testing.T) {
	os.Setenv("TRACE", "stack")
	tests := []expectation{
		expect(asFile, `package foo`, parsesAs{""}, isBasicLit{INT}),
	}
	for _, expectation := range tests {
		testExpectation(t, expectation)
	}
}

// vim: fdm=marker fdl=0