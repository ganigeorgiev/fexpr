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
		{"   ", []output{{false, "{whitespace    }"}}},
		{"test 123", []output{{false, "{identifier test}"}, {false, "{whitespace  }"}, {false, "{number 123}"}}},
		// identifier
		{`test`, []output{{false, `{identifier test}`}}},
		{`@`, []output{{true, `{identifier @}`}}},
		{`test:`, []output{{true, `{identifier test:}`}}},
		{`test.`, []output{{true, `{identifier test.}`}}},
		{`@test.123`, []output{{false, `{identifier @test.123}`}}},
		{`_test.123`, []output{{false, `{identifier _test.123}`}}},
		{`#test.123:456`, []output{{false, `{identifier #test.123:456}`}}},
		{`.test.123`, []output{{true, `{unexpected .}`}, {false, `{identifier test.123}`}}},
		{`:test.123`, []output{{true, `{unexpected :}`}, {false, `{identifier test.123}`}}},
		{`a.b:c`, []output{{false, `{identifier a.b:c}`}}},
		{`test#@`, []output{{false, `{identifier test}`}, {true, `{identifier #}`}, {true, `{identifier @}`}}},
		{`test'`, []output{{false, `{identifier test}`}, {true, `{text '}`}}},
		{`test"d`, []output{{false, `{identifier test}`}, {true, `{text "d}`}}},
		// number
		{`123`, []output{{false, `{number 123}`}}},
		{`-123`, []output{{false, `{number -123}`}}},
		{`-123.456`, []output{{false, `{number -123.456}`}}},
		{`123.456`, []output{{false, `{number 123.456}`}}},
		{`12.34.56`, []output{{false, `{number 12.34}`}, {true, `{unexpected .}`}, {false, `{number 56}`}}},
		{`.123`, []output{{true, `{unexpected .}`}, {false, `{number 123}`}}},
		{`- 123`, []output{{true, `{number -}`}, {false, `{whitespace  }`}, {false, `{number 123}`}}},
		{`12-3`, []output{{false, `{number 12}`}, {false, `{number -3}`}}},
		{`123.abc`, []output{{true, `{number 123.}`}, {false, `{identifier abc}`}}},
		// text
		{`""`, []output{{false, `{text }`}}},
		{`''`, []output{{false, `{text }`}}},
		{`'test'`, []output{{false, `{text test}`}}},
		{`'te\'st'`, []output{{false, `{text te'st}`}}},
		{`"te\"st"`, []output{{false, `{text te"st}`}}},
		{`"tes@#,;!@#%^'\"t"`, []output{{false, `{text tes@#,;!@#%^'"t}`}}},
		{`'tes@#,;!@#%^\'"t'`, []output{{false, `{text tes@#,;!@#%^'"t}`}}},
		{`"test`, []output{{true, `{text "test}`}}},
		{`'test`, []output{{true, `{text 'test}`}}},
		{`'АБЦ`, []output{{true, `{text 'АБЦ}`}}},
		// join types
		{`&&||`, []output{{true, `{join &&||}`}}},
		{`&& ||`, []output{{false, `{join &&}`}, {false, `{whitespace  }`}, {false, `{join ||}`}}},
		{`'||test&&'&&123`, []output{{false, `{text ||test&&}`}, {false, `{join &&}`}, {false, `{number 123}`}}},
		// expression signs
		{`=!=`, []output{{true, `{sign =!=}`}}},
		{`= != ~ !~ > >= < <= ?= ?!= ?~ ?!~ ?> ?>= ?< ?<=`, []output{
			{false, `{sign =}`},
			{false, `{whitespace  }`},
			{false, `{sign !=}`},
			{false, `{whitespace  }`},
			{false, `{sign ~}`},
			{false, `{whitespace  }`},
			{false, `{sign !~}`},
			{false, `{whitespace  }`},
			{false, `{sign >}`},
			{false, `{whitespace  }`},
			{false, `{sign >=}`},
			{false, `{whitespace  }`},
			{false, `{sign <}`},
			{false, `{whitespace  }`},
			{false, `{sign <=}`},
			{false, `{whitespace  }`},
			{false, `{sign ?=}`},
			{false, `{whitespace  }`},
			{false, `{sign ?!=}`},
			{false, `{whitespace  }`},
			{false, `{sign ?~}`},
			{false, `{whitespace  }`},
			{false, `{sign ?!~}`},
			{false, `{whitespace  }`},
			{false, `{sign ?>}`},
			{false, `{whitespace  }`},
			{false, `{sign ?>=}`},
			{false, `{whitespace  }`},
			{false, `{sign ?<}`},
			{false, `{whitespace  }`},
			{false, `{sign ?<=}`},
		}},
		// groups/parenthesis
		{`a)`, []output{{false, `{identifier a}`}, {true, `{unexpected )}`}}},
		{`(a b c`, []output{{true, `{group a b c}`}}},
		{`(a b c)`, []output{{false, `{group a b c}`}}},
		{`((a b c))`, []output{{false, `{group (a b c)}`}}},
		{`((a )b c))`, []output{{false, `{group (a )b c}`}, {true, `{unexpected )}`}}},
		{`("ab)("c)`, []output{{false, `{group "ab)("c}`}}},
		{`("ab)(c)`, []output{{true, `{group "ab)(c)}`}}},
		// comments
		{`/ test`, []output{{true, `{comment }`}, {false, `{identifier test}`}}},
		{`/ / test`, []output{{true, `{comment }`}, {true, `{comment }`}, {false, `{identifier test}`}}},
		{`//`, []output{{false, `{comment }`}}},
		{`//test`, []output{{false, `{comment test}`}}},
		{`// test`, []output{{false, `{comment test}`}}},
		{`//   test1 //test2  `, []output{{false, `{comment test1 //test2}`}}},
		{`///test`, []output{{false, `{comment /test}`}}},
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
