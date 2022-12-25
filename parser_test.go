package fexpr

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	testScenarios := []struct {
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
		// valid simple expression and sign operators check
		{`1=12`, false, `[{&& {{number 1} = {number 12}}}]`},
		{`   1    =    12    `, false, `[{&& {{number 1} = {number 12}}}]`},
		{`"demo" != test`, false, `[{&& {{text demo} != {identifier test}}}]`},
		{`a~1`, false, `[{&& {{identifier a} ~ {number 1}}}]`},
		{`a !~ 1`, false, `[{&& {{identifier a} !~ {number 1}}}]`},
		{`test>12`, false, `[{&& {{identifier test} > {number 12}}}]`},
		{`test > 12`, false, `[{&& {{identifier test} > {number 12}}}]`},
		{`test >="test"`, false, `[{&& {{identifier test} >= {text test}}}]`},
		{`test<@demo.test2`, false, `[{&& {{identifier test} < {identifier @demo.test2}}}]`},
		{`1<="test"`, false, `[{&& {{number 1} <= {text test}}}]`},
		{`1<="te'st"`, false, `[{&& {{number 1} <= {text te'st}}}]`},
		{`demo='te\'st'`, false, `[{&& {{identifier demo} = {text te'st}}}]`},
		{`demo="te\'st"`, false, `[{&& {{identifier demo} = {text te\'st}}}]`},
		{`demo="te\"st"`, false, `[{&& {{identifier demo} = {text te"st}}}]`},
		// invalid parenthesis
		{`(a=1`, true, `[]`},
		{`a=1)`, true, `[]`},
		{`((a=1)`, true, `[]`},
		{`{a=1}`, true, `[]`},
		{`[a=1]`, true, `[]`},
		{`((a=1 || a=2) && c=1))`, true, `[]`},
		// valid parenthesis
		{`()`, true, `[]`},
		{`(a=1)`, false, `[{&& [{&& {{identifier a} = {number 1}}}]}]`},
		{`(a="test(")`, false, `[{&& [{&& {{identifier a} = {text test(}}}]}]`},
		{`(a="test)")`, false, `[{&& [{&& {{identifier a} = {text test)}}}]}]`},
		{`((a=1))`, false, `[{&& [{&& [{&& {{identifier a} = {number 1}}}]}]}]`},
		{`a=1 || 2!=3`, false, `[{&& {{identifier a} = {number 1}}} {|| {{number 2} != {number 3}}}]`},
		{`a=1 && 2!=3`, false, `[{&& {{identifier a} = {number 1}}} {&& {{number 2} != {number 3}}}]`},
		{`a=1 && 2!=3 || "b"=a`, false, `[{&& {{identifier a} = {number 1}}} {&& {{number 2} != {number 3}}} {|| {{text b} = {identifier a}}}]`},
		{`(a=1 && 2!=3) || "b"=a`, false, `[{&& [{&& {{identifier a} = {number 1}}} {&& {{number 2} != {number 3}}}]} {|| {{text b} = {identifier a}}}]`},
		{`((a=1 || a=2) && (c=1))`, false, `[{&& [{&& [{&& {{identifier a} = {number 1}}} {|| {{identifier a} = {number 2}}}]} {&& [{&& {{identifier c} = {number 1}}}]}]}]`},
	}

	for i, scenario := range testScenarios {
		v, err := Parse(scenario.input)

		if scenario.expectedError && err == nil {
			t.Errorf("(%d) Expected error, got nil (%q)", i, scenario.input)
		}

		if !scenario.expectedError && err != nil {
			t.Errorf("(%d) Did not expect error, got %q (%q).", i, err, scenario.input)
		}

		vPrint := fmt.Sprintf("%v", v)

		if vPrint != scenario.expectedPrint {
			t.Errorf("(%d) Expected %s, got %s", i, scenario.expectedPrint, vPrint)
		}
	}
}
