package gnolang

import (
	"strings"

	j "github.com/grepsuzette/joeson"
)

// A Golang PEG grammar.
//   - This grammar is written against the *formal go syntax* with semicolons ";"
//     as terminators, as is produced for example by "go/scanner".
//   - The following rules are named after https://go.dev/ref/spec labels (such
//     as "bx:") are used by rules callbacks.
//   - Rules that don't have a name immediately listed in spec/spec.html should
//     be prefixed by an underscore `_` (e.g. _T, the ";" terminator). Note
//     those underlined rules aren't captured (see joeson/parser_ref.go).
//
// note (?...) are lookaheads, they can optimize away certain branches but consume nothing
var (
	gm       *j.Grammar
	gnoRules = rules(
		o(named("Input", "SimpleStmt _semicolon?")),
		o(named("Block", rules(
			o("'{' Statement*_semicolon '}'"),
			i(named("Statement", "SimpleStmt")),
			i(named("SimpleStmt", "ExpressionStmt"), fSimpleStmt),
			i(named("ExpressionStmt", "Expression")),
			i(named("Expression", rules(
				o("bx:(Expression binary_op Expression) | ux:UnaryExpr"),
				o(named("UnaryExpr", "PrimaryExpr | ux:(unary_op UnaryExpr)"), fUnary),
				o(named("unary_op", revQuote("+ - ! ^ * & <-"))),
				o(named("binary_op", "mul_op | add_op | rel_op | '&&' | '||'")),
				o(named("mul_op", revQuote("* / % << >> & &^"))),
				o(named("add_op", revQuote("+ - | ^"))),
				o(named("rel_op", revQuote("== != < <= > >="))),
				// o(named("PrimaryExpr", "Operand | Conversion | MethodExpr | PrimaryExpr ( Selector | Index | Slice | TypeAssertion | Arguments )")),
				o(named("PrimaryExpr", rules(
					o("p:PrimaryExpr a:Arguments", fPrimaryExprArguments), // e.g. `math.Atan2(x, y)`
					o("p:PrimaryExpr i:Index", fPrimaryExprIndex),         // e.g. `something[1]`
					o("p:PrimaryExpr s:Slice", fPrimaryExprSlice),         // e.g. `a[23 : 87]`
					o("p:PrimaryExpr s:Selector", fPrimaryExprSelector),   // e.g. `x.f` for a PrimaryExpr x that is not a package name
					o("Operand"),
					i(named("Operand", rules(
						// o("'(' Expression ')' | OperandName TypeArgs? | Literal"), // TODO this is the original
						o("Literal | OperandName | '(' Expression ')'"),
						i(named("Literal", rules(
							o("BasicLit | FunctionLit | CompositeLit"),
							i(named("BasicLit", rules(
								o("(?'\\'') rune_lit | (?[\"`]) string_lit | (?[0-9.]) imaginary_lit | (?[0-9.]) float_lit | (?[0-9]) int_lit"), // (?xx) is lookahead, quickly pruning away
								i(named("rune_lit", `'\'' ( byte_value | unicode_value | [^\n] ) '\''`), f_rune_lit),
								i(named("string_lit", rules(
									o(named("raw_string_lit", "'`' [^`]* '`'"), fraw_string_lit),
									o(named("interpreted_string_lit", `'"' (!'\"' ('\\' [\s\S] | unicode_value | byte_value))* '"'`), finterpreted_string_lit),
								))),
								i(named("int_lit", rules(
									o(named("binary_lit", "('0b'|'0B') '_'? binary_digits"), ffInt(2)),
									o(named("hex_lit", "('0x'|'0X') '_'? hex_digits"), ffInt(16)),
									o(named("octal_lit", "[0] [oO]? '_'? octal_digit octal_digits?"), ffInt(8)),
									o(named("decimal_lit", "[0] | [1-9] ( '_'? decimal_digits)?"), ffInt(10)),
								))),
								i(named("float_lit", rules(
									o("decimal_float_lit | hex_float_lit"),
									i(named("decimal_float_lit",
										"DOT decimal_digits decimal_exponent? | "+
											"decimal_digits DOT decimal_digits? decimal_exponent? | "+
											"decimal_digits decimal_exponent"), ffFloatFormat("%g")),
									i(named("hex_float_lit", "[0] [xX] hex_mantissa hex_exponent"), ffFloatFormat("%x")),
									i(named("decimal_exponent", "[eE] [+-]? decimal_digits")),
									i(named("hex_mantissa", "'_'? hex_digits DOT hex_digits? |"+
										"'_'? hex_digits | DOT hex_digits",
									)),
									i(named("hex_exponent", "[pP] [+-]? decimal_digits")),
								))),
								i(named("imaginary_lit", "(float_lit | int_lit | decimal_digits ) [i]"), fImaginary),
								// avoid regexes with PEG in general, regexes are greedy and this can
								// create ambiguity and buggy grammars. As a special case, character classes are OK.
								// Regexes can be used to optimize but again avoid them unless
								// you know what you're doing.
								i(named("decimal_digits", "decimal_digit ( '_'? decimal_digit )*")),
								i(named("binary_digits", "binary_digit ( '_'? binary_digit )*")),
								i(named("octal_digits", "octal_digit ( '_'? octal_digit )*")),
								i(named("hex_digits", "hex_digit ( '_'? hex_digit )*")),
								i(named("decimal_digit", "[0-9]")),
								i(named("binary_digit", "[01]")),
								i(named("octal_digit", "[0-7]")),
								i(named("hex_digit", "[0-9a-fA-F]")),
								i(named("byte_value", rules(
									o(`  (?'\\' octal_digit) (octal_byte_value_err1 | octal_byte_value | octal_byte_value_err2) |`+
										`(?'\\x'           ) (hex_byte_value_err1 | hex_byte_value | hex_byte_value_err2)`),
									i(named("octal_byte_value_err1", `a:'\\' (?octal_digit{4,})`), ffPanic("illegal: too many octal digits")),
									i(named("octal_byte_value", `a:'\\' b:octal_digit{3,3}`), foctal_byte_value), // passthru unless "illegal: octal value over 255"
									i(named("octal_byte_value_err2", `a:'\\' (?octal_digit{1,})`), ffPanic("illegal: too few octal digits")),
									i(named("hex_byte_value_err1", `a:'\\x' b:hex_digit{3,}`), ffPanic("illegal: too many hexadecimal digits")),
									i(named("hex_byte_value", `a:'\\x' b:hex_digit{2,2}`)),
									i(named("hex_byte_value_err2", `a:'\\x' b:hex_digit{0,1}`), ffPanic("illegal: too few hexadecimal digits")),
								))),
								i(named("unicode_value", rules(
									o("escaped_char | little_u_value | big_u_value | unicode_char | _error_unicode_char_toomany"),
									i(named("escaped_char", `esc:'\\' char:[abfnrtv\\\'"]`)),
									i(named("little_u_value", `a:'\\u' b:hex_digit*`), ff_u_value("little_u_value", 4)), // 4 hex_digit or error
									i(named("big_u_value", `a:'\\U' b:hex_digit*`), ff_u_value("big_u_value", 8)),       // 8 hex digit or error
									i(named("_error_unicode_char_toomany", "[^\\x{0a}]{2,}"), func(it j.Ast, ctx *j.ParseContext) j.Ast { return ctx.Error("too many characters") }),
								))),
							))),
							i(named("FunctionLit", rules(
								o("'func' Signature FunctionBody"),
								// i(named("Signature", "TODO")),
								// i(named("FunctionBody", "TODO")),
							))),
							i(named("CompositeLit", rules())),
						))),
						i(named("OperandName", rules(
							o("QualifiedIdent | identifier"),
							i(named("QualifiedIdent", "p:PackageName DOT i:identifier"), fQualifiedIdent),
						))),
					))),
					// TODO add to below alternation: Type [ "," ExpressionList ]
					i(named("Arguments", "'(' (Args:(ExpressionList) Varg:THREEDOTS? ','? )? ')'"), fArguments),
					i(named("Index", `'[' Expression ','? ']'`)),
					i(named("Slice", `'[' (Expression?)*':'{2,3} ']'`)),
					i(named("Selector", `'.' identifier`)),
				))),
			)), fExpression),
			i(named("ExpressionList", "Expression+_COMMA")),
		))),
		i(named("PackageClause", "'package' PackageName")),
		i(named("PackageName", "identifier"), fPackageName),
		// i(named("identifier", "[a-zA-Z_] ([a-zA-Z0-9_])*"), func(it j.Ast) j.Ast { return Nx(it.(*j.NativeArray).Concat()) }),
		i(named("identifier", "letter (letter | unicode_digit)*"), fIdentifier),
		i(named("IdentifierList", "identifier+','")),
		i(named("characters", "(newline | unicode_char | unicode_letter | unicode_digit)")),
		i(named("newline", `[\x{0a}]`)),
		i(named("letter", `(?[^0-9 \t\n\r+(){}[\]<>-])`), fLetter), // lookahead next rune, not impossible letter? then try parse with fLetter. gospec = "unicode_letter | '_'"
		i(named("unicode_char", "[^\\x{0a}]")),                     // "an arbitrary Unicode code point except newline"
		i(named("unicode_letter", "[a-zA-Z]")),                     // "a Unicode code point categorized as "Letter" TODO it misses all non ASCII
		i(named("unicode_digit", "[0-9]")),                         // "a Unicode code point categorized as "Number, decimal digit" TODO it misses all non ASCII
		//                         ^^^
		// For now we'll stick to ANSI for letters and digits. It can later be improved
		// We looked into unicode specs for them but there are not defined
		// in https://www.unicode.org/versions/Unicode8.0.0/ch04.pdf Section 4.5

		i(named("_COMMA", "_ ','")),
		i(named("_", "( ' ' | '\t' | '\n' | '\r' )*")),
		i(named("DOT", "'.'")), // when it needs to get captured (by default '.' in a sequence is not captured)
		i(named("THREEDOTS", "'...'")),
		i(named("_semicolon", "';' '\n'?")),
	)
)

// helps writing rules for PEG grammars.
// It splits upon space, reverse order, adds single quotes, and joins upon '|'
// For example:
//
// "* / %"      becomes      "'%'|'/'|'*'".
func revQuote(spaceSeparatedElements string) string {
	a := strings.Fields(spaceSeparatedElements)
	s := ""
	for i := len(a) - 1; i >= 0; i-- {
		s += "'" + a[i] + "'|"
	}
	return s[:len(s)-1]
}

// vim: fdm=indent
