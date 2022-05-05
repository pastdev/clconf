package template_test

import (
	"testing"

	"github.com/pastdev/clconf/v2/pkg/template"
	"github.com/stretchr/testify/assert"
)

func TestEscapeOsgi(t *testing.T) {
	// quotes, double quotes, backslash, the equals sign and spaces need to be escaped
	tests := [][]string{
		{"", ""},
		{"a", "a"},
		{"'", "\\'"},
		{"\"", "\\\""},
		{"\\", "\\\\"},
		{"=", "\\="},
		{" ", "\\ "},
		{
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
		{"host", "domain", "host.domain"},
		{"host.domain", "domain", "host.domain"},
		{"host.domaina", "domain", "host.domaina"},
		{"host", "subdomain.domain", "host.subdomain.domain"},
		{"host.subdomain.domain", "subdomainb.domain", "host.subdomain.domain"},
	}
	for _, test := range tests {
		if actual := template.Fqdn(test[0], test[1]); actual != test[2] {
			t.Errorf("Fqdn failed: [%s] != [%s]", actual, test[2])
		}
	}
}

func TestMarshalJsonString(t *testing.T) {
	tester := func(test string, data interface{}, expected string) {
		t.Run(test, func(t *testing.T) {
			actual, err := template.MarshalJSONString(data)
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}

	tester("simple", "foo", "\"foo\"")
	tester("two lines",
		`foo
bar`,
		"\"foo\\nbar\"")
	tester("with quotes", "foo\"bar", "\"foo\\\"bar\"")
	tester("number", 1, "\"1\"")
	tester("number", false, "\"false\"")
}

func TestRegexReplace(t *testing.T) {
	tests := [][]string{
		{".", "abc", "Z", "ZZZ"},
		{"a(.)c", "abc", "A${1}C", "AbC"},
		{"^a", "abca", "", "bca"},
		{"^abc$", "abc", "def", "def"},
		{"c", "ab\ncd", "C", "ab\nCd"},
	}
	for idx, test := range tests {
		if actual, _ := template.RegexReplace(test[0], test[1], test[2]); actual != test[3] {
			t.Errorf("RegexReplace %d failed: [%s] != [%s]", idx, actual, test[3])
		}
	}
}
