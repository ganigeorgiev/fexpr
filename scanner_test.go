package fexpr

import (
	"fmt"
	"testing"
)

func TestNewScanner(t *testing.T) {
	s := NewScanner([]byte("test"))

	data := string(s.data)

	if data != "test" {
		t.Errorf("Expected the scanner reader data to be %q, got %q", "test", data)
	}
}

func TestScannerScan(t *testing.T) {
	type output struct {
		error bool
		print string
	}
	testScenarios := []struct {
		text    string
		expects []output
	}{
		// whitespace
		{"   ", []output{{false, "{<nil> whitespace    }"}}},
		{"test 123", []output{{false, "{<nil> identifier test}"}, {false, "{<nil> whitespace  }"}, {false, "{<nil> number 123}"}}},
		// identifier
		{`test`, []output{{false, `{<nil> identifier test}`}}},
		{`@`, []output{{true, `{<nil> identifier @}`}}},
		{`test:`, []output{{true, `{<nil> identifier test:}`}}},
		{`test.`, []output{{true, `{<nil> identifier test.}`}}},
		{`@test.123:c`, []output{{false, `{<nil> identifier @test.123:c}`}}},
		{`_test_a.123`, []output{{false, `{<nil> identifier _test_a.123}`}}},
		{`#test.123:456`, []output{{false, `{<nil> identifier #test.123:456}`}}},
		{`.test.123`, []output{{true, `{<nil> unexpected .}`}, {false, `{<nil> identifier test.123}`}}},
		{`:test.123`, []output{{true, `{<nil> unexpected :}`}, {false, `{<nil> identifier test.123}`}}},
		{`test#@`, []output{{false, `{<nil> identifier test}`}, {true, `{<nil> identifier #}`}, {true, `{<nil> identifier @}`}}},
		{`test'`, []output{{false, `{<nil> identifier test}`}, {true, `{<nil> text '}`}}},
		{`test"d`, []output{{false, `{<nil> identifier test}`}, {true, `{<nil> text "d}`}}},
		// number
		{`123`, []output{{false, `{<nil> number 123}`}}},
		{`-123`, []output{{false, `{<nil> number -123}`}}},
		{`-123.456`, []output{{false, `{<nil> number -123.456}`}}},
		{`123.456`, []output{{false, `{<nil> number 123.456}`}}},
		{`12.34.56`, []output{{false, `{<nil> number 12.34}`}, {true, `{<nil> unexpected .}`}, {false, `{<nil> number 56}`}}},
		{`.123`, []output{{true, `{<nil> unexpected .}`}, {false, `{<nil> number 123}`}}},
		{`- 123`, []output{{true, `{<nil> number -}`}, {false, `{<nil> whitespace  }`}, {false, `{<nil> number 123}`}}},
		{`12-3`, []output{{false, `{<nil> number 12}`}, {false, `{<nil> number -3}`}}},
		{`123.abc`, []output{{true, `{<nil> number 123.}`}, {false, `{<nil> identifier abc}`}}},
		// text
		{`""`, []output{{false, `{<nil> text }`}}},
		{`''`, []output{{false, `{<nil> text }`}}},
		{`'test'`, []output{{false, `{<nil> text test}`}}},
		{`'te\'st'`, []output{{false, `{<nil> text te'st}`}}},
		{`"te\"st"`, []output{{false, `{<nil> text te"st}`}}},
		{`"tes@#,;!@#%^'\"t"`, []output{{false, `{<nil> text tes@#,;!@#%^'"t}`}}},
		{`'tes@#,;!@#%^\'"t'`, []output{{false, `{<nil> text tes@#,;!@#%^'"t}`}}},
		{`"test`, []output{{true, `{<nil> text "test}`}}},
		{`'test`, []output{{true, `{<nil> text 'test}`}}},
		{`'АБЦ`, []output{{true, `{<nil> text 'АБЦ}`}}},
		// join types
		{`&&||`, []output{{true, `{<nil> join &&||}`}}},
		{`&& ||`, []output{{false, `{<nil> join &&}`}, {false, `{<nil> whitespace  }`}, {false, `{<nil> join ||}`}}},
		{`'||test&&'&&123`, []output{{false, `{<nil> text ||test&&}`}, {false, `{<nil> join &&}`}, {false, `{<nil> number 123}`}}},
		// expression signs
		{`=!=`, []output{{true, `{<nil> sign =!=}`}}},
		{`= != ~ !~ > >= < <= ?= ?!= ?~ ?!~ ?> ?>= ?< ?<=`, []output{
			{false, `{<nil> sign =}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign !=}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ~}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign !~}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign >}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign >=}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign <}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign <=}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?=}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?!=}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?~}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?!~}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?>}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?>=}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?<}`},
			{false, `{<nil> whitespace  }`},
			{false, `{<nil> sign ?<=}`},
		}},
		// comments
		{`/ test`, []output{{true, `{<nil> comment }`}, {false, `{<nil> identifier test}`}}},
		{`/ / test`, []output{{true, `{<nil> comment }`}, {true, `{<nil> comment }`}, {false, `{<nil> identifier test}`}}},
		{`//`, []output{{false, `{<nil> comment }`}}},
		{`//test`, []output{{false, `{<nil> comment test}`}}},
		{`// test`, []output{{false, `{<nil> comment test}`}}},
		{`//   test1 //test2  `, []output{{false, `{<nil> comment test1 //test2}`}}},
		{`///test`, []output{{false, `{<nil> comment /test}`}}},
		// funcs
		{`test()`, []output{{false, `{[] function test}`}}},
		{`test(a, b`, []output{{true, `{[{<nil> identifier a} {<nil> identifier b}] function test}`}}},
		{`@test:abc()`, []output{{false, `{[] function @test:abc}`}}},
		{`test(  a  )`, []output{{false, `{[{<nil> identifier a}] function test}`}}}, // with whitespaces
		{`test(a, b)`, []output{{false, `{[{<nil> identifier a} {<nil> identifier b}] function test}`}}},
		{`test(a, b,  )`, []output{{false, `{[{<nil> identifier a} {<nil> identifier b}] function test}`}}},                                                                          // single trailing comma
		{`test(a,,)`, []output{{true, `{[{<nil> identifier a}] function test}`}, {true, `{<nil> unexpected )}`}}},                                                                    // unexpected trailing commas
		{`test(a,,,b)`, []output{{true, `{[{<nil> identifier a}] function test}`}, {true, `{<nil> unexpected ,}`}, {false, `{<nil> identifier b}`}, {true, `{<nil> unexpected )}`}}}, // unexpected mid-args commas
		{`test(   @test.a.b:test  , 123, "ab)c", 'd,ce', false)`, []output{{false, `{[{<nil> identifier @test.a.b:test} {<nil> number 123} {<nil> text ab)c} {<nil> text d,ce} {<nil> identifier false}] function test}`}}},
		{"test(a //test)", []output{{true, `{[{<nil> identifier a}] function test}`}}},    // invalid simple comment
		{"test(a //test\n)", []output{{false, `{[{<nil> identifier a}] function test}`}}}, // valid simple comment
		{"test(a, //test\n, b)", []output{{true, `{[{<nil> identifier a}] function test}`}, {false, `{<nil> whitespace  }`}, {false, `{<nil> identifier b}`}, {true, `{<nil> unexpected )}`}}},
		{"test(a, //test\n b)", []output{{false, `{[{<nil> identifier a} {<nil> identifier b}] function test}`}}},
		{"test(a, test(test(b), c), d)", []output{{false, `{[{<nil> identifier a} {[{[{<nil> identifier b}] function test} {<nil> identifier c}] function test} {<nil> identifier d}] function test}`}}},
		// max funcs depth
		{"a(b(c(1)))", []output{{false, `{[{[{[{<nil> number 1}] function c}] function b}] function a}`}}},
		{"a(b(c(d(1))))", []output{{true, `{[] function a}`}, {false, `{<nil> number 1}`}, {true, `{<nil> unexpected )}`}, {true, `{<nil> unexpected )}`}, {true, `{<nil> unexpected )}`}, {true, `{<nil> unexpected )}`}}},
		// groups/parenthesis
		{`a)`, []output{{false, `{<nil> identifier a}`}, {true, `{<nil> unexpected )}`}}},
		{`(a b c`, []output{{true, `{<nil> group a b c}`}}},
		{`(a b c)`, []output{{false, `{<nil> group a b c}`}}},
		{`((a b c))`, []output{{false, `{<nil> group (a b c)}`}}},
		{`((a )b c))`, []output{{false, `{<nil> group (a )b c}`}, {true, `{<nil> unexpected )}`}}},
		{`("ab)("c)`, []output{{false, `{<nil> group "ab)("c}`}}},
		{`("ab)(c)`, []output{{true, `{<nil> group "ab)(c)}`}}},
		{`( func(1, 2, 3, func(4)) a b c )`, []output{{false, `{<nil> group  func(1, 2, 3, func(4)) a b c }`}}},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.text, func(t *testing.T) {
			s := NewScanner([]byte(scenario.text))

			// scan the text tokens
			for j, expect := range scenario.expects {
				token, err := s.Scan()

				hasErr := err != nil
				if expect.error != hasErr {
					t.Errorf("[%d] Expected hasErr %v, got %v: %v (%v)", j, expect.error, hasErr, err, token)
				}

				tokenPrint := fmt.Sprintf("%v", token)
				if tokenPrint != expect.print {
					t.Errorf("[%d] Expected token %s, got %s", j, expect.print, tokenPrint)
				}
			}

			// the last remaining token should be the eof
			lastToken, err := s.Scan()
			if err != nil || lastToken.Type != TokenEOF {
				t.Fatalf("Expected EOF token, got %v (%v)", lastToken, err)
			}
		})
	}
}
