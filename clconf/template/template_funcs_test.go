package template_test

import (
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
