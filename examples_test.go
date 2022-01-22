package fexpr_test

import (
	"fmt"
	"strings"

	"github.com/ganigeorgiev/fexpr"
)

func ExampleNewScanner() {
	fexpr.NewScanner(strings.NewReader("id"))
}

func ExampleScanner_Scan() {
	s := fexpr.NewScanner(strings.NewReader("id > 123"))

	for {
		t, err := s.Scan()
		if t.Type == fexpr.TokenEOF || err != nil {
			break
		}

		fmt.Println(t)
	}

	// Output:
	// {identifier id}
	// {whitespace  }
	// {sign >}
	// {whitespace  }
	// {number 123}
}

func ExampleParse() {
	result, _ := fexpr.Parse("id > 123")

	fmt.Println(result)

	// Output:
	// [{&& {{identifier id} > {number 123}}}]
}
