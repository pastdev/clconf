package template_test

import (
	"fmt"
	"testing"

	"github.com/pastdev/clconf/v2/clconf/template"
)

func TestEscapeOsgi(t *testing.T) {
	// quotes, double quotes, backslash, the equals sign and spaces need to be escaped
	tests := [][]string{
		[]string{"", ""},
		[]string{"a", "a"},
		[]string{"'", "\\'"},
		[]string{"\"", "\\\""},
		[]string{"\\", "\\\\"},
		[]string{"=", "\\="},
		[]string{" ", "\\ "},
		[]string{
			"a long 'sentence' using \"some\" of the = characters",
			"a\\ long\\ \\'sentence\\'\\ using\\ \\\"some\\\"\\ of\\ the\\ \\=\\ characters",
		},
	}
	for _, test := range tests {
		if actual := template.EscapeOsgi(test[0]); actual != test[1] {
			t.Errorf("EscapeOsgi failed: [%s] != [%s]", actual, test[1])
		}
	}
}

func TestFqdn(t *testing.T) {
	tests := [][]string{
		[]string{"host", "domain", "host.domain"},
		[]string{"host.domain", "domain", "host.domain"},
		[]string{"host.domaina", "domain", "host.domaina"},
		[]string{"host", "subdomain.domain", "host.subdomain.domain"},
		[]string{"host.subdomain.domain", "subdomainb.domain", "host.subdomain.domain"},
	}
	for _, test := range tests {
		if actual := template.Fqdn(test[0], test[1]); actual != test[2] {
			t.Errorf("Fqdn failed: [%s] != [%s]", actual, test[2])
		}
	}
}

func TestRegexReplace(t *testing.T) {
	tests := [][]string{
		[]string{".", "abc", "Z", "ZZZ"},
		[]string{"a(.)c", "abc", "A${1}C", "AbC"},
		[]string{"^a", "abca", "", "bca"},
		[]string{"^abc$", "abc", "def", "def"},
		[]string{"c", "ab\ncd", "C", "ab\nCd"},
	}
	for idx, test := range tests {
		if actual, _ := template.RegexReplace(test[0], test[1], test[2]); actual != test[3] {
			t.Errorf("RegexReplace %d failed: [%s] != [%s]", idx, actual, test[3])
		}
	}
}

func TestSort(t *testing.T) {
	test := func(name, sortAs string, src []string, expected []string) {
		t.Run(name, func(t *testing.T) {
			actual, err := template.Sort(src, sortAs)
			if err != nil {
				t.Fatalf("Error sorting: %v", err)
			}
			if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
				t.Errorf("Actual != Expected: %v != %v", actual, expected)
			}
		});
	}
	test("string", "string", []string{"b", "A", "def", "a"}, []string{"A", "a", "b", "def"})
	test("int", "int", []string{"10", "-1", "0", "2"}, []string{"-1", "0", "2", "10"})
}
