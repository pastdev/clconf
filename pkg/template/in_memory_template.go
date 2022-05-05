package template

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"text/template"

	"github.com/pastdev/clconf/v2/pkg/core"
	"github.com/pastdev/clconf/v2/pkg/memkv"
	"github.com/pastdev/clconf/v2/pkg/secret"
)

// TemplateConfig allows for optional configuration.
type TemplateConfig struct { //nolint:revive
	Prefix      string
	SecretAgent *secret.SecretAgent
	LeftDelim   string
	RightDelim  string
}

// Template is a wrapper for template.Template to include custom template
// functions corresponding to confd functions.
type Template struct {
	config   *TemplateConfig
	store    *memkv.Store
	template *template.Template
}

/////// mapped to confd resource.go ///////
func addCryptFuncs(funcMap map[string]interface{}, sa *secret.SecretAgent) {
	AddFuncs(funcMap, map[string]interface{}{
		"cget": func(key string) (memkv.KVPair, error) {
			kv, err := funcMap["get"].(func(string) (memkv.KVPair, error))(key)
			if err == nil {
				var decrypted string
				decrypted, err = sa.Decrypt(kv.Value)
				if err == nil {
					kv.Value = decrypted
				}
			}
			return kv, fmt.Errorf("cget: %w", err)
		},
		"cgets": func(pattern string) (memkv.KVPairs, error) {
			kvs, err := funcMap["gets"].(func(string) (memkv.KVPairs, error))(pattern)
			if err == nil {
				for i := range kvs {
					decrypted, err := sa.Decrypt(kvs[i].Value)
					if err != nil {
						return memkv.KVPairs(nil), err
					}
					kvs[i].Value = decrypted
				}
			}
			return kvs, fmt.Errorf("cgets: %w", err)
		},
		"cgetv": func(key string) (string, error) {
			v, err := funcMap["getv"].(func(string, ...string) (string, error))(key)
			if err == nil {
				var decrypted string
				decrypted, err = sa.Decrypt(v)
				if err == nil {
					return decrypted, nil
				}
			}
			return v, fmt.Errorf("cgetv: %w", err)
		},
		"cgetvs": func(pattern string) ([]string, error) {
			vs, err := funcMap["getvs"].(func(string) ([]string, error))(pattern)
			if err == nil {
				for i := range vs {
					decrypted, err := sa.Decrypt(vs[i])
					if err != nil {
						return []string(nil), err
					}
					vs[i] = decrypted
				}
			}
			return vs, fmt.Errorf("cgetvs: %w", err)
		},
	})
}

// NewTemplate returns a parsed Template configured with standard functions.
func NewTemplate(name, text string, config *TemplateConfig) (*Template, error) {
	if config == nil {
		config = &TemplateConfig{}
	}

	store := memkv.New()

	funcMap := NewFuncMap(&store)
	AddFuncs(funcMap, store.FuncMap)
	if config.SecretAgent != nil {
		addCryptFuncs(funcMap, config.SecretAgent)
	}

	tmpl, err := template.
		New(name).
		Delims(config.LeftDelim, config.RightDelim).
		Funcs(funcMap).
		Parse(text)
	if err != nil {
		return nil, fmt.Errorf("process template %s: %w", name, err)
	}

	return &Template{
		config:   config,
		store:    &store,
		template: tmpl,
	}, nil
}

// NewTemplateFromBase64 decodes base64 then calls NewTemplate with the result.
func NewTemplateFromBase64(name, template string, config *TemplateConfig) (*Template, error) {
	content, err := base64.StdEncoding.DecodeString(template)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	return NewTemplate(name, string(content), config)
}

// NewTemplateFromFile reads file then calls NewTemplate with the result.
func NewTemplateFromFile(name, file string, config *TemplateConfig) (*Template, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	return NewTemplate(name, string(contents), config)
}

// Execute will process the template text using data and the function map from
// confd.
func (tmpl *Template) Execute(data interface{}) (string, error) {
	tmpl.setVars(data)

	var buf bytes.Buffer
	if err := tmpl.template.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

/////// mapped to confd resource.go ///////
func (tmpl *Template) setVars(data interface{}) {
	value, _ := core.GetValue(data, tmpl.config.Prefix)
	tmpl.store.Purge()
	for k, v := range core.ToKvMap(value) {
		tmpl.store.Set(k, v)
	}
}
