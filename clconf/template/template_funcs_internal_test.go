package template

import (
	"fmt"
	"testing"
)

func TestSort(t *testing.T) {
	test := func(name, sortType string, src []string, expected []string) {
		t.Run(name, func(t *testing.T) {
			actual, err := sortAs(src, sortType)
			if err != nil {
				t.Fatalf("Error sorting: %v", err)
			}
			if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
				t.Errorf("Actual != Expected: %v != %v", actual, expected)
			}
		})
	}
	test("string", "string", []string{"b", "A", "def", "a"}, []string{"A", "a", "b", "def"})
	test("int", "int", []string{"10", "-1", "0", "2"}, []string{"-1", "0", "2", "10"})
	test("mixed int", "int",
		[]string{"10", "-1", "0", "2", "foo", "bar", "1foo"},
		[]string{"-1", "0", "2", "10", "1foo", "bar", "foo"})
}
