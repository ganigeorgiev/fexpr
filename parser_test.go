package fexpr

import (
	"fmt"
	"testing"
)

func TestExprIzZero(t *testing.T) {
	scenarios := []struct {
		expr   Expr
		result bool
	}{
		{Expr{}, true},
		{Expr{Op: SignAnyEq}, false},
		{Expr{Left: Token{Literal: "123"}}, false},
		{Expr{Left: Token{Type: TokenWS}}, false},
		{Expr{Right: Token{Literal: "123"}}, false},
		{Expr{Right: Token{Type: TokenWS}}, false},
	}

	for i, s := range scenarios {
		t.Run(fmt.Sprintf("s%d", i), func(t *testing.T) {
			if v := s.expr.IsZero(); v != s.result {
				t.Fatalf("Expected %v, got %v for \n%v", s.result, v, s.expr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	scenarios := []struct {
		input         string
		expectedError bool
		expectedPrint string
	}{
		{`> 1`, true, "[]"},
		{`a >`, true, "[]"},
		{`a > >`, true, "[]"},
		{`a > %`, true, "[]"},
		{`a ! 1`, true, "[]"},
		{`a - 1`, true, "[]"},
		{`a + 1`, true, "[]"},
		{`1 - 1`, true, "[]"},
		{`1 + 1`, true, "[]"},
		{`> a 1`, true, "[]"},
		{`a || 1`, true, "[]"},
		{`a && 1`, true, "[]"},
		{`test > 1 &&`, true, `[]`},
		{`|| test = 1`, true, `[]`},
		{`test = 1 && ||`, true, "[]"},
		{`test = 1 && a`, true, "[]"},
		{`test = 1 && a`, true, "[]"},
		{`test = 1 && "a"`, true, "[]"},
		{`test = 1 a`, true, "[]"},
		{`test = 1 a`, true, "[]"},
		{`test = 1 "a"`, true, "[]"},
		{`test = 1@test`, true, "[]"},
		{`test = .@test`, true, "[]"},
		// mismatched text quotes
		{`test = "demo'`, true, "[]"},
		{`test = 'demo"`, true, "[]"},
		{`test = 'demo'"`, true, "[]"},
		{`test = 'demo''`, true, "[]"},
		{`test = "demo"'`, true, "[]"},
		{`test = "demo""`, true, "[]"},
		{`test = ""demo""`, true, "[]"},
		{`test = ''demo''`, true, "[]"},
		{"test = `demo`", true, "[]"},
		// comments
		{"test = / demo", true, "[]"},
		{"test = // demo", true, "[]"},
		{"// demo", true, "[]"},
		{"test = 123 // demo", false, "[{{{<nil> identifier test} = {<nil> number 123}} &&}]"},
		{"test = // demo\n123", false, "[{{{<nil> identifier test} = {<nil> number 123}} &&}]"},
		{`
			a = 123 &&
			// demo
			b = 456
		`, false, "[{{{<nil> identifier a} = {<nil> number 123}} &&} {{{<nil> identifier b} = {<nil> number 456}} &&}]"},
		// functions
		{`test() = 12`, false, `[{{{[] function test} = {<nil> number 12}} &&}]`},
		{`(a.b.c(1) = d.e.f(2)) || 1=2`, false, `[{[{{{[{<nil> number 1}] function a.b.c} = {[{<nil> number 2}] function d.e.f}} &&}] &&} {{{<nil> number 1} = {<nil> number 2}} ||}]`},
		// valid simple expression and sign operators check
		{`1=12`, false, `[{{{<nil> number 1} = {<nil> number 12}} &&}]`},
		{`   1    =    12    `, false, `[{{{<nil> number 1} = {<nil> number 12}} &&}]`},
		{`"demo" != test`, false, `[{{{<nil> text demo} != {<nil> identifier test}} &&}]`},
		{`a~1`, false, `[{{{<nil> identifier a} ~ {<nil> number 1}} &&}]`},
		{`a !~ 1`, false, `[{{{<nil> identifier a} !~ {<nil> number 1}} &&}]`},
		{`test>12`, false, `[{{{<nil> identifier test} > {<nil> number 12}} &&}]`},
		{`test > 12`, false, `[{{{<nil> identifier test} > {<nil> number 12}} &&}]`},
		{`test >="test"`, false, `[{{{<nil> identifier test} >= {<nil> text test}} &&}]`},
		{`test<@demo.test2`, false, `[{{{<nil> identifier test} < {<nil> identifier @demo.test2}} &&}]`},
		{`1<="test"`, false, `[{{{<nil> number 1} <= {<nil> text test}} &&}]`},
		{`1<="te'st"`, false, `[{{{<nil> number 1} <= {<nil> text te'st}} &&}]`},
		{`demo='te\'st'`, false, `[{{{<nil> identifier demo} = {<nil> text te'st}} &&}]`},
		{`demo="te\'st"`, false, `[{{{<nil> identifier demo} = {<nil> text te\'st}} &&}]`},
		{`demo="te\"st"`, false, `[{{{<nil> identifier demo} = {<nil> text te"st}} &&}]`},
		// invalid parenthesis
		{`(a=1`, true, `[]`},
		{`a=1)`, true, `[]`},
		{`((a=1)`, true, `[]`},
		{`{a=1}`, true, `[]`},
		{`[a=1]`, true, `[]`},
		{`((a=1 || a=2) && c=1))`, true, `[]`},
		// valid parenthesis
		{`()`, true, `[]`},
		{`(a=1)`, false, `[{[{{{<nil> identifier a} = {<nil> number 1}} &&}] &&}]`},
		{`(a="test(")`, false, `[{[{{{<nil> identifier a} = {<nil> text test(}} &&}] &&}]`},
		{`(a="test)")`, false, `[{[{{{<nil> identifier a} = {<nil> text test)}} &&}] &&}]`},
		{`((a=1))`, false, `[{[{[{{{<nil> identifier a} = {<nil> number 1}} &&}] &&}] &&}]`},
		{`a=1 || 2!=3`, false, `[{{{<nil> identifier a} = {<nil> number 1}} &&} {{{<nil> number 2} != {<nil> number 3}} ||}]`},
		{`a=1 && 2!=3`, false, `[{{{<nil> identifier a} = {<nil> number 1}} &&} {{{<nil> number 2} != {<nil> number 3}} &&}]`},
		{`a=1 && 2!=3 || "b"=a`, false, `[{{{<nil> identifier a} = {<nil> number 1}} &&} {{{<nil> number 2} != {<nil> number 3}} &&} {{{<nil> text b} = {<nil> identifier a}} ||}]`},
		{`(a=1 && 2!=3) || "b"=a`, false, `[{[{{{<nil> identifier a} = {<nil> number 1}} &&} {{{<nil> number 2} != {<nil> number 3}} &&}] &&} {{{<nil> text b} = {<nil> identifier a}} ||}]`},
		{`((a=1 || a=2) && (c=1))`, false, `[{[{[{{{<nil> identifier a} = {<nil> number 1}} &&} {{{<nil> identifier a} = {<nil> number 2}} ||}] &&} {[{{{<nil> identifier c} = {<nil> number 1}} &&}] &&}] &&}]`},
		// https://github.com/pocketbase/pocketbase/issues/5017
		{`(a='"')`, false, `[{[{{{<nil> identifier a} = {<nil> text "}} &&}] &&}]`},
		{`(a='\'')`, false, `[{[{{{<nil> identifier a} = {<nil> text '}} &&}] &&}]`},
		{`(a="'")`, false, `[{[{{{<nil> identifier a} = {<nil> text '}} &&}] &&}]`},
		{`(a="\"")`, false, `[{[{{{<nil> identifier a} = {<nil> text "}} &&}] &&}]`},
	}

	for i, scenario := range scenarios {
		t.Run(fmt.Sprintf("s%d:%s", i, scenario.input), func(t *testing.T) {
			v, err := Parse(scenario.input)

			if scenario.expectedError && err == nil {
				t.Fatalf("Expected error, got nil (%q)", scenario.input)
			}

			if !scenario.expectedError && err != nil {
				t.Fatalf("Did not expect error, got %q (%q).", err, scenario.input)
			}

			vPrint := fmt.Sprintf("%v", v)

			if vPrint != scenario.expectedPrint {
				t.Fatalf("Expected %s, got %s", scenario.expectedPrint, vPrint)
			}
		})
	}
}
