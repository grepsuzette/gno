package gnolang

import (
	"fmt"
	"strings"

	j "github.com/grepsuzette/joeson"
)

func grammar() *j.Grammar {
	if gm == nil {
		gm = j.GrammarFromLines(
			"GNO-grammar",
			gnoRules,
		)
	}
	return gm
}

func rules(a ...j.Line) []j.Line { return a }
func i(a ...any) j.ILine { return j.I(a...) }
func o(a ...any) j.OLine { return j.O(a...) }
func named(name string, thing interface{}) j.NamedRule { return j.Named(name, thing) }

// Rewrite of X() with Joeson
func Xnew(x interface{}, args ...interface{}) Expr {
	switch cx := x.(type) {
	case Expr:
		return cx
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return X(fmt.Sprintf("%v", x))
	case string:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
	case Name:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
		x = string(cx)
	default:
		panic(fmt.Sprintf("unexpected input type for X(): %T", x))
	}
	sexpr := x.(string)
	sexpr = fmt.Sprintf(sexpr, args...)
	sexpr = strings.TrimSpace(sexpr)
	ast := parseX(sexpr)
	if expr, ok := ast.(Expr); ok {
		return expr
	} else {
		panic(fmt.Sprintf("%s produced by X(%T) does not implement Expr", ast.String(), x))
	}
}

// Producing joeson.Ast, joeson.ParseError or gnolang.Node
// When a ParseError panic happens, it just returns that ParseError,
// allowing it to short-circuit the grammar.
func parseX(s string) (result j.Ast) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(j.ParseError); ok {
				result = pe
			} else {
				panic(e)
			}
		}
	}()
	if tokens, e := j.TokenStreamFromGoCode(s); e != nil {
		result = j.NewParseError(nil, e.Error())
	} else {
		result = grammar().ParseTokens(tokens)
	}
	return
}
