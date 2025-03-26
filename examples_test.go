package fexpr_test

import (
	"fmt"

	"github.com/ganigeorgiev/fexpr"
)

func ExampleScanner_Scan() {
	s := fexpr.NewScanner([]byte("id > 123"))

	for {
		t, err := s.Scan()
		if t.Type == fexpr.TokenEOF || err != nil {
			break
		}

		fmt.Println(t)
	}

	// Output:
	// {<nil> identifier id}
	// {<nil> whitespace  }
	// {<nil> sign >}
	// {<nil> whitespace  }
	// {<nil> number 123}
}

func ExampleParse() {
	result, _ := fexpr.Parse("id > 123")

	fmt.Println(result)

	// Output:
	// [{{{<nil> identifier id} > {<nil> number 123}} &&}]
}
